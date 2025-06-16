package mongo

import (
	"context"
	"time"

	"github.com/danmuck/dps_http/api/logs"
	"github.com/danmuck/dps_http/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	name     string
	location string
	t        string // e.g., "MongoDB"
	client   *mongo.Client
	db       *mongo.Database
	buckets  map[string]*MongoBucket
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
		name:    dbName,
		t:       "mongo",
		client:  client,
		db:      db,
		buckets: make(map[string]*MongoBucket),
	}, nil
}
func (ms *MongoClient) Name() string {
	return ms.name
}
func (ms *MongoClient) Type() string {
	return ms.t
}
func (md *MongoClient) Location() string {
	return md.location
}
func (ms *MongoClient) Ping(ctx context.Context) error {
	return ms.client.Ping(ctx, nil)
}

// ConnectOrCreateBucket connects to an existing bucket or creates a new one if it doesn't exist.
// It returns the collection for the specified bucket.
func (ms *MongoClient) ConnectOrCreateBucket(bucket string) storage.Bucket {
	logs.Init("ConnectOrCreateBucket [%s]", bucket)
	collection, exists := ms.buckets[bucket]
	if !exists || collection == nil {
		logs.Log("Create [%s]", bucket)
		collection = NewMongoBucket(ms.db, bucket)
		ms.buckets[bucket] = collection
	}
	logs.Log("Connect [%s]", bucket)
	return collection
}

// basic CRUD operations
func (ms *MongoClient) Store(bucket string, key string, value any) error {
	logs.Init("Store [%q] : { %q : %q }", bucket, key, value)
	collection := ms.ConnectOrCreateBucket(bucket)
	err := collection.Store(key, value)
	logs.Log("Store result: %v (success if <nil>)", err)
	return err
}

// retreieves a value by key from a bucket
// note: this wraps Lookup() with a specific filter for the key
// it returns the value directly as `any` type
func (ms *MongoClient) Retrieve(bucket string, key string) (any, error) {
	logs.Init("Retrieve [%q] : { %q }", bucket, key)
	result, _ := ms.Lookup(bucket, bson.M{"key": key})
	return result["value"], nil
}

// TODO: implement
func (ms *MongoClient) Delete(bucket string, key string) error {
	logs.Init("Delete [%q] key %q", bucket, key)
	collection := ms.ConnectOrCreateBucket(bucket)
	err := collection.Delete(key)
	return err
}

// TODO: implement
func (ms *MongoClient) Update(bucket string, key string, value any) error {
	logs.Init("Update [%q] { %q : %v }", bucket, key, value)
	collection := ms.ConnectOrCreateBucket(bucket)
	return collection.Update(key, value)
}
func (ms *MongoClient) Patch(bucket, key string, updates map[string]any) error {
	logs.Init("Patch [%q] key %q updates: %v", bucket, key, updates)
	collection := ms.ConnectOrCreateBucket(bucket)
	return collection.Patch(key, updates)
}

func (ms *MongoClient) Lookup(bucket string, filter any) (map[string]any, bool) {
	logs.Init("Lookup [%q] filter: %v", bucket, filter)
	collection := ms.ConnectOrCreateBucket(bucket)
	return collection.Lookup(filter)
}

func (ms *MongoClient) List(bucket string) ([]any, error) {
	logs.Init("List [%s]", bucket)
	collection := ms.ConnectOrCreateBucket(bucket)
	return collection.ListKeys()
}

func (ms *MongoClient) Count(bucket string) (int64, error) {
	logs.Init("Count [%s]", bucket)
	collection := ms.ConnectOrCreateBucket(bucket)
	return collection.Count()
}

//		Example filters:
//	       filter := bson.M{
//				"key": "someKey",
//			}
//
//	       OR
//
//			filter := bson.M{
//				"$or": []any{
//					bson.M{"username": "someKey"},
//					bson.M{"email": "someValue"},
//				},
//				"value.status": "active",
//			}
//			result, found := ms.Lookup("myBucket", filter)
//
// //
