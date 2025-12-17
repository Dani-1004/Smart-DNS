package db

import (
	"context"

	"github.com/redis/go-redis/v9"
)

var (
	Ctx = context.Background()
	Rdb *redis.Client
)

func InitRedis() {
	Rdb = redis.NewClient(&redis.Options{
		Addr:     "192.168.50.15:6379",
		Password: "x",
		DB:       0,
	})
}
