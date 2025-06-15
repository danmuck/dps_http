package storage

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// todo: need to move this to mongo.go
func (ms *MongoStore) Count(bucket string) (int64, error) {
	log.Printf("Counting documents in bucket: %s", bucket)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := ms.db.Collection(bucket).CountDocuments(ctx, bson.M{})
	log.Printf("Counted %d documents in bucket: %s", count, bucket)
	return count, err
}
