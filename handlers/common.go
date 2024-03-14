package handlers

import (
	"github.com/redis/go-redis/v9"
)

type HandlerDependencies struct {
	RedisClient *redis.Client
}

func NewHandlerDependencies(redisClient *redis.Client) *HandlerDependencies {
	return &HandlerDependencies{
		RedisClient: redisClient,
	}
}
