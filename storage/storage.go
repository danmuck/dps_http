package storage

import (
	"context"
)

type StorageClient interface {
	Connect(ctx context.Context) error
}

type Database interface {
	Name() string
}

type Bucket interface {
	Name() string
}

type Storage interface {
	Name() string
	Type() string
	Client() StorageClient
	Database() Database
	Ping(ctx context.Context) error
	Count(bucket string) (int64, error)         // Count documents in a bucket
	ConnectOrCreateBucket(bucket string) Bucket // Connect to an existing bucket or create a new one if it doesn't exist

	// Basic CRUD operations
	Store(bucket string, key string, value any) error
	Retrieve(bucket string, key string) (any, error)
	Delete(bucket string, key string) error
	Update(bucket string, key string, value any) error
	Patch(bucket, key string, updates map[string]any) error // Patch updates specific fields in a document
	List(bucket string) ([]any, error)                      // List all keys in a bucket``
	Lookup(bucket string, key any) (map[string]any, bool)   // Lookup a specific key in a bucket
}
