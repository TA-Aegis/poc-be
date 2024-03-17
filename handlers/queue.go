package handlers

import (
	"aegis/poc/models"
	"aegis/poc/utils"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

const queueKey = "queue"

func getQueue(ctx context.Context, redisClient *redis.Client) ([]*models.Client, error) {
	queueData, err := redisClient.Get(ctx, queueKey).Result()
	if err == redis.Nil {
		return []*models.Client{}, nil
	} else if err != nil {
		fmt.Println("Error getting queue from redis", err.Error())
		return nil, err
	}
	var queue []*models.Client
	err = json.Unmarshal([]byte(queueData), &queue)
	if err != nil {
		return nil, err
	}
	return queue, nil
}

func updateQueue(ctx context.Context, redisClient *redis.Client, queue []*models.Client) error {
	queueData, err := json.Marshal(queue)
	if err != nil {
		return err
	}
	return redisClient.Set(ctx, queueKey, string(queueData), 0).Err()
}

func calculateWaitTime(position int) time.Duration {
	return 15 * time.Second * time.Duration(position+1)
}

func processQueue(ctx context.Context, redisClient *redis.Client, client *models.Client, queue []*models.Client) error {
	idx, found := utils.FindClientIndex(client.ID, queue)
	if !found {
		return errors.New("No client found")
	}

	var message string
	var isFinished bool
	var percetageProgress float64

	waitTime := 15 * time.Second * time.Duration(idx+1)
	elapsed := time.Since(client.RegisterAt)
	isWaitDone := elapsed >= waitTime && idx == 0

	if isWaitDone {
		message = "Your turn has arrived."
		isFinished = true
		queue = append(queue[:idx], queue[idx+1:]...)
		percetageProgress = 100
		waitTime = 0
	} else {
		waitTime = 15*time.Second*time.Duration(idx+1) - elapsed
		message = fmt.Sprintf("Please wait, estimated time: %.0f seconds.", math.Floor(waitTime.Seconds()))
		isFinished = false
		percetageProgress = float64(idx+1) / float64(len(queue)+1) * 100
	}

	if waitTime < 0 {
		waitTime = 0
	}

	client.Events <- &models.QueueEvents{
		UserID:            client.ID,
		QueueNumber:       idx + 1,
		EstimatedTime:     math.Floor(waitTime.Seconds()),
		Message:           message,
		PercetageProgress: percetageProgress,
		IsFinished:        isFinished,
	}

	if isFinished {
		if err := updateQueue(ctx, redisClient, queue); err != nil {
			return fmt.Errorf("failed to update queue in Redis: %v", err)
		}
	}

	return nil
}

func (hd *HandlerDependencies) QueueHandler(w http.ResponseWriter, r *http.Request) {
	var queue []*models.Client
	ctx := r.Context()
	rc := hd.RedisClient

	queue, err := getQueue(ctx, rc)
	if err != nil {
		http.Error(w, "Failed to deserialize queue", http.StatusInternalServerError)
		return
	}

	client := &models.Client{ID: r.RemoteAddr, RegisterAt: time.Now()}
	client.Events = make(chan *models.QueueEvents, 10)
	queue = append(queue, client)

	err = updateQueue(ctx, rc, queue)
	if err != nil {
		fmt.Println("Error updating queue from redis", err.Error())
		http.Error(w, "Failed to store updated queue in Redis", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher := w.(http.Flusher)

	for {
		currentqueue, err := getQueue(ctx, rc)
		if err != nil {
			http.Error(w, "Failed to deserialize queue", http.StatusInternalServerError)
			close(client.Events)
			return
		}
		err = processQueue(ctx, rc, client, currentqueue)
		if err != nil {
			fmt.Println("Error processing queue", err.Error())
			http.Error(w, "Failed to process queue", http.StatusInternalServerError)
			close(client.Events)
			return
		}
		select {
		case event := <-client.Events:
			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			enc.Encode(event)
			fmt.Printf("data: %v\n", buf.String())
			fmt.Fprintf(w, "data: %v\n", buf.String())
			flusher.Flush()
		}
		time.Sleep(1 * time.Second)
	}
}

func (hd *HandlerDependencies) ResetQueueHandler(w http.ResponseWriter, r *http.Request) {
	err := hd.RedisClient.FlushAll(r.Context()).Err()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to reset the queue: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Success reset the queue"))
}
