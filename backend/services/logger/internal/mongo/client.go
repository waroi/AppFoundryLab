package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/example/appfoundrylab/backend/pkg/env"
	"github.com/example/appfoundrylab/backend/pkg/retryutil"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoMu      sync.Mutex
	mongoConn    *mongo.Client
	mongoDB      *mongo.Database
	connectMongo = func(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
		return mongo.Connect(ctx, opts...)
	}
	pingMongo = func(ctx context.Context, client *mongo.Client) error {
		return client.Ping(ctx, nil)
	}
	disconnectMongo = func(ctx context.Context, client *mongo.Client) error {
		return client.Disconnect(ctx)
	}
)

func Collection(ctx context.Context) (*mongo.Collection, error) {
	return CollectionByName(ctx, env.GetWithDefault("MONGO_COLLECTION", "request_logs"))
}

func IncidentCollection(ctx context.Context) (*mongo.Collection, error) {
	return CollectionByName(ctx, env.GetWithDefault("MONGO_INCIDENT_COLLECTION", "incident_events"))
}

func CollectionByName(ctx context.Context, collectionName string) (*mongo.Collection, error) {
	db, err := database(ctx)
	if err != nil {
		return nil, err
	}

	collection := db.Collection(collectionName)
	if err := ensureCollectionIndexes(ctx, collection, collectionName); err != nil {
		return nil, err
	}
	return collection, nil
}

func Health(ctx context.Context) error {
	mongoMu.Lock()
	client := mongoConn
	mongoMu.Unlock()

	if client == nil {
		_, err := database(ctx)
		return err
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout())
	defer cancel()
	if err := pingMongo(pingCtx, client); err != nil {
		ResetClient()
		_, reconnectErr := database(ctx)
		if reconnectErr != nil {
			return reconnectErr
		}
		return nil
	}
	return nil
}

func ResetClient() {
	mongoMu.Lock()
	defer mongoMu.Unlock()

	if mongoConn != nil {
		disconnectCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		_ = disconnectMongo(disconnectCtx, mongoConn)
	}

	mongoConn = nil
	mongoDB = nil
}

func database(ctx context.Context) (*mongo.Database, error) {
	mongoMu.Lock()
	defer mongoMu.Unlock()

	if mongoDB != nil {
		return mongoDB, nil
	}

	client, err := retryutil.Do(ctx, dependencyConnectAttempts(), dependencyConnectBackoff(), func(attemptCtx context.Context) (*mongo.Client, error) {
		candidate, err := connectMongo(attemptCtx, options.Client().ApplyURI(mongoURI()).SetConnectTimeout(3*time.Second))
		if err != nil {
			return nil, err
		}

		pingCtx, cancel := context.WithTimeout(attemptCtx, pingTimeout())
		defer cancel()
		if err := pingMongo(pingCtx, candidate); err != nil {
			_ = disconnectMongo(context.Background(), candidate)
			return nil, err
		}

		return candidate, nil
	})
	if err != nil {
		return nil, err
	}

	mongoConn = client
	mongoDB = client.Database(env.GetWithDefault("MONGO_DB", "appfoundrylab_logs"))
	return mongoDB, nil
}

func mongoURI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%s",
		env.MustGet("MONGO_INITDB_ROOT_USERNAME"),
		env.MustGet("MONGO_INITDB_ROOT_PASSWORD"),
		env.MustGet("MONGO_HOST"),
		env.GetWithDefault("MONGO_PORT", "27017"),
	)
}

func dependencyConnectAttempts() int {
	attempts := env.GetIntWithDefault("DEPENDENCY_CONNECT_MAX_ATTEMPTS", 4)
	if attempts < 1 {
		return 1
	}
	return attempts
}

func dependencyConnectBackoff() time.Duration {
	backoffMS := env.GetIntWithDefault("DEPENDENCY_CONNECT_BACKOFF_MS", 250)
	if backoffMS < 0 {
		backoffMS = 0
	}
	return time.Duration(backoffMS) * time.Millisecond
}

func pingTimeout() time.Duration {
	timeoutMS := env.GetIntWithDefault("DEPENDENCY_PING_TIMEOUT_MS", 1500)
	if timeoutMS <= 0 {
		timeoutMS = 1500
	}
	return time.Duration(timeoutMS) * time.Millisecond
}
