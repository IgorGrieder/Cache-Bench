package database

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func SetupRedis() *redis.Client {

	redis := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", "redis", 6379),
		Password: "",
		DB:       0,
		Protocol: 2,
	})

	err := redis.Ping(context.Background()).Err()
	if err != nil {
		fmt.Printf("Ending the execution %v", err)
		os.Exit(1)
	}

	return redis
}
