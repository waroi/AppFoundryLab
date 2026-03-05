package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"github.com/redis/go-redis/v9"
)

var (
	redisOnce sync.Once
	redisDB   *redis.Client
	redisErr  error
)

func RedisClient(ctx context.Context) (*redis.Client, error) {
	redisOnce.Do(func() {
		addr := fmt.Sprintf("%s:%s", env.MustGet("REDIS_HOST"), env.GetWithDefault("REDIS_PORT", "6379"))
		redisDB = redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     env.MustGet("REDIS_PASSWORD"),
			DB:           0,
			PoolSize:     30,
			MinIdleConns: 5,
			DialTimeout:  3 * time.Second,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
		})
	})
	return redisDB, redisErr
}
