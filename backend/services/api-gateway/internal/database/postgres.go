package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"github.com/example/appfoundrylab/backend/pkg/retryutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pgMu   sync.Mutex
	pgPool *pgxpool.Pool
)

func PostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	pgMu.Lock()
	defer pgMu.Unlock()

	if pgPool != nil {
		return pgPool, nil
	}

	dsn := PostgresDSN()
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = int32(env.GetIntWithDefault("PG_POOL_MAX_CONNS", 20))
	cfg.MinConns = int32(env.GetIntWithDefault("PG_POOL_MIN_CONNS", 2))
	cfg.MaxConnLifetime = time.Duration(env.GetIntWithDefault("PG_POOL_MAX_CONN_LIFETIME_MIN", 30)) * time.Minute
	cfg.MaxConnIdleTime = time.Duration(env.GetIntWithDefault("PG_POOL_MAX_CONN_IDLE_TIME_MIN", 5)) * time.Minute

	pool, err := retryutil.Do(ctx, dependencyConnectAttempts(), dependencyConnectBackoff(), func(attemptCtx context.Context) (*pgxpool.Pool, error) {
		candidate, err := pgxpool.NewWithConfig(attemptCtx, cfg)
		if err != nil {
			return nil, err
		}

		pingCtx, cancel := context.WithTimeout(attemptCtx, dependencyPingTimeout())
		defer cancel()
		if err := candidate.Ping(pingCtx); err != nil {
			candidate.Close()
			return nil, err
		}

		return candidate, nil
	})
	if err != nil {
		return nil, err
	}

	pgPool = pool
	return pgPool, nil
}

func ResetPostgresPool() {
	pgMu.Lock()
	defer pgMu.Unlock()

	if pgPool != nil {
		pgPool.Close()
		pgPool = nil
	}
}

func dependencyConnectAttempts() int {
	attempts := env.GetIntWithDefault("DEPENDENCY_CONNECT_MAX_ATTEMPTS", 4)
	if attempts < 1 {
		return 1
	}
	return attempts
}

func dependencyConnectBackoff() time.Duration {
	backoffMS := env.GetIntWithDefault("DEPENDENCY_CONNECT_BACKOFF_MS", 250)
	if backoffMS < 0 {
		backoffMS = 0
	}
	return time.Duration(backoffMS) * time.Millisecond
}

func dependencyPingTimeout() time.Duration {
	timeoutMS := env.GetIntWithDefault("DEPENDENCY_PING_TIMEOUT_MS", 1500)
	if timeoutMS <= 0 {
		timeoutMS = 1500
	}
	return time.Duration(timeoutMS) * time.Millisecond
}

func PostgresDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		env.MustGet("POSTGRES_USER"),
		env.MustGet("POSTGRES_PASSWORD"),
		env.MustGet("POSTGRES_HOST"),
		env.GetWithDefault("POSTGRES_PORT", "5432"),
		env.MustGet("POSTGRES_DB"),
	)
}
