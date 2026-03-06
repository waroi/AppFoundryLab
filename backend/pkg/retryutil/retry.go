package retryutil

import (
	"context"
	"time"
)

func Do[T any](ctx context.Context, attempts int, backoff time.Duration, fn func(context.Context) (T, error)) (T, error) {
	var zero T
	if attempts < 1 {
		attempts = 1
	}
	if backoff < 0 {
		backoff = 0
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		value, err := fn(ctx)
		if err == nil {
			return value, nil
		}
		lastErr = err

		if attempt == attempts || backoff == 0 {
			continue
		}

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return zero, ctx.Err()
		case <-timer.C:
		}
	}

	return zero, lastErr
}
