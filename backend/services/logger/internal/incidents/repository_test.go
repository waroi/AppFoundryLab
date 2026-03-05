package incidents

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestSummaryPipeline(t *testing.T) {
	pipeline := summaryPipeline()
	if len(pipeline) != 3 {
		t.Fatalf("expected 3 aggregation stages, got %d", len(pipeline))
	}

	expectedSort := bson.D{{Key: "$sort", Value: bson.D{{Key: "lastSeenAt", Value: -1}, {Key: "triggeredAt", Value: -1}}}}
	if !reflect.DeepEqual(pipeline[0], expectedSort) {
		t.Fatalf("unexpected sort stage: %#v", pipeline[0])
	}

	groupStage := pipeline[1].Map()["$group"]
	groupDoc, ok := groupStage.(bson.D)
	if !ok {
		t.Fatalf("expected group stage to be bson.D, got %T", groupStage)
	}
	groupMap := groupDoc.Map()
	if _, ok := groupMap["totalEvents"]; !ok {
		t.Fatal("expected totalEvents accumulator in group stage")
	}
	if _, ok := groupMap["activeEvents"]; !ok {
		t.Fatal("expected activeEvents accumulator in group stage")
	}
	if _, ok := groupMap["latestEventAt"]; !ok {
		t.Fatal("expected latestEventAt accumulator in group stage")
	}
	if _, ok := groupMap["lastEventStatus"]; !ok {
		t.Fatal("expected lastEventStatus accumulator in group stage")
	}

	expectedProject := bson.D{{Key: "$project", Value: bson.D{
		{Key: "_id", Value: 0},
		{Key: "totalEvents", Value: 1},
		{Key: "activeEvents", Value: 1},
		{Key: "latestEventAt", Value: 1},
		{Key: "lastEventStatus", Value: 1},
	}}}
	if !reflect.DeepEqual(pipeline[2], expectedProject) {
		t.Fatalf("unexpected project stage: %#v", pipeline[2])
	}
}
