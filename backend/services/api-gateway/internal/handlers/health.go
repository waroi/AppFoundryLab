package handlers

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/database"
	"github.com/example/appfoundrylab/backend/services/api-gateway/internal/worker"
	"github.com/example/appfoundrylab/backend/services/api-gateway/pkg/httpx"
)

func LiveHandler(w http.ResponseWriter, _ *http.Request) {
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "api-gateway",
	})
}

type readySnapshot struct {
	httpStatus int
	payload    map[string]any
	expiresAt  time.Time
	createdAt  time.Time
}

type readyCache struct {
	mu   sync.Mutex
	snap readySnapshot
}

func newReadyCache() *readyCache {
	return &readyCache{}
}

func (c *readyCache) GetFresh(now time.Time) (readySnapshot, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.snap.payload == nil || !now.Before(c.snap.expiresAt) {
		return readySnapshot{}, false
	}
	return c.snap, true
}

func (c *readyCache) GetStale(now time.Time, staleWindow time.Duration) (readySnapshot, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if staleWindow <= 0 || c.snap.payload == nil {
		return readySnapshot{}, false
	}
	if now.Sub(c.snap.expiresAt) > staleWindow {
		return readySnapshot{}, false
	}
	return c.snap, true
}

func (c *readyCache) Set(snap readySnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snap = snap
}

func (c *readyCache) Invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snap = readySnapshot{}
}

type readinessProbeFunc func(ctx context.Context, workerClient *worker.Client) (int, map[string]any, bool)

type ReadyEndpoints struct {
	workerClient   *worker.Client
	cache          *readyCache
	cacheTTL       time.Duration
	staleIfError   time.Duration
	now            func() time.Time
	readinessProbe readinessProbeFunc
}

func NewReadyEndpoints(workerClient *worker.Client) *ReadyEndpoints {
	cacheTTL := time.Duration(env.GetIntWithDefault("HEALTH_READY_CACHE_TTL_MS", 1000)) * time.Millisecond
	if cacheTTL < 0 {
		cacheTTL = 0
	}
	staleIfError := time.Duration(env.GetIntWithDefault("HEALTH_READY_STALE_IF_ERROR_MS", 10000)) * time.Millisecond
	if staleIfError < 0 {
		staleIfError = 0
	}
	return &ReadyEndpoints{
		workerClient:   workerClient,
		cache:          newReadyCache(),
		cacheTTL:       cacheTTL,
		staleIfError:   staleIfError,
		now:            time.Now,
		readinessProbe: evaluateReadiness,
	}
}

func ReadyHandler(workerClient *worker.Client) http.HandlerFunc {
	return NewReadyEndpoints(workerClient).Ready
}

func (h *ReadyEndpoints) Ready(w http.ResponseWriter, r *http.Request) {
	now := h.now()
	refresh := r.URL.Query().Get("refresh") == "1"
	if !refresh {
		if fresh, ok := h.cache.GetFresh(now); ok {
			w.Header().Set("X-Ready-Cache", "fresh")
			httpx.WriteJSON(w, fresh.httpStatus, fresh.payload)
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	httpStatus, payload, hasError := h.readinessProbe(ctx, h.workerClient)
	if hasError {
		if stale, ok := h.cache.GetStale(now, h.staleIfError); ok {
			w.Header().Set("X-Ready-Cache", "stale")
			httpx.WriteJSON(w, stale.httpStatus, stale.payload)
			return
		}
	}

	h.cache.Set(readySnapshot{
		httpStatus: httpStatus,
		payload:    payload,
		expiresAt:  now.Add(h.cacheTTL),
		createdAt:  now,
	})
	w.Header().Set("X-Ready-Cache", "miss")
	httpx.WriteJSON(w, httpStatus, payload)
}

func (h *ReadyEndpoints) Invalidate(w http.ResponseWriter, _ *http.Request) {
	h.cache.Invalidate()
	httpx.WriteJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"cache":  "invalidated",
	})
}

func evaluateReadiness(ctx context.Context, workerClient *worker.Client) (int, map[string]any, bool) {
	pg, pgErr := database.PostgresPool(ctx)
	if pgErr == nil {
		pgErr = pg.Ping(ctx)
		if pgErr != nil {
			database.ResetPostgresPool()
		}
	}

	redisClient, redisErr := database.RedisClient(ctx)
	if redisErr == nil {
		redisErr = redisClient.Ping(ctx).Err()
		if redisErr != nil {
			database.ResetRedisClient()
		}
	}

	var workerErr error
	if workerClient != nil {
		_, workerErr = workerClient.Health(ctx)
	} else {
		workerErr = context.DeadlineExceeded
	}

	status := "ok"
	httpStatus := http.StatusOK
	if pgErr != nil || redisErr != nil || workerErr != nil {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	payload := map[string]any{
		"status": status,
		"checks": map[string]string{
			"postgres": errorState(pgErr),
			"redis":    errorState(redisErr),
			"worker":   errorState(workerErr),
		},
	}

	return httpStatus, payload, pgErr != nil || redisErr != nil || workerErr != nil
}

func errorState(err error) string {
	if err != nil {
		return "down"
	}
	return "up"
}
