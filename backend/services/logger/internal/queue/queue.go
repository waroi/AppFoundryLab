package queue

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/example/appfoundrylab/backend/services/logger/internal/ingest"
	"github.com/example/appfoundrylab/backend/services/logger/internal/mongo"
)

type Stats struct {
	QueueDepth                            int     `json:"queueDepth"`
	QueueCapacity                         int     `json:"queueCapacity"`
	Workers                               int     `json:"workers"`
	EnqueuedTotal                         uint64  `json:"enqueuedTotal"`
	DroppedTotal                          uint64  `json:"droppedTotal"`
	ProcessedTotal                        uint64  `json:"processedTotal"`
	FailedTotal                           uint64  `json:"failedTotal"`
	RetriedTotal                          uint64  `json:"retriedTotal"`
	InflightWorkers                       int64   `json:"inflightWorkers"`
	DropRatio                             float64 `json:"dropRatio"`
	DropAlertThresholdPct                 float64 `json:"dropAlertThresholdPct"`
	DropAlertThresholdHit                 bool    `json:"dropAlertThresholdHit"`
	LoggerQueueDropRatio                  float64 `json:"logger_queue_drop_ratio"`
	LoggerQueueDropAlertThresholdPct      float64 `json:"logger_queue_drop_alert_threshold_pct"`
	LoggerQueueDropAlertThresholdHit      bool    `json:"logger_queue_drop_alert_threshold_hit"`
	LoggerQueueDropAlertThresholdHitGauge int     `json:"logger_queue_drop_alert_threshold_hit_gauge"`
}

type AsyncQueue struct {
	ch       chan ingest.RequestLog
	workers  int
	retryMax int
	backoffBase time.Duration
	backoffMax  time.Duration

	wg sync.WaitGroup

	enqueued  atomic.Uint64
	dropped   atomic.Uint64
	processed atomic.Uint64
	failed    atomic.Uint64
	retried   atomic.Uint64
	inflight  atomic.Int64

	dropAlertThresholdPct float64
}

func New(size, workers, retryMax int) *AsyncQueue {
	if size <= 0 {
		size = 2048
	}
	if workers <= 0 {
		workers = 4
	}
	if retryMax < 0 {
		retryMax = 0
	}

	return &AsyncQueue{
		ch:          make(chan ingest.RequestLog, size),
		workers:     workers,
		retryMax:    retryMax,
		backoffBase: 100 * time.Millisecond,
		backoffMax:  time.Second,
	}
}

func (q *AsyncQueue) SetRetryBackoff(base, max time.Duration) {
	if base <= 0 {
		base = 100 * time.Millisecond
	}
	if max <= 0 || max < base {
		max = base
	}
	q.backoffBase = base
	q.backoffMax = max
}

func (q *AsyncQueue) Enqueue(entry ingest.RequestLog) bool {
	select {
	case q.ch <- entry:
		q.enqueued.Add(1)
		return true
	default:
		q.dropped.Add(1)
		return false
	}
}

func (q *AsyncQueue) SetDropAlertThresholdPct(percent float64) {
	if percent < 0 {
		percent = 0
	}
	q.dropAlertThresholdPct = percent
}

func (q *AsyncQueue) StartWorkers(ctx context.Context) {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(ctx)
	}
}

func (q *AsyncQueue) Wait() {
	q.wg.Wait()
}

func (q *AsyncQueue) worker(ctx context.Context) {
	defer q.wg.Done()
	for {
		select {
		case <-ctx.Done():
			q.drain()
			return
		case item := <-q.ch:
			q.process(item)
		}
	}
}

func (q *AsyncQueue) drain() {
	for {
		select {
		case item := <-q.ch:
			q.process(item)
		default:
			return
		}
	}
}

func (q *AsyncQueue) process(item ingest.RequestLog) {
	q.inflight.Add(1)
	defer q.inflight.Add(-1)

	var lastErr error
	for attempt := 0; attempt <= q.retryMax; attempt++ {
		insertCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		collection, err := mongo.Collection(insertCtx)
		if err == nil {
			_, err = collection.InsertOne(insertCtx, item)
		}
		cancel()

		if err == nil {
			q.processed.Add(1)
			return
		}

		lastErr = err
		if attempt < q.retryMax {
			q.retried.Add(1)
			time.Sleep(q.retryDelay(attempt))
		}
	}

	q.failed.Add(1)
	log.Printf("logger insert failed after retries: %v", lastErr)
}

