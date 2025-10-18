package database

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

var redisClient *redis.Client

func InitRedis(url string) error {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return err
	}

	redisClient = redis.NewClient(opt)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return err
	}

	logrus.Info("Connected to Redis successfully")
	return nil
}

func GetRedisClient() *redis.Client {
	return redisClient
}

func CloseRedis() {
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logrus.Error("Error closing Redis connection: ", err)
		} else {
			logrus.Info("Disconnected from Redis")
		}
	}
}
