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

	// Basic CRUD operations
	Store(bucket string, key string, value any) error
	Retrieve(bucket string, key string) (any, error)
	Delete(bucket string, key string) error
	Update(bucket string, key string, value any) error
	List(bucket string) ([]any, error)                    // List all keys in a bucket``
	Lookup(bucket string, key any) (map[string]any, bool) // Lookup a specific key in a bucket
}
