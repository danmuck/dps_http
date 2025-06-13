package storage

import (
	"context"
	"log"
	"slices"
	// "strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var forbidden []string = []string{
	"password_hash", "token", "created_at", "updated_at",
	"bio", "avatar_url", "email",
}

var allowed []string = []string{
	"username", "roles",
}

type MongoStore struct {
	name    string
	t       string // e.g., "MongoDB"
	client  *mongo.Client
	db      *mongo.Database
	buckets map[string]*mongo.Collection
}

func (ms *MongoStore) Name() string {
	return ms.name
}
func (ms *MongoStore) Type() string {
	return ms.t
}
func (ms *MongoStore) Client() StorageClient {
	return ms.client
}
func (ms *MongoStore) Database() Database {
	return ms.db
}
func (ms *MongoStore) Ping(ctx context.Context) error {
	return ms.client.Ping(ctx, nil)
}

// Basic CRUD operations
func (ms *MongoStore) Store(bucket string, key string, value any) error {
	log.Printf("Storing key %q with value %v in bucket %q", key, value, bucket)
	collection := ms.connectOrCreatBucket(bucket)

	_, err := collection.InsertOne(context.Background(), map[string]any{
		"key":   key,
		"value": value,
	})
	log.Printf("Store result: %v", err)
	return err
}

/*
*	Expects key to be a string and a username to retrieve
 */
func (ms *MongoStore) Retrieve(bucket string, key string) (any, error) {
	log.Printf("Retrieving key %q from bucket %q", key, bucket)

	result, _ := ms.Lookup(bucket, map[string]any{"key": key})

	return result["value"], nil
}

func (ms *MongoStore) Delete(bucket string, key string) error {
	collection := ms.connectOrCreatBucket(bucket)

	_, err := collection.DeleteOne(context.Background(), map[string]any{"key": key})
	return err
}

func (ms *MongoStore) Update(bucket string, key string, value any) error {
	collection := ms.connectOrCreatBucket(bucket)

	_, err := collection.UpdateOne(context.Background(), map[string]any{"key": key}, map[string]any{
		"$set": map[string]any{"value": value},
	})
	return err
}

/**************************************************************************************
*
*	Lookup retrieves a specific key in a bucket with a filter
*
*	The filter is expected to be a map-like structure (e.g., bson.M) that can be used to build a MongoDB query.
*	It returns the first matching document as a map[string]any and a boolean indicating if it was found.
*	If no document matches the filter, it returns nil and false.
*
*	The filter can include MongoDB operators like $or, $and, etc., and will be prefixed with "value." for fields.
*
*	Note: This function iterates through the filter to build a complete MongoDB query and it does **not**
*	require the key to be present in the filter.
*
*	Example usage:
*       filter := bson.M{
*			"key": "someKey",
*		}
*
*       OR
*
*		filter := bson.M{
*			"$or": []any{
*				bson.M{"username": "someKey"},
*				bson.M{"email": "someValue"},
*			},
*			"value.status": "active",
*		}
*		result, found := ms.Lookup("myBucket", filter)
*
**************************************************************************************/
func (ms *MongoStore) Lookup(bucket string, filter any) (map[string]any, bool) {
	col := ms.connectOrCreatBucket(bucket)

	mongoFilter := cleanAndPrefix(filter)
	log.Printf("Lookup: using mongoFilter: %v", mongoFilter)

	if len(mongoFilter) == 0 {
		// nothing allowed to search on
		return nil, false
	}

	var rawDoc map[string]any
	err := col.FindOne(context.Background(), mongoFilter).Decode(&rawDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, false
		}
		log.Printf("Lookup error: %v", err)
		return nil, false
	}

	// unwrap and return
	userMap, ok := rawDoc["value"].(map[string]any)
	if !ok {
		log.Printf("Lookup: unexpected document shape %T", rawDoc["value"])
		return nil, false
	}
	return userMap, true
}

func (ms *MongoStore) List(bucket string) ([]any, error) {
	log.Println("List: all keys in bucket:", bucket)
	collection := ms.connectOrCreatBucket(bucket)

	cursor, err := collection.Find(context.Background(), map[string]any{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	var results []any
	for cursor.Next(context.Background()) {
		var result map[string]any
		if err := cursor.Decode(&result); err != nil {
			return nil, err
		}
		results = append(results, result["value"])
	}

	return results, nil
}

func cleanForbidden(clean []string, filter any) bson.M {
	fm, ok := filter.(bson.M)
	if !ok {
		return bson.M{}
	}
	cleaned := make(bson.M, len(fm))
	for key, val := range fm {
		if slices.Contains(clean, key) {
			log.Printf("Clean: Skipping forbidden key: %s", key)
			continue
		}
		cleaned[key] = val
	}
	return cleaned
}

// cleanAndPrefix takes your incoming filter (bson.M), drops any
// key not in the whitelist, and outputs a bson.M where each kept
// key is prefixed with "value." in one pass.
func cleanAndPrefix(filter any) bson.M {
	fm, ok := filter.(bson.M)
	if !ok {
		return bson.M{}
	}

	out := bson.M{}
	for key, val := range fm {
		if slices.Contains(allowed, key) {
			prefixed := "value." + key
			log.Printf("cleanAndPrefix: allowing %q → %q", key, prefixed)
			out[prefixed] = val
		}
	}
	return out
}

func (ms *MongoStore) connectOrCreatBucket(bucket string) *mongo.Collection {
	log.Printf("Connect: to bucket: %s", bucket)
	collection, exists := ms.buckets[bucket]
	if !exists || collection == nil {
		log.Println("Connect: creating Bucket:", bucket)
		collection = ms.db.Collection(bucket)
		ms.buckets[bucket] = collection
	}
	log.Printf("Connect: using collection: %s", collection.Name())
	return collection
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
		buckets: make(map[string]*mongo.Collection),
	}, nil
}
