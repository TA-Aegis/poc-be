package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
)

type Client struct {
	id         string
	registerAt time.Time
	events     chan *QueueEvents
}
type DashBoard struct {
	User uint
}

type Queue struct {
	QueueNumber   int    `json:"queueNumber"`
	EstimatedTime string `json:"estimatedTime"`
	Message       string `json:"message,omitempty"`
}

type QueueEvents struct {
	UserID        string  `json:"userId"`
	QueueNumber   int     `json:"queueNumber"`
	EstimatedTime float64 `json:"estimatedTime"`
	Message       string  `json:"message,omitempty"`
}

var (
	queue           = make([]*Client, 0)
	nextQueueNumber = 1
	mutex           sync.Mutex
)

func main() {
	app := fiber.New()

	// Serve static files
	app.Static("/", "./public")

	app.Get("/events", adaptor.HTTPHandler(handler(queueHandler)))
	app.Listen(":8080")
}

func handler(f http.HandlerFunc) http.Handler {
	return http.HandlerFunc(f)
}

func queueHandler(w http.ResponseWriter, r *http.Request) {
	client := &Client{id: r.RemoteAddr, registerAt: time.Now(), events: make(chan *QueueEvents, 10)}
	queue = append(queue, client)
	go processQueue(client)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for event := range client.events {
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.Encode(event)
		fmt.Printf("data: %v\n", buf.String())
	}

	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	} else {
		fmt.Println("Failed to flush")
	}

}

func processQueue(client *Client) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mutex.Lock()
			idx, found := findClientIndex(client.id)
			if !found {
				mutex.Unlock()
				return
			}

			waitTime := 30 * time.Second * time.Duration(idx+1)
			elapsed := time.Since(client.registerAt)
			isWaitDone := elapsed >= waitTime && idx == 0

			var msg string
			if isWaitDone {
				msg = "Your turn has arrived."
				waitTime = 30 * time.Second * time.Duration(0)
				client.events <- &QueueEvents{
					UserID:        client.id,
					QueueNumber:   idx + 1,
					EstimatedTime: math.Floor(waitTime.Seconds()),
					Message:       msg,
				}

				close(client.events)
				queue = append(queue[:idx], queue[idx+1:]...)
				mutex.Unlock()
				return
			} else {
				waitTime = 30*time.Second*time.Duration(idx+1) - elapsed
				client.events <- &QueueEvents{
					UserID:        client.id,
					QueueNumber:   idx + 1,
					EstimatedTime: math.Floor(waitTime.Seconds()),
					Message:       "",
				}
			}
			mutex.Unlock()

		}
	}
}

func findClientIndex(id string) (int, bool) {
	for i, c := range queue {
		if c.id == id {
			return i, true
		}
	}
	return -1, false
}
