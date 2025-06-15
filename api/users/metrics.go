package users

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/danmuck/dps_http/storage"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const delay = 60 * time.Second

// UserMetricsService is a service that provides user metrics.
type UserMetricsService struct {
	version         string
	endpoint        string
	store           storage.Storage
	bucket          string
	total_users     int64
	users_over_time map[string]int64
	total_roles     map[string]int64
	running         bool
	b               storage.Bucket

	mu sync.Mutex
}

func NewUserMetricsService(store storage.Storage) *UserMetricsService {
	return &UserMetricsService{
		version:         "1.0.0",
		endpoint:        "/metrics",
		store:           store,
		bucket:          "users",
		total_users:     0,
		users_over_time: make(map[string]int64),
		total_roles:     make(map[string]int64),
		running:         true,
	}
}

func (svc *UserMetricsService) Register(router *gin.Engine) {
	log.Printf("[api:users] Registering UserMetricsService at %s", svc.endpoint)
	rg := router.Group(svc.endpoint)
	rg.GET(svc.endpoint, MetricsHandler(svc.store))
}

// Point represents a single data point in the time series for user metrics.
// it matches the format of the users_over_time map and the typescript type
type Point struct {
	Timestamp string `json:"timestamp"`
	Count     int64  `json:"count"`
}

// SortMapByKey sorts a map of timestamps to counts into a slice of Points.
// it assumes the timestamps are in ISO8601 format like "2024-06-15T15:04:05Z07:00"
func SortMapByKey(m map[string]int64) []Point {
	points := make([]Point, 0, len(m))
	for k, v := range m {
		points = append(points, Point{
			Timestamp: k,
			Count:     v,
		})
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].Timestamp < points[j].Timestamp
	})

	return points
}

func (svc *UserMetricsService) String() string {
	return fmt.Sprintf(`
	UserMetricsService:
		Version: %s
		Endpoint: %s
		Bucket: %s
		Total Users: %d
		Total Roles: %v
		Roles Over Time: %v
	`,
		svc.version, svc.endpoint, svc.bucket, svc.total_users, svc.total_roles, svc.users_over_time)
}

func (svc *UserMetricsService) Stop() {
	svc.mu.Lock()
	svc.running = false
	svc.mu.Unlock()
	log.Printf("[api:users] UserMetricsService stopped")
}
func (svc *UserMetricsService) Start() {
	svc.mu.Lock()
	svc.running = true
	svc.mu.Unlock()
	go serviceHandler(svc)
	log.Printf("[api:users] UserMetricsService started")
}
func (svc *UserMetricsService) IsRunning() bool {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	return svc.running
}
func (svc *UserMetricsService) GetVersion() string {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	return svc.version
}
func (svc *UserMetricsService) GetEndpoint() string {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	return svc.endpoint
}
func (svc *UserMetricsService) GetBucket() string {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	return svc.bucket
}

func serviceHandler(svc *UserMetricsService) {
	log.Printf("[api:users] UserMetricsService is running: %v", svc.IsRunning())
	if !svc.IsRunning() {
		log.Printf("[api:users] UserMetricsService is not running, exiting handler")
		return
	}

	log.Printf("[api:users] UserMetricsService version: %s", svc.GetVersion())
	log.Printf("[api:users] UserMetricsService endpoint: %s", svc.GetEndpoint())
	log.Printf("[api:users] UserMetricsService bucket: %s", svc.GetBucket())

	for svc.IsRunning() {
		svc.mu.Lock()
		defer svc.mu.Unlock()
		log.Printf("[api:users] UserMetricsService is running, processing metrics...")
		// Here you would typically gather metrics, update the store, etc.
		total_users, err := svc.store.Count(svc.bucket)
		if err != nil {
			log.Printf("[api:users] UserMetricsService failed to retrieve user count: %v", err)
			return
		}
		svc.users_over_time[time.Now().Format(time.Stamp)] = total_users
		time.Sleep(delay) // simulate work
		log.Printf("[api:users] UserMetricsService processed metrics")
	}
}

// MetricsHandler returns a gin.HandlerFunc that retrieves user metrics.
// It counts total users, roles, and tracks user counts over time.
// This handler is intended to be used with the /metrics endpoint.
// It initializes the UserMetricsService and returns the metrics in JSON format.
// It also handles errors gracefully and logs relevant information.
// Note: This service is a singleton, so it can be initialized once and reused.
// It is designed to be thread-safe using a mutex for concurrent access.
func MetricsHandler(store storage.Storage) gin.HandlerFunc {
	// note: this is a singleton service, so we can use a single instance
	// it needs to be initialized at the server
	svc := &UserMetricsService{
		version:         "1.0.0",
		endpoint:        "/metrics",
		store:           store,
		bucket:          "users",
		total_users:     0,
		users_over_time: make(map[string]int64),
		total_roles:     make(map[string]int64),
		running:         true,
	}

	return func(c *gin.Context) {
		total_users, err := store.Count(svc.bucket)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user count"})
			return
		}
		total_roles, err := CountUsersByRole(svc.store, svc.bucket)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve user roles"})
			return
		}
		for k, v := range total_roles {
			fmt.Printf("{ %s -- %d }\n", k, v)
		}
		svc.total_users = total_users
		svc.total_roles = total_roles
		svc.mu.Lock()
		svc.users_over_time[time.Now().Format(time.Stamp)] = total_users
		svc.mu.Unlock()

		log.Printf("[api:users] Users over time: %v", svc.users_over_time)
		log.Printf("[api:users] Service: %+v", svc)

		returnedUsers := SortMapByKey(svc.users_over_time)
		c.JSON(http.StatusOK, gin.H{
			"total_users":     total_users,
			"total_roles":     total_roles,
			"users_over_time": returnedUsers,
			"message":         "user metrics retrieved successfully",
		})
	}
}

func CountUsersByRole(store storage.Storage, bucket string) (map[string]int64, error) {
	log.Printf("[api:users] Counting users by role in bucket: %s", bucket)
	roleCounts := make(map[string]int64)

	// list all users from the specified bucket
	// note: should probably point this to the users bucket, this allows access to any bucket
	users, err := store.List(bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	log.Printf("[api:users] Found %d users in bucket: %s", len(users), bucket)
	// count each role across all users
	for _, raw := range users {
		user, ok := raw.(map[string]any)
		if !ok {
			log.Printf("Skipping malformed user record: %v", raw)
			continue
		}
		rolesRaw, ok := user["roles"]
		if !ok {
			log.Printf("Skipping user with no roles field: %v", user)
			continue
		}
		roles, ok := rolesRaw.(primitive.A)
		if !ok {
			log.Printf("Skipping user with non-list roles: %v %T", rolesRaw, roles)
			continue
		}

		for _, r := range roles {
			roleStr, ok := r.(string)
			if ok {
				roleCounts[roleStr]++
			}
		}
	}

	return roleCounts, nil
}
