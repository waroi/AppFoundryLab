package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoOnce sync.Once
	mongoDB   *mongo.Database
	mongoErr  error
)

func Collection(ctx context.Context) (*mongo.Collection, error) {
	return CollectionByName(ctx, env.GetWithDefault("MONGO_COLLECTION", "request_logs"))
}

func IncidentCollection(ctx context.Context) (*mongo.Collection, error) {
	return CollectionByName(ctx, env.GetWithDefault("MONGO_INCIDENT_COLLECTION", "incident_events"))
}

func CollectionByName(ctx context.Context, collectionName string) (*mongo.Collection, error) {
	mongoOnce.Do(func() {
		uri := fmt.Sprintf("mongodb://%s:%s@%s:%s",
			env.MustGet("MONGO_INITDB_ROOT_USERNAME"),
			env.MustGet("MONGO_INITDB_ROOT_PASSWORD"),
			env.MustGet("MONGO_HOST"),
			env.GetWithDefault("MONGO_PORT", "27017"),
		)

		client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetConnectTimeout(3*time.Second))
		if err != nil {
			mongoErr = err
			return
		}

		dbName := env.GetWithDefault("MONGO_DB", "appfoundrylab_logs")
		mongoDB = client.Database(dbName)
	})

	if mongoErr != nil {
		return nil, mongoErr
	}
	collection := mongoDB.Collection(collectionName)
	if err := ensureCollectionIndexes(ctx, collection, collectionName); err != nil {
		return nil, err
	}
	return collection, nil
}
