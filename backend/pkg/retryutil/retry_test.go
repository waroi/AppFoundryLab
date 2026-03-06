package retryutil

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDoRetriesUntilSuccess(t *testing.T) {
	attempts := 0

	value, err := Do(context.Background(), 3, 0, func(context.Context) (string, error) {
		attempts++
		if attempts < 3 {
			return "", errors.New("temporary failure")
		}
		return "ok", nil
	})
	if err != nil {
		t.Fatalf("expected retry loop to succeed, got %v", err)
	}
	if value != "ok" {
		t.Fatalf("expected value ok, got %q", value)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestDoHonorsContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Do(ctx, 3, 50*time.Millisecond, func(context.Context) (string, error) {
		return "", errors.New("temporary failure")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context cancellation, got %v", err)
	}
}
