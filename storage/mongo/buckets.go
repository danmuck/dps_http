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

type MongoBucket struct {
	id    string
	count int64 // for internal use, future use
	size  int64 // for internal use, future use
	*mongo.Collection
}

func NewMongoBucket(db *mongo.Database, id string) *MongoBucket {
	logs.Init("NewMongoBucket %q", id)
	collection := db.Collection(id)
	return &MongoBucket{
		id:         id,
		count:      0, // not in use
		size:       0, // not in use
		Collection: collection,
	}
}

func (b *MongoBucket) Name() string {
	return b.id
}
func (b *MongoBucket) String() string {
	return fmt.Sprintf("MongoBucket(id=%q, count=%d, size=%d)", b.id, b.count, b.size)
}

func (b *MongoBucket) Store(key string, value any) error {
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
func (b *MongoBucket) Retrieve(key string) (any, error) {
	logs.Dev("Retrieve not implemented for MongoBucket")
	return nil, fmt.Errorf("Retrieve not applicable for a bucket")
}
func (b *MongoBucket) Delete(key string) error {
	logs.Dev("Delete [%s] key=%q", b.Name(), key)
	_, err := b.DeleteOne(context.Background(), map[string]any{"key": key})
	return err
}
func (b *MongoBucket) Update(key string, value any) error {
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
func (b *MongoBucket) Patch(key string, updates map[string]any) error {
	logs.Init("Patch [%q] { %q : %v }", b.Name(), key, updates)

	// Build a $set document that prefixes each field with "value."
	patch := bson.M{}
	for field, val := range updates {
		patch["value."+field] = val
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

func (b *MongoBucket) Lookup(filter any) (map[string]any, bool) {
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
func (b *MongoBucket) ListKeys() ([]any, error) {
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

func (b *MongoBucket) ListItems() ([]map[string]any, error) {
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

func (b *MongoBucket) Count() (int64, error) {
	logs.Init("Count [%s]", b.Name())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := b.CountDocuments(ctx, bson.M{})
	logs.Log("[%s] count: %d", b.Name(), count)
	return count, err
}
