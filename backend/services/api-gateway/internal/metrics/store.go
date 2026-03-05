package metrics

import (
	"sync"
	"time"
)

var defaultBucketsMS = []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000}

type HistogramBucket struct {
	UpperBoundMS float64
	Count        uint64
}

type Snapshot struct {
	RequestsTotal     uint64
	RequestErrors     uint64
	ErrorRate         float64
	LatencyCount      uint64
	LatencySumMS      float64
	LatencyBucketsMS  []HistogramBucket
	LatencyOverflowMS uint64
	LoadShedTotal     uint64
	InflightCurrent   int64
	InflightPeak      int64
	RecentHistory     []TrendPoint
}

type Store struct {
	mu sync.Mutex

	requestsTotal uint64
	requestErrors uint64
	latencyCount  uint64
	latencySumMS  float64
	bucketsMS     []HistogramBucket
	overflowCount uint64
	loadShedTotal uint64
	inflight      int64
	inflightPeak  int64
	history       []TrendPoint
	lastHistoryAt time.Time
}

type TrendPoint struct {
	RecordedAt       string  `json:"recordedAt"`
	RequestsTotal    uint64  `json:"requestsTotal"`
	RequestErrors    uint64  `json:"requestErrors"`
	ErrorRate        float64 `json:"errorRate"`
	LatencyAverageMS float64 `json:"latencyAverageMs"`
	LoadShedTotal    uint64  `json:"loadShedTotal"`
	InflightCurrent  int64   `json:"inflightCurrent"`
	InflightPeak     int64   `json:"inflightPeak"`
}

const (
	historyLimit       = 24
	historyMinInterval = 5 * time.Second
)

func NewStore() *Store {
	buckets := make([]HistogramBucket, 0, len(defaultBucketsMS))
	for _, upper := range defaultBucketsMS {
		buckets = append(buckets, HistogramBucket{UpperBoundMS: upper})
	}
	return &Store{bucketsMS: buckets}
}

func (s *Store) Observe(statusCode int, duration time.Duration) {
	ms := float64(duration.Milliseconds())
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	s.requestsTotal++
	if statusCode >= 500 {
		s.requestErrors++
	}
	s.latencyCount++
	s.latencySumMS += ms

	for i := range s.bucketsMS {
		if ms <= s.bucketsMS[i].UpperBoundMS {
			s.bucketsMS[i].Count++
			s.recordTrendLocked(now, false)
			return
		}
	}
	s.overflowCount++
	s.recordTrendLocked(now, false)
}

func (s *Store) ObserveLoadShed() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loadShedTotal++
	s.recordTrendLocked(time.Now().UTC(), false)
}

func (s *Store) IncInflight() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.inflight++
	if s.inflight > s.inflightPeak {
		s.inflightPeak = s.inflight
	}
	s.recordTrendLocked(time.Now().UTC(), false)
}

func (s *Store) DecInflight() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.inflight > 0 {
		s.inflight--
	}
	s.recordTrendLocked(time.Now().UTC(), false)
}

func (s *Store) Snapshot() Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.recordTrendLocked(time.Now().UTC(), true)

	buckets := make([]HistogramBucket, len(s.bucketsMS))
	copy(buckets, s.bucketsMS)
	history := make([]TrendPoint, len(s.history))
	copy(history, s.history)

	errorRate := 0.0
	if s.requestsTotal > 0 {
		errorRate = float64(s.requestErrors) / float64(s.requestsTotal)
	}

	return Snapshot{
		RequestsTotal:     s.requestsTotal,
		RequestErrors:     s.requestErrors,
		ErrorRate:         errorRate,
		LatencyCount:      s.latencyCount,
		LatencySumMS:      s.latencySumMS,
		LatencyBucketsMS:  buckets,
		LatencyOverflowMS: s.overflowCount,
		LoadShedTotal:     s.loadShedTotal,
		InflightCurrent:   s.inflight,
		InflightPeak:      s.inflightPeak,
		RecentHistory:     history,
	}
}

func (s *Store) recordTrendLocked(now time.Time, force bool) {
	if !force && !s.lastHistoryAt.IsZero() && now.Sub(s.lastHistoryAt) < historyMinInterval {
		return
	}

	errorRate := 0.0
	latencyAverage := 0.0
	if s.requestsTotal > 0 {
		errorRate = float64(s.requestErrors) / float64(s.requestsTotal)
	}
	if s.latencyCount > 0 {
		latencyAverage = s.latencySumMS / float64(s.latencyCount)
	}

	point := TrendPoint{
		RecordedAt:       now.Format(time.RFC3339Nano),
		RequestsTotal:    s.requestsTotal,
		RequestErrors:    s.requestErrors,
		ErrorRate:        errorRate,
		LatencyAverageMS: latencyAverage,
		LoadShedTotal:    s.loadShedTotal,
		InflightCurrent:  s.inflight,
		InflightPeak:     s.inflightPeak,
	}

	if len(s.history) > 0 && force {
		last := s.history[len(s.history)-1]
		if last.RecordedAt == point.RecordedAt {
			s.history[len(s.history)-1] = point
			s.lastHistoryAt = now
			return
		}
	}

	if len(s.history) == historyLimit {
		copy(s.history, s.history[1:])
		s.history[len(s.history)-1] = point
	} else {
		s.history = append(s.history, point)
	}
	s.lastHistoryAt = now
}
