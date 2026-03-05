package queue

import (
	"strings"
	"testing"

	"github.com/example/appfoundrylab/backend/services/logger/internal/ingest"
)

func TestStatsIncludesDropRatioAndThresholdAlert(t *testing.T) {
	q := New(1, 1, 0)
	q.SetDropAlertThresholdPct(40)

	if ok := q.Enqueue(ingest.RequestLog{}); !ok {
		t.Fatal("expected first enqueue to succeed")
	}
	if ok := q.Enqueue(ingest.RequestLog{}); ok {
		t.Fatal("expected second enqueue to be dropped")
	}

	stats := q.Stats()
	if stats.DroppedTotal != 1 || stats.EnqueuedTotal != 1 {
		t.Fatalf("unexpected counters: %+v", stats)
	}
	if stats.DropRatio <= 0 {
		t.Fatalf("expected positive drop ratio, got %f", stats.DropRatio)
	}
	if !stats.DropAlertThresholdHit {
		t.Fatal("expected threshold hit for drop ratio")
	}
	if !stats.LoggerQueueDropAlertThresholdHit {
		t.Fatal("expected prometheus-compatible threshold hit")
	}
	if stats.LoggerQueueDropAlertThresholdHitGauge != 1 {
		t.Fatalf("expected threshold hit gauge=1, got=%d", stats.LoggerQueueDropAlertThresholdHitGauge)
	}
	if stats.LoggerQueueDropRatio != stats.DropRatio {
		t.Fatalf("expected prometheus drop ratio to mirror drop ratio, got=%f want=%f", stats.LoggerQueueDropRatio, stats.DropRatio)
	}
}

func TestStatsThresholdDisabled(t *testing.T) {
	q := New(1, 1, 0)
	q.SetDropAlertThresholdPct(0)

	_ = q.Enqueue(ingest.RequestLog{})
	_ = q.Enqueue(ingest.RequestLog{})

	stats := q.Stats()
	if stats.DropAlertThresholdHit {
		t.Fatal("expected threshold alert disabled when threshold is zero")
	}
	if stats.LoggerQueueDropAlertThresholdHit {
		t.Fatal("expected prometheus-compatible threshold alert disabled when threshold is zero")
	}
	if stats.LoggerQueueDropAlertThresholdHitGauge != 0 {
		t.Fatalf("expected threshold hit gauge=0, got=%d", stats.LoggerQueueDropAlertThresholdHitGauge)
	}
}

func TestPrometheusMetricsContainsExpectedSeries(t *testing.T) {
	q := New(4, 2, 1)
	_ = q.Enqueue(ingest.RequestLog{})

	metrics := q.PrometheusMetrics()
	for _, token := range []string{
		"logger_queue_depth",
		"logger_queue_capacity",
		"logger_enqueued_total",
		"logger_queue_drop_ratio",
	} {
		if !strings.Contains(metrics, token) {
			t.Fatalf("expected prometheus metrics output to contain %q, got=%s", token, metrics)
		}
	}
}