func (q *AsyncQueue) retryDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return q.backoffBase
	}

	delay := q.backoffBase
	for i := 0; i < attempt; i++ {
		if delay >= q.backoffMax {
			return q.backoffMax
		}
		delay *= 2
	}
	if delay > q.backoffMax {
		return q.backoffMax
	}
	return delay
}

func (q *AsyncQueue) Stats() Stats {
	enqueued := q.enqueued.Load()
	dropped := q.dropped.Load()
	total := enqueued + dropped
	dropRatio := 0.0
	if total > 0 {
		dropRatio = float64(dropped) / float64(total)
	}
	thresholdHit := false
	if q.dropAlertThresholdPct > 0 {
		thresholdHit = (dropRatio * 100) >= q.dropAlertThresholdPct
	}

	return Stats{
		QueueDepth:                            len(q.ch),
		QueueCapacity:                         cap(q.ch),
		Workers:                               q.workers,
		EnqueuedTotal:                         enqueued,
		DroppedTotal:                          dropped,
		ProcessedTotal:                        q.processed.Load(),
		FailedTotal:                           q.failed.Load(),
		RetriedTotal:                          q.retried.Load(),
		InflightWorkers:                       q.inflight.Load(),
		DropRatio:                             dropRatio,
		DropAlertThresholdPct:                 q.dropAlertThresholdPct,
		DropAlertThresholdHit:                 thresholdHit,
		LoggerQueueDropRatio:                  dropRatio,
		LoggerQueueDropAlertThresholdPct:      q.dropAlertThresholdPct,
		LoggerQueueDropAlertThresholdHit:      thresholdHit,
		LoggerQueueDropAlertThresholdHitGauge: boolToGauge(thresholdHit),
	}
}

func (q *AsyncQueue) PrometheusMetrics() string {
	stats := q.Stats()
	lines := []string{
		"# HELP logger_queue_depth Current logger queue depth.",
		"# TYPE logger_queue_depth gauge",
		fmt.Sprintf("logger_queue_depth %d", stats.QueueDepth),
		"# HELP logger_queue_capacity Logger queue capacity.",
		"# TYPE logger_queue_capacity gauge",
		fmt.Sprintf("logger_queue_capacity %d", stats.QueueCapacity),
		"# HELP logger_workers Configured logger worker count.",
		"# TYPE logger_workers gauge",
		fmt.Sprintf("logger_workers %d", stats.Workers),
		"# HELP logger_enqueued_total Total log entries accepted by the queue.",
		"# TYPE logger_enqueued_total counter",
		fmt.Sprintf("logger_enqueued_total %d", stats.EnqueuedTotal),
		"# HELP logger_dropped_total Total log entries dropped because the queue was full.",
		"# TYPE logger_dropped_total counter",
		fmt.Sprintf("logger_dropped_total %d", stats.DroppedTotal),
		"# HELP logger_processed_total Total log entries written successfully.",
		"# TYPE logger_processed_total counter",
		fmt.Sprintf("logger_processed_total %d", stats.ProcessedTotal),
		"# HELP logger_failed_total Total log entries that failed after retries.",
		"# TYPE logger_failed_total counter",
		fmt.Sprintf("logger_failed_total %d", stats.FailedTotal),
		"# HELP logger_retried_total Total retry attempts for logger queue writes.",
		"# TYPE logger_retried_total counter",
		fmt.Sprintf("logger_retried_total %d", stats.RetriedTotal),
		"# HELP logger_inflight_workers Current number of workers processing writes.",
		"# TYPE logger_inflight_workers gauge",
		fmt.Sprintf("logger_inflight_workers %d", stats.InflightWorkers),
		"# HELP logger_queue_drop_ratio Ratio of dropped entries to total queue attempts.",
		"# TYPE logger_queue_drop_ratio gauge",
		fmt.Sprintf("logger_queue_drop_ratio %.6f", stats.LoggerQueueDropRatio),
		"# HELP logger_queue_drop_alert_threshold_pct Configured drop alert threshold percentage.",
		"# TYPE logger_queue_drop_alert_threshold_pct gauge",
		fmt.Sprintf("logger_queue_drop_alert_threshold_pct %.2f", stats.LoggerQueueDropAlertThresholdPct),
		"# HELP logger_queue_drop_alert_threshold_hit_gauge Whether the drop alert threshold is currently breached.",
		"# TYPE logger_queue_drop_alert_threshold_hit_gauge gauge",
		fmt.Sprintf("logger_queue_drop_alert_threshold_hit_gauge %d", stats.LoggerQueueDropAlertThresholdHitGauge),
	}
	return strings.Join(lines, "\n") + "\n"
}

func boolToGauge(v bool) int {
	if v {
		return 1
	}
	return 0
}
