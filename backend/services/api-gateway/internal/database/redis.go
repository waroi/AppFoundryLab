package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"github.com/example/appfoundrylab/backend/pkg/retryutil"
	"github.com/redis/go-redis/v9"
)

var (
	redisMu sync.Mutex
	redisDB *redis.Client
)

func RedisClient(ctx context.Context) (*redis.Client, error) {
	redisMu.Lock()
	defer redisMu.Unlock()

	if redisDB != nil {
		return redisDB, nil
	}

	client, err := retryutil.Do(ctx, dependencyConnectAttempts(), dependencyConnectBackoff(), func(attemptCtx context.Context) (*redis.Client, error) {
		addr := fmt.Sprintf("%s:%s", env.MustGet("REDIS_HOST"), env.GetWithDefault("REDIS_PORT", "6379"))
		candidate := redis.NewClient(&redis.Options{
			Addr:         addr,
			Password:     env.MustGet("REDIS_PASSWORD"),
			DB:           0,
			PoolSize:     30,
			MinIdleConns: 5,
			DialTimeout:  3 * time.Second,
			ReadTimeout:  2 * time.Second,
			WriteTimeout: 2 * time.Second,
		})

		pingCtx, cancel := context.WithTimeout(attemptCtx, dependencyPingTimeout())
		defer cancel()
		if err := candidate.Ping(pingCtx).Err(); err != nil {
			_ = candidate.Close()
			return nil, err
		}

		return candidate, nil
	})
	if err != nil {
		return nil, err
	}

	redisDB = client
	return redisDB, nil
}

func ResetRedisClient() {
	redisMu.Lock()
	defer redisMu.Unlock()

	if redisDB != nil {
		_ = redisDB.Close()
		redisDB = nil
	}
}
