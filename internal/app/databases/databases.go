package databases

import (
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitializeRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}
