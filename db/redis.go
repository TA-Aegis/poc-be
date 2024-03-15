package db

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func SetupRedis() *redis.Client {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		fmt.Println("REDIS_URL environment variable is not set")
		return nil
	}
	fmt.Println(redisURL)
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		fmt.Println("Error parsing REDIS_URL:", err)
		return nil
	}
	client := redis.NewClient(opt)
	if err := client.Ping(ctx); err != nil {
		fmt.Println("Error ping to redis:", err.Err())
	}
	return client
}
