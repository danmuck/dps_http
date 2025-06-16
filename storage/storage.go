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
	Name() string
	Store(key string, value any) error
	Retrieve(key string) (any, error)
	Delete(key string) error
	Update(key string, value any) error
	Patch(key string, updates map[string]any) error // Patch updates specific fields in a document
	Lookup(key any) (map[string]any, bool)          // Lookup a specific key in a bucket
	List() ([]any, error)                           // List all keys in a bucket
	ListItems() ([]map[string]any, error)           // List all items in a bucket
	Count() (int64, error)                          // Count documents in a bucket
}

type Client interface {
	Name() string
	Location() string
	Type() string
	Ping(ctx context.Context) error

	ConnectOrCreateBucket(bucket string) Bucket // Connect to an existing bucket or create a new one if it doesn't exist
	// bucket operations
	Store(bucket string, key string, value any) error
	Retrieve(bucket string, key string) (any, error)
	Delete(bucket string, key string) error
	Update(bucket string, key string, value any) error
	Patch(bucket, key string, updates map[string]any) error // Patch updates specific fields in a document
	List(bucket string) ([]any, error)                      // List all keys in a bucket``
	Lookup(bucket string, key any) (map[string]any, bool)   // Lookup a specific key in a bucket
	Count(bucket string) (int64, error)                     // Count documents in a bucket
}
