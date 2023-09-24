package cache

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

func Start() (*redis.Client, context.Context) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",
		DB:       0,
	})

	ctx := context.Background()
	return client, ctx
}
