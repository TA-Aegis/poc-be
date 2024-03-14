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
	return client
}
