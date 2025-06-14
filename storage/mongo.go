package storage

import (
	"context"
	"fmt"
	"log"
	"slices"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// var forbidden []string = []string{
// 	"password_hash", "token", "created_at", "updated_at",
// 	"bio", "avatar_url", "email",
// }

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

// //////////////////////////////////////////////////////////////
// [Helpers]
// //////////////////////////////////////////////////////////////

// connectOrCreatBucket connects to an existing bucket or creates a new one if it doesn't exist.
// It returns the collection for the specified bucket.
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

func cleanAndPrefix(filter any) bson.M {
	fm, ok := filter.(bson.M)
	if !ok {
		return bson.M{}
	}

	out := bson.M{}
	for key, val := range fm {
		if key == "key" {
			// special case for "key" to avoid prefixing
			log.Printf("cleanAndPrefix: allowing key %q without prefix", key)
			out[key] = val
			continue
		}
		if slices.Contains(allowed, key) {
			prefixed := "value." + key
			log.Printf("cleanAndPrefix: allowing %q → %q", key, prefixed)
			out[prefixed] = val
		}
	}
	return out
}

// basic CRUD operations
func (ms *MongoStore) Store(bucket string, key string, value any) error {
	log.Printf("Storing key %q in bucket %q", key, bucket)
	collection := ms.connectOrCreatBucket(bucket)

	_, err := collection.InsertOne(context.Background(), map[string]any{
		"key":   key,
		"value": value,
	})
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
	collection := ms.connectOrCreatBucket(bucket)

	_, err := collection.DeleteOne(context.Background(), map[string]any{"key": key})
	return err
}

// TODO: implement
func (ms *MongoStore) Update(bucket string, key string, value any) error {
	col := ms.connectOrCreatBucket(bucket)
	log.Printf("Updating key %q in bucket %q with value: %v", key, bucket, value)
	// 1) Explicit BSON filter on the top‐level "key" field
	filter := bson.M{"key": key}
	// 2) Explicit BSON update of the nested "value"
	update := bson.M{"$set": bson.M{"value": value}}

	result, err := col.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	// 3) If no document matched, surface an error
	if result.MatchedCount == 0 {
		return fmt.Errorf("no document with key=%q in bucket=%q", key, bucket)
	}
	// (Optionally, you can also check result.ModifiedCount==0 to warn if the value
	// was identical and thus not modified.)

	return nil
}
func (ms *MongoStore) Patch(bucket, key string, updates map[string]any) error {
	log.Printf("Patching key %q in bucket %q with updates: %v", key, bucket, updates)
	col := ms.connectOrCreatBucket(bucket)

	// Build a $set document that prefixes each field with "value."
	setDoc := bson.M{}
	for field, val := range updates {
		setDoc["value."+field] = val
	}
	log.Printf("Patch: using setDoc: %v", setDoc)
	// Run the update
	result, err := col.UpdateOne(
		context.Background(),
		bson.M{"key": key},
		bson.M{"$set": setDoc},
	)
	if err != nil {
		log.Printf("Patch error: %v", err)
		return err
	}
	if result.MatchedCount == 0 {
		log.Printf("Patch: no document matched for key %q in bucket %q", key, bucket)
		return fmt.Errorf("no document with key=%q in bucket=%q", key, bucket)
	}
	log.Printf("Patch: updated %d document(s) for key %q in bucket %q", result.ModifiedCount, key, bucket)
	return nil
}

//		Lookup retrieves a specific key in a bucket with a filter
//
//		The filter is expected to be a map-like structure (e.g., bson.M) that can be used to build a MongoDB query.
//		It returns the first matching document as a map[string]any and a boolean indicating if it was found.
//		If no document matches the filter, it returns nil and false.
//
//		The filter can include MongoDB operators like $or, $and, etc., and will be prefixed with "value." for fields.
//
//		Note: This function iterates through the filter to build a complete MongoDB query and it does **not**
//		require the key to be present in the filter.
//
//		Example usage:
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
			log.Printf("Lookup: no document found for filter %v", mongoFilter)
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
	log.Printf("Lookup: found user map: %v", userMap["username"])
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
