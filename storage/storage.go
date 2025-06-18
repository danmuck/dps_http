package storage

import (
	"context"
)

//	var forbidden []string = []string{
//		"password_hash", "token", "created_at", "updated_at",
//		"bio", "avatar_url", "email",
//	}
var allowed []string = []string{
	"username", "roles",
}

type Bucket interface {
	Name() string                                   // returns the bucket name
	Store(key string, value any) error              // stores a value by key
	Retrieve(key string) (any, error)               // retrieves a value by key
	Delete(key string) error                        // deletes a value by key
	Update(key string, value any) error             // updates a value by key
	Patch(key string, updates map[string]any) error // updates specific fields in a document
	Lookup(key any) (map[string]any, bool)          // looks up a specific key in the bucket
	ListKeys() ([]any, error)                       // lists all keys in the bucket
	ListItems() ([]map[string]any, error)           // lists all items in the bucket
	Count() (int64, error)                          // counts documents in the bucket
}

type Client interface {
	Name() string                   // returns the client name
	Location() string               // returns the client location
	Type() string                   // returns the client type
	Ping(ctx context.Context) error // checks if the client is reachable

	ConnectOrCreateBucket(bucket string) Bucket             // connects to or creates a bucket
	Store(bucket string, key string, value any) error       // stores a value in a bucket by key
	Retrieve(bucket string, key string) (any, error)        // retrieves a value from a bucket by key
	Delete(bucket string, key string) error                 // deletes a value from a bucket by key
	Update(bucket string, key string, value any) error      // updates a value in a bucket by key
	Patch(bucket, key string, updates map[string]any) error // updates specific fields in a document
	List(bucket string) ([]any, error)                      // lists all keys in a bucket
	Lookup(bucket string, key any) (map[string]any, bool)   // looks up a specific key in a bucket
	Count(bucket string) (int64, error)                     // counts documents in a bucket
}
