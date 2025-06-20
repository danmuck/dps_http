package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/danmuck/dps_http/api/logs"
	"github.com/danmuck/dps_http/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoBucket is a wrapper around a MongoDB collection for storing key-value pairs.
type mongoBucket struct {
	id    string
	count int64 // for internal use, future use
	size  int64 // for internal use, future use
	*mongo.Collection
}

// newMongoBucket creates an instance using the MongoDB collection to the given id
// bucket name is aliased to id for future
func newMongoBucket(db *mongo.Database, id string) *mongoBucket {
	logs.Init("NewMongoBucket %q", id)
	collection := db.Collection(id)
	return &mongoBucket{
		id:         id,
		count:      0, // not in use
		size:       0, // not in use
		Collection: collection,
	}
}

// Name returns the bucket id
func (b *mongoBucket) Name() string {
	return b.id
}

// String returns a string representation of the bucket
// note: this will change in the future therefore its structure
// is not reliable
func (b *mongoBucket) String() string {
	return fmt.Sprintf("MongoBucket(id=%q, count=%d, size=%d)", b.id, b.count, b.size)
}

// Store stores the given key-value pair in the bucket
func (b *mongoBucket) Store(key string, value any) error {
	_, err := b.UpdateOne(
		context.Background(),
		map[string]any{
			"key":   key,
			"value": value,
		},
		bson.M{"$set": map[string]any{
			"key":   key,
			"value": value,
		}}, options.Update().SetUpsert(true),
	)
	if err != nil {
		logs.Err("Store() : %v", err)
		return err
	}
	return nil
}

// Retrieve retrieves the value associated with the given key from the bucket
// @TODO kinda bypassing this currently
func (b *mongoBucket) Retrieve(key string) (any, error) {
	logs.Dev("Retrieve not implemented for MongoBucket")
	return nil, fmt.Errorf("Retrieve not applicable for a bucket")
}

// Delete deletes the key-value pair associated with the given key from the bucket
func (b *mongoBucket) Delete(key string) error {
	logs.Dev("Delete [%s] key=%q", b.Name(), key)
	_, err := b.DeleteOne(context.Background(), bson.M{"key": key})
	return err
}

// Update simply replaces the value for a key
// note: this is due to dps_storage integration down the road
func (b *mongoBucket) Update(key string, value any) error {
	filter := bson.M{"key": key}
	update := bson.M{"$set": bson.M{"value": value}}

	result, err := b.UpdateOne(context.Background(), filter, update)
	if err != nil {
		logs.Err("Update() : %v", err)
		return err
	}
	if result.MatchedCount == 0 {
		logs.Err("no document matched for key %q in bucket %q", key, b.Name())
		return fmt.Errorf("no document with key=%q in bucket=%q", key, b.Name())
	}
	// (Optionally, you can also check result.ModifiedCount==0 to warn if the value
	// was identical and thus not modified.)

	return nil
}

// Patch the value for a key
// this prepends the field name with "value." so that it complies with the
// top level document structure {key: value} in which "username" is a field of value
// as well as the MongoDB schema
func (b *mongoBucket) Patch(key string, updates map[string]any) error {
	logs.Init("Patch [%q] { %q : %v }", b.Name(), key, updates)

	// Build a $set document that prefixes each field with "value."
	patch := bson.M{}
	for field, val := range updates {
		patch[storage.Prefix(field)] = val
	}
	logs.Log("patch bson : %v", patch)

	result, err := b.UpdateOne(
		context.Background(),
		bson.M{"key": key},
		bson.M{"$set": patch},
	)
	if err != nil {
		logs.Err("Patch() : %v", err)
		return err
	}
	if result.MatchedCount == 0 {
		logs.Err("no document matched for key %q in bucket %q", key, b.Name())
		return fmt.Errorf("no document with key=%q in bucket=%q", key, b.Name())
	}
	logs.Log("[%q] patched %d documents for key %q",
		b.Name(), result.ModifiedCount, key)
	return nil
}

// Lookup retrieves a document from the bucket based on the provided filter.
// The filter is cleaned according to config @REMINDER currently in utils.go
// and prefixed to ensure compliance with the MongoDB schema. (storage/utils.go)
func (b *mongoBucket) Lookup(filter any) (map[string]any, bool) {
	mongoFilter := storage.CleanAndPrefix(filter)
	logs.Init("Lookup filter : %v", mongoFilter)

	if len(mongoFilter) == 0 {
		// nothing allowed to search on
		logs.Err("empty filter provided, nothing to search on")
		return nil, false
	}

	var rawDoc map[string]any
	err := b.FindOne(context.Background(), mongoFilter).Decode(&rawDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			logs.Warn("no document found for filter %v", mongoFilter)
			return nil, false
		}
		logs.Err("Lookup error: %v", err)
		return nil, false
	}

	// unwrap and return
	userMap, ok := rawDoc["value"].(map[string]any)
	if !ok {
		logs.Err("unexpected document shape %T", rawDoc["value"])
		return nil, false
	}
	logs.Log("found user map: %v", userMap["username"])
	return userMap, true
}

// ListKeys retrieves all keys from the bucket
func (b *mongoBucket) ListKeys() ([]any, error) {
	logs.Init("List [%s]", b.Name())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := b.Find(ctx, map[string]any{})
	if err != nil {
		logs.Err("error: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []any
	for cursor.Next(ctx) {
		var result map[string]any
		if err := cursor.Decode(&result); err != nil {
			logs.Err("decode error: %v", err)
			return nil, err
		}
		results = append(results, result["value"])
	}

	return results, nil
}

// ListItems retrieves all items from the bucket
// note: you will need to validate the result before using it
func (b *mongoBucket) ListItems() ([]map[string]any, error) {
	logs.Init("ListItems [%s]", b.Name())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := b.Find(ctx, map[string]any{})
	if err != nil {
		logs.Err("error: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []map[string]any
	for cursor.Next(ctx) {
		var result map[string]any
		if err := cursor.Decode(&result); err != nil {
			logs.Err("decode error: %v", err)
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

// Get the number of keys in the bucket
func (b *mongoBucket) Count() (int64, error) {
	logs.Init("Count [%s]", b.Name())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := b.CountDocuments(ctx, bson.M{})
	logs.Log("[%s] count: %d", b.Name(), count)
	return count, err
}
