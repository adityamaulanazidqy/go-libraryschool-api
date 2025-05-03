package config

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"os"
)

var (
	ctx       = context.Background()
	logLogrus = logrus.New()
)

func ConnRedis() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6380",
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		logLogrus.WithError(err).Fatal("Could not connect to redis")
		return nil
	}

	return rdb
}
