package mongo

import "testing"

func TestIndexModelsForCollection(t *testing.T) {
	requestLogModels := indexModelsForCollection("request_logs")
	if len(requestLogModels) != 2 {
		t.Fatalf("expected 2 request log indexes, got %d", len(requestLogModels))
	}
	if name := requestLogModels[0].Options.Name; name == nil || *name != "request_logs_occurredAt_desc" {
		t.Fatalf("unexpected request log index name: %v", name)
	}
	if name := requestLogModels[1].Options.Name; name == nil || *name != "request_logs_traceId_occurredAt_desc" {
		t.Fatalf("unexpected request log trace index name: %v", name)
	}

	incidentModels := indexModelsForCollection("incident_events")
	if len(incidentModels) != 2 {
		t.Fatalf("expected 2 incident indexes, got %d", len(incidentModels))
	}
	if name := incidentModels[0].Options.Name; name == nil || *name != "incident_events_lastSeenAt_triggeredAt_desc" {
		t.Fatalf("unexpected incident index name: %v", name)
	}
	if name := incidentModels[1].Options.Name; name == nil || *name != "incident_events_status_lastSeenAt_desc" {
		t.Fatalf("unexpected incident status index name: %v", name)
	}

	if got := indexModelsForCollection("unknown"); got != nil {
		t.Fatalf("expected nil models for unknown collection, got %d", len(got))
	}
}
