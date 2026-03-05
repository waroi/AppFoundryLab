package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pgOnce sync.Once
	pgPool *pgxpool.Pool
	pgErr  error
)

func PostgresPool(ctx context.Context) (*pgxpool.Pool, error) {
	pgOnce.Do(func() {
		dsn := PostgresDSN()
		cfg, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			pgErr = err
			return
		}
		cfg.MaxConns = 20
		cfg.MinConns = 2
		cfg.MaxConnLifetime = 30 * time.Minute
		cfg.MaxConnIdleTime = 5 * time.Minute

		pgPool, pgErr = pgxpool.NewWithConfig(ctx, cfg)
	})
	return pgPool, pgErr
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
