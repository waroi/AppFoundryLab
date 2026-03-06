package mongo

import (
	"context"
	"errors"
	"testing"

	drivermongo "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestHealthReconnectsWhenCurrentClientFails(t *testing.T) {
	t.Setenv("MONGO_INITDB_ROOT_USERNAME", "root")
	t.Setenv("MONGO_INITDB_ROOT_PASSWORD", "password")
	t.Setenv("MONGO_HOST", "mongo")
	t.Setenv("MONGO_PORT", "27017")
	t.Setenv("MONGO_DB", "appfoundrylab_logs")

	ResetClient()
	originalConnect := connectMongo
	originalPing := pingMongo
	originalDisconnect := disconnectMongo
	t.Cleanup(func() {
		ResetClient()
		connectMongo = originalConnect
		pingMongo = originalPing
		disconnectMongo = originalDisconnect
	})

	mongoConn = &drivermongo.Client{}
	mongoDB = mongoConn.Database("stale")

	pingCalls := 0
	connectCalls := 0
	disconnectCalls := 0

	connectMongo = func(context.Context, ...*options.ClientOptions) (*drivermongo.Client, error) {
		connectCalls++
		return &drivermongo.Client{}, nil
	}
	pingMongo = func(context.Context, *drivermongo.Client) error {
		pingCalls++
		if pingCalls == 1 {
			return errors.New("mongo down")
		}
		return nil
	}
	disconnectMongo = func(context.Context, *drivermongo.Client) error {
		disconnectCalls++
		return nil
	}

	if err := Health(context.Background()); err != nil {
		t.Fatalf("expected health recovery, got %v", err)
	}
	if connectCalls != 1 {
		t.Fatalf("expected one reconnect attempt, got %d", connectCalls)
	}
	if disconnectCalls == 0 {
		t.Fatal("expected stale client to be disconnected")
	}
}
