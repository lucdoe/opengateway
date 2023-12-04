package internal

import (
	"context"

	"github.com/redis/go-redis/v9"
)

// https://redis.io/docs/connect/clients/go/
var RDB *redis.Client

func InitializeRedis() error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := RDB.Ping(context.Background()).Result()
	if err != nil {
		return err
	}

	return nil
}
