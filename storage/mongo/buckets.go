package mongo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/danmuck/dps_http/storage"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoBucket struct {
	id    string
	count int64 // for internal use, future use
	size  int64 // for internal use, future use
	*mongo.Collection
}

func NewMongoBucket(db *mongo.Database, id string) *MongoBucket {
	log.Printf("Creating MongoBucket with id=%q", id)
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

func (b *MongoBucket) Store(key string, value any) error {
	_, err := b.InsertOne(context.Background(), map[string]any{
		"key":   key,
		"value": value,
	})
	if err != nil {
		log.Printf("Store error: %v", err)
		return err
	}
	return nil
}
func (b *MongoBucket) Retrieve(key string) (any, error) {
	return nil, fmt.Errorf("Retrieve not applicable for a bucket")
}
func (b *MongoBucket) Delete(key string) error {
	_, err := b.DeleteOne(context.Background(), map[string]any{"key": key})
	return err
}
func (b *MongoBucket) Update(key string, value any) error {
	filter := bson.M{"key": key}
	update := bson.M{"$set": bson.M{"value": value}}

	result, err := b.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return fmt.Errorf("no document with key=%q in bucket=%q", key, b.Name())
	}
	// (Optionally, you can also check result.ModifiedCount==0 to warn if the value
	// was identical and thus not modified.)

	return nil
}
func (b *MongoBucket) Patch(key string, updates map[string]any) error {
	log.Printf("Patching key %q in bucket %q with updates: %v", key, b.Name(), updates)

	// Build a $set document that prefixes each field with "value."
	setDoc := bson.M{}
	for field, val := range updates {
		setDoc["value."+field] = val
	}
	log.Printf("Patch: using setDoc: %v", setDoc)
	// Run the update
	result, err := b.UpdateOne(
		context.Background(),
		bson.M{"key": key},
		bson.M{"$set": setDoc},
	)
	if err != nil {
		log.Printf("Patch error: %v", err)
		return err
	}
	if result.MatchedCount == 0 {
		log.Printf("Patch: no document matched for key %q in bucket %q", key, b.Name())
		return fmt.Errorf("no document with key=%q in bucket=%q", key, b.Name())
	}
	log.Printf("Patch: updated %d document(s) for key %q in bucket %q", result.ModifiedCount, key, b.Name())
	return nil
}

func (b *MongoBucket) Lookup(filter any) (map[string]any, bool) {
	mongoFilter := storage.CleanAndPrefix(filter)
	log.Printf("Lookup: using mongoFilter: %v", mongoFilter)

	if len(mongoFilter) == 0 {
		// nothing allowed to search on
		return nil, false
	}

	var rawDoc map[string]any
	err := b.FindOne(context.Background(), mongoFilter).Decode(&rawDoc)
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
func (b *MongoBucket) List(bucket string) ([]any, error) {
	log.Println("List: all keys in bucket:", bucket)

	cursor, err := b.Find(context.Background(), map[string]any{})
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

func (b *MongoBucket) Count() (int64, error) {
	log.Printf("Counting documents in bucket: %s", b.Name())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := b.CountDocuments(ctx, bson.M{})
	log.Printf("Counted %d documents in bucket: %s", count, b.Name())
	return count, err
}
