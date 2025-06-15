package mongo

import (
	"context"
	"log"
	"time"

	"github.com/danmuck/dps_http/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoStore struct {
	name     string
	location string
	t        string // e.g., "MongoDB"
	client   *mongo.Client
	db       *mongo.Database
	buckets  map[string]*MongoBucket
}

func NewMongoStore(uri, dbName string) (*MongoStore, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	log.Println("pinging ...")
	if err := client.Ping(ctx, nil); err != nil {
		log.Println("failed to ping MongoDB:", err)
		return nil, err
	}
	log.Println("connected to MongoDB at", uri)
	db := client.Database(dbName)
	return &MongoStore{
		name:    dbName,
		t:       "mongo",
		client:  client,
		db:      db,
		buckets: make(map[string]*MongoBucket),
	}, nil
}
func (ms *MongoStore) Name() string {
	return ms.name
}
func (ms *MongoStore) Type() string {
	return ms.t
}
func (md *MongoStore) Location() string {
	return md.location
}
func (ms *MongoStore) Ping(ctx context.Context) error {
	return ms.client.Ping(ctx, nil)
}

// ConnectOrCreateBucket connects to an existing bucket or creates a new one if it doesn't exist.
// It returns the collection for the specified bucket.
func (ms *MongoStore) ConnectOrCreateBucket(bucket string) storage.Bucket {
	log.Printf("Connect: to bucket: %s", bucket)
	collection, exists := ms.buckets[bucket]
	if !exists || collection == nil {
		log.Println("Connect: creating Bucket:", bucket)
		collection = NewMongoBucket(ms.db, bucket)
		ms.buckets[bucket] = collection
	}
	log.Printf("Connect: using collection: %s", collection.Name())
	return collection
}

// basic CRUD operations
func (ms *MongoStore) Store(bucket string, key string, value any) error {
	log.Printf("Storing key %q in bucket %q", key, bucket)
	collection := ms.ConnectOrCreateBucket(bucket)
	err := collection.Store(key, value)
	log.Printf("Store result: %v (success if <nil>)", err)
	return err
}

// retreieves a value by key from a bucket
// note: this wraps Lookup() with a specific filter for the key
// it returns the value directly as `any` type
func (ms *MongoStore) Retrieve(bucket string, key string) (any, error) {
	log.Printf("Retrieving key %q from bucket %q", key, bucket)
	result, _ := ms.Lookup(bucket, bson.M{"key": key})
	return result["value"], nil
}

// TODO: implement
func (ms *MongoStore) Delete(bucket string, key string) error {
	collection := ms.ConnectOrCreateBucket(bucket)
	err := collection.Delete(key)
	return err
}

// TODO: implement
func (ms *MongoStore) Update(bucket string, key string, value any) error {
	log.Printf("Updating key %q in bucket %q with value: %v", key, bucket, value)
	collection := ms.ConnectOrCreateBucket(bucket)
	return collection.Update(key, value)
}
func (ms *MongoStore) Patch(bucket, key string, updates map[string]any) error {
	log.Printf("Patching key %q in bucket %q with updates: %v", key, bucket, updates)
	collection := ms.ConnectOrCreateBucket(bucket)
	return collection.Patch(key, updates)
}

func (ms *MongoStore) Lookup(bucket string, filter any) (map[string]any, bool) {
	log.Printf("Lookup: in bucket %q with filter: %v", bucket, filter)
	collection := ms.ConnectOrCreateBucket(bucket)
	return collection.Lookup(filter)
}

func (ms *MongoStore) List(bucket string) ([]any, error) {
	log.Println("List: all keys in bucket:", bucket)
	collection := ms.ConnectOrCreateBucket(bucket)
	return collection.List(bucket)
}

func (ms *MongoStore) Count(bucket string) (int64, error) {
	log.Printf("Counting documents in bucket: %s", bucket)
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
