package db

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client{
	addr:=os.Getenv("REDIS_ADDR")
	client:=redis.NewClient(&redis.Options{
		Addr:addr,
	})
	return client
}