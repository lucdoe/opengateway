package internal

import (
	"github.com/redis/go-redis/v9"
)

// https://redis.io/docs/connect/clients/go/
var RDB *redis.Client

func InitializeRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}
