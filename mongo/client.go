package mongo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/danmuck/dps_lib/logs"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Wrapper for MongoDB operations
//
// server initializes a MongoManager
// registers buckets for each service w/ permissions
// services strictly have access to *Bucket.Mongo()
// which returns a *mongo.Collection
type MongoManager struct {
	name     string
	client   *mongo.Client
	appdb    *mongo.Database
	registry map[string]*Bucket // map[endpoint]*Bucket

	mu sync.Mutex
}
type Bucket struct {
	endpoint string
	coll     *mongo.Collection
	public   bool
}

// this is the only gateway for a service to interact with MongoDB
//
// returns the inner MongoDB collection
func (b *Bucket) Mongo() *mongo.Collection {
	return b.coll
}

func (mm *MongoManager) RegisterBucket(endpoint string, public bool) *Bucket {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mm.registry == nil {
		// if there are no buckets yet, initialize the map
		mm.registry = make(map[string]*Bucket)
	}
	if _, exists := mm.registry[endpoint]; exists {
		// Bucket already exists, return nil
		return nil
	}
	// Bucket does not exist, register the endpoint and create a new bucket
	mm.registry[endpoint] = &Bucket{
		endpoint: endpoint,
		coll:     mm.appdb.Collection(endpoint),
		public:   public,
	}
	return mm.registry[endpoint]
}

func (mm *MongoManager) Bucket(endpoint string) (*Bucket, error) {
	mm.mu.Lock()
	defer mm.mu.Unlock()

	if mm.registry == nil {
		return nil, fmt.Errorf("no buckets registered")
	}
	bucket, exists := mm.registry[endpoint]
	if !exists {
		return nil, fmt.Errorf("bucket %s does not exist", endpoint)
	}
	if !bucket.public {
		return nil, fmt.Errorf("bucket %s is private", endpoint)
	}
	return bucket, nil
}

func NewMongoManager(uri, dbName string) (*MongoManager, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		logs.Log("failed to connect to MongoDB at %s: %v", uri, err)
		return nil, err
	}
	logs.Log("connecting to MongoDB at %s", uri)
	db := client.Database(dbName)
	return &MongoManager{
		name:   dbName,
		client: client,
		appdb:  db,
	}, nil
}

// Name returns the name of the MongoDB Database
func (mm *MongoManager) Name() string {
	return mm.name
}

// MongoDB client Ping wrapper
func (mm *MongoManager) Ping(ctx context.Context) error {
	return mm.client.Ping(ctx, nil)
}
