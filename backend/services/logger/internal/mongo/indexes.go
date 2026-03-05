package mongo

import (
	"context"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	drivermongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type collectionIndexState struct {
	mu    sync.Mutex
	ready bool
}

var collectionIndexes sync.Map

func ensureCollectionIndexes(ctx context.Context, collection *drivermongo.Collection, collectionName string) error {
	stateValue, _ := collectionIndexes.LoadOrStore(collectionName, &collectionIndexState{})
	state := stateValue.(*collectionIndexState)

	state.mu.Lock()
	defer state.mu.Unlock()
	if state.ready {
		return nil
	}

	models := indexModelsForCollection(collectionName)
	if len(models) == 0 {
		state.ready = true
		return nil
	}

	if _, err := collection.Indexes().CreateMany(ctx, models); err != nil {
		return err
	}
	state.ready = true
	return nil
}

func indexModelsForCollection(collectionName string) []drivermongo.IndexModel {
	switch collectionName {
	case "request_logs":
		return []drivermongo.IndexModel{
			{
				Keys:    bson.D{{Key: "occurredAt", Value: -1}},
				Options: options.Index().SetName("request_logs_occurredAt_desc"),
			},
			{
				Keys:    bson.D{{Key: "traceId", Value: 1}, {Key: "occurredAt", Value: -1}},
				Options: options.Index().SetName("request_logs_traceId_occurredAt_desc").SetSparse(true),
			},
		}
	case "incident_events":
		return []drivermongo.IndexModel{
			{
				Keys:    bson.D{{Key: "lastSeenAt", Value: -1}, {Key: "triggeredAt", Value: -1}},
				Options: options.Index().SetName("incident_events_lastSeenAt_triggeredAt_desc"),
			},
			{
				Keys:    bson.D{{Key: "status", Value: 1}, {Key: "lastSeenAt", Value: -1}},
				Options: options.Index().SetName("incident_events_status_lastSeenAt_desc"),
			},
		}
	default:
		return nil
	}
}
