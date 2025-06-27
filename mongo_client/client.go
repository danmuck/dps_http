package mongo_client

import (
	"context"
	"time"

	"github.com/danmuck/dps_lib/logs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	name   string
	client *mongo.Client
	db     *mongo.Database
}

func NewMongoStore(uri, dbName string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		logs.Log("failed to connect to MongoDB at %s: %v", uri, err)
		return nil, err
	}
	logs.Log("connecting to MongoDB at %s", uri)
	db := client.Database(dbName)
	return &MongoClient{
		name:   dbName,
		client: client,
		db:     db,
	}, nil
}

// Name returns the name of the MongoDB Database
func (ms *MongoClient) Name() string {
	return ms.name
}
func (ms *MongoClient) Client() *mongo.Client {
	return ms.client
}
func (ms *MongoClient) Database() *mongo.Database {
	return ms.db
}
func (ms *MongoClient) Collection(name string) *mongo.Collection {
	return ms.db.Collection(name)
}

// MongoDB client Ping wrapper
func (ms *MongoClient) Ping(ctx context.Context) error {
	return ms.client.Ping(ctx, nil)
}
