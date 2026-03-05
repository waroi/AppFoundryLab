package metrics

import "testing"

func TestStoreInflightTracking(t *testing.T) {
	store := NewStore()

	store.IncInflight()
	store.IncInflight()
	store.DecInflight()

	snapshot := store.Snapshot()
	if snapshot.InflightCurrent != 1 {
		t.Fatalf("expected inflight current=1, got %d", snapshot.InflightCurrent)
	}
	if snapshot.InflightPeak != 2 {
		t.Fatalf("expected inflight peak=2, got %d", snapshot.InflightPeak)
	}
}

func TestStoreLoadShedCounter(t *testing.T) {
	store := NewStore()
	store.ObserveLoadShed()
	store.ObserveLoadShed()

	snapshot := store.Snapshot()
	if snapshot.LoadShedTotal != 2 {
		t.Fatalf("expected load shed total=2, got %d", snapshot.LoadShedTotal)
	}
}

func TestStoreSnapshotIncludesRecentHistory(t *testing.T) {
	store := NewStore()
	store.ObserveLoadShed()
	store.Observe(200, 10)
	store.IncInflight()

	snapshot := store.Snapshot()
	if len(snapshot.RecentHistory) == 0 {
		t.Fatal("expected recent history to be populated")
	}

	last := snapshot.RecentHistory[len(snapshot.RecentHistory)-1]
	if last.RequestsTotal == 0 {
		t.Fatalf("expected requests total in history, got %+v", last)
	}
}
