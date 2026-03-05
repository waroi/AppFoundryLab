package incidents

import (
	"context"

	mongostore "github.com/example/appfoundrylab/backend/services/logger/internal/mongo"
	"go.mongodb.org/mongo-driver/bson"
	drivermongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Insert(ctx context.Context, event Event) error {
	collection, err := mongostore.IncidentCollection(ctx)
	if err != nil {
		return err
	}
	_, err = collection.InsertOne(ctx, event)
	return err
}

func ListRecent(ctx context.Context, limit int64) ([]Event, error) {
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	collection, err := mongostore.IncidentCollection(ctx)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(
		ctx,
		bson.D{},
		options.Find().SetLimit(limit).SetSort(bson.D{{Key: "lastSeenAt", Value: -1}, {Key: "triggeredAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	events := make([]Event, 0, limit)
	for cursor.Next(ctx) {
		var event Event
		if err := cursor.Decode(&event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, cursor.Err()
}

func Summarize(ctx context.Context) (Summary, error) {
	collection, err := mongostore.IncidentCollection(ctx)
	if err != nil {
		return Summary{}, err
	}

	cursor, err := collection.Aggregate(ctx, summaryPipeline())
	if err != nil {
		return Summary{}, err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		if err := cursor.Err(); err != nil {
			return Summary{}, err
		}
		return Summary{}, nil
	}

	var summary Summary
	if err := cursor.Decode(&summary); err != nil {
		return Summary{}, err
	}
	return summary, cursor.Err()
}

func summaryPipeline() drivermongo.Pipeline {
	return drivermongo.Pipeline{
		{{Key: "$sort", Value: bson.D{{Key: "lastSeenAt", Value: -1}, {Key: "triggeredAt", Value: -1}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "totalEvents", Value: bson.D{{Key: "$sum", Value: 1}}},
			{Key: "activeEvents", Value: bson.D{{Key: "$sum", Value: bson.D{{Key: "$cond", Value: bson.A{
				bson.D{{Key: "$eq", Value: bson.A{"$status", "active"}}},
				1,
				0,
			}}}}}},
			{Key: "latestEventAt", Value: bson.D{{Key: "$first", Value: "$lastSeenAt"}}},
			{Key: "lastEventStatus", Value: bson.D{{Key: "$first", Value: "$status"}}},
		}}},
		{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 0},
			{Key: "totalEvents", Value: 1},
			{Key: "activeEvents", Value: 1},
			{Key: "latestEventAt", Value: 1},
			{Key: "lastEventStatus", Value: 1},
		}}},
	}
}
