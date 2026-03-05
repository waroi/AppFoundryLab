package requestlogs

import (
	"context"

	"github.com/example/appfoundrylab/backend/services/logger/internal/ingest"
	mongostore "github.com/example/appfoundrylab/backend/services/logger/internal/mongo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ListRecent(ctx context.Context, limit int64, traceID string) ([]ingest.RequestLog, error) {
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	collection, err := mongostore.Collection(ctx)
	if err != nil {
		return nil, err
	}

	filter := bson.D{}
	if traceID != "" {
		filter = append(filter, bson.E{Key: "traceId", Value: traceID})
	}

	cursor, err := collection.Find(
		ctx,
		filter,
		options.Find().SetLimit(limit).SetSort(bson.D{{Key: "occurredAt", Value: -1}}),
	)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	items := make([]ingest.RequestLog, 0, limit)
	for cursor.Next(ctx) {
		var item ingest.RequestLog
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, cursor.Err()
}
