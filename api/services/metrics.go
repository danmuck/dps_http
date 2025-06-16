package services

import (
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/danmuck/dps_http/api/logs"
	"github.com/danmuck/dps_http/storage"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// func logs.Log(format string, v ...any) {
// 	if strings.Contains(format, "api:users:metrics") {
// 		log.Printf(format, v...)
// 	}
// }

const delay = 30 * time.Second

// UserMetricsService is a service that provides user metrics.
type UserMetricsService struct {
	version         string
	endpoint        string
	bucket          string
	total_users     int64
	users_over_time map[string]int64
	total_roles     map[string]int64
	running         bool
	buckets         []storage.Bucket

	mu sync.Mutex
}

func NewUserMetricsService(store ...storage.Bucket) *UserMetricsService {
	return &UserMetricsService{
		version:     "1.0.0",
		endpoint:    "/metrics",
		bucket:      "users",
		total_users: 0,
		running:     true,
		buckets:     store,

		users_over_time: make(map[string]int64),
		total_roles:     make(map[string]int64),
	}
}

func (svc *UserMetricsService) Register(router *gin.Engine) {
	logs.Log("[api:users:metrics] Registering UserMetricsService at %s", svc.endpoint)
	rg := router.Group(svc.endpoint)
	rg.GET("/users", MetricsHandler(svc))
}

// Point represents a single data point in the time series for user metrics.
// it matches the format of the users_over_time map and the typescript type
type Point struct {
	Timestamp string `json:"timestamp"`
	Count     int64  `json:"count"`
}

// DataMapToPoints sorts a map of timestamps to counts into a slice of Points.
func DataMapToPoints(m map[string]int64) []Point {
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
		User Growth Data: %v
		Buckets: %v
	`,
		svc.version, svc.endpoint, svc.bucket, svc.total_users, svc.total_roles, len(svc.users_over_time), svc.buckets)
}

func (svc *UserMetricsService) Stop() {
	svc.mu.Lock()
	svc.running = false
	svc.mu.Unlock()
	logs.Log("[api:users:metrics] stopped")
}

func (svc *UserMetricsService) Start() {
	logs.Log("[api:users:metrics] starting ...")
	total_users, err := svc.GetBucket("users").Count()
	if err != nil {
		logs.Log("[api:users:metrics] failed to retrieve user count: %v", err)
		return
	}
	logs.Log("[api:users:metrics] found %d users", total_users)
	svc.total_users = total_users
	users_over_time, err := svc.GetBucket("metrics").ListItems()
	if err != nil {
		logs.Log("[api:users:metrics] failed to retrieve user metrics: %v", err)
		return
	}
	logs.Log("[api:users:metrics] found %d user metrics points", len(users_over_time))
	total_over_time := make(map[string]int64)
	for idx, raw := range users_over_time {
		if idx%500 == 0 {
			logs.Log("[api:users:metrics] processing point %d/%d : %v", idx, len(users_over_time), raw)
		}
		timestamp, ok := raw["key"].(string)
		if !ok {
			logs.Log("[api:users:metrics] found malformed user metrics point without timestamp: %v", raw)
			continue
		}
		count, ok := raw["value"].(int64)
		if !ok {
			logs.Log("[api:users:metrics] found malformed user metrics point without count: %v", raw)
			continue
		}
		total_over_time[timestamp] = count
		if idx%1000 == 0 {
			logs.Log("[api:users:metrics] processing point %d/%d { %s : %d }", idx, len(users_over_time), timestamp, count)
		}
	}
	logs.Log("[api:users:metrics] found %d user metrics points over time", len(total_over_time))
	total_roles, err := CountUsersByRole(svc.GetBucket("users"), svc.bucket)
	if err != nil {
		logs.Log("[api:users:metrics] failed to retrieve user roles: %v", err)
		return
	}
	svc.mu.Lock()
	svc.running = true
	svc.users_over_time = total_over_time
	svc.total_users = total_users
	svc.total_roles = total_roles
	svc.mu.Unlock()

	logs.Log("[api:users:metrics] initialized with %d users, roles: %v, users_over_time points: %v",
		total_users, total_roles, len(total_over_time))

	go serviceHandler(svc)
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
func (svc *UserMetricsService) GetBucket(name string) storage.Bucket {
	svc.mu.Lock()
	defer svc.mu.Unlock()
	for _, b := range svc.buckets {
		if b.Name() == name {
			return b
		}
	}
	return nil
}

func serviceHandler(svc *UserMetricsService) {
	logs.Log("[api:users:metrics] is running: %v", svc.IsRunning())
	if !svc.IsRunning() {
		logs.Log("[api:users:metrics] is not running, exiting handler")
		return
	}

	logs.Log("[api:users:metrics] version: %s", svc.GetVersion())
	logs.Log("[api:users:metrics] endpoint: %s", svc.GetEndpoint())
	logs.Log("[api:users:metrics] bucket: %s", svc.GetBucket("users"))
	logs.Log("[api:users:metrics] cycle: %s", delay.String())
	defer svc.Stop() // ensure we stop the service when done

	for svc.IsRunning() {
		logs.Log("[api:users:metrics] is running, processing metrics...")
		total_users, err := svc.GetBucket("users").Count()
		if err != nil {
			logs.Log("[api:users:metrics] failed to retrieve user count: %v", err)
			return
		}

		roles, err := CountUsersByRole(svc.GetBucket("users"), svc.bucket)
		if err != nil {
			logs.Log("[api:users:metrics] failed to retrieve user roles: %v", err)
			svc.mu.Unlock()
			return
		}
		for k, v := range roles {
			fmt.Printf("{ %s -- %d }\n", k, v)
		}
		svc.mu.Lock()
		svc.users_over_time[time.Now().Format(time.Stamp)] = total_users
		svc.total_roles = roles
		svc.total_users = total_users
		svc.mu.Unlock()
		logs.Log("[api:users:metrics] wrote to memory: %d", total_users)

		collection := svc.GetBucket("metrics")
		logs.Log("[api:users:metrics] storing user metrics @%v", collection.Name())
		for timestamp, users := range svc.users_over_time {
			go func(ts string, us int64) {
				if err := collection.Store(ts, us); err != nil {
					logs.Log("[api:users:metrics] failed to store user metrics: %v", err)
					return
				}
			}(timestamp, users)
		}

		time.Sleep(delay) // wait for the next cycle
		logs.Log("[api:users:metrics] processed metrics")
	}
	logs.Log("[api:users:metrics] handler exiting, service is no longer running")
}

// MetricsHandler returns a gin.HandlerFunc that retrieves user metrics.
// It counts total users, roles, and tracks user counts over time.
// This handler is intended to be used with the /metrics endpoint.
// It initializes the UserMetricsService and returns the metrics in JSON format.
// It also handles errors gracefully and logs relevant information.
func MetricsHandler(svc *UserMetricsService) gin.HandlerFunc {
	// note: this is a singleton service, so we can use a single instance
	// it needs to be initialized at the server
	logs.Log("[api:users:metrics] initialized empty service: %+v", svc)
	return func(c *gin.Context) {
		svc.mu.Lock()
		total_users := svc.total_users
		total_roles := svc.total_roles
		users_over_time_points := DataMapToPoints(svc.users_over_time)
		logs.Log("[api:users:metrics] total users: %d, total roles: %v", total_users, total_roles)
		if total_users == 0 {
			logs.Log("[api:users:metrics] no users found, returning empty metrics")
			c.JSON(http.StatusOK, gin.H{
				"total_users":     0,
				"total_roles":     make(map[string]int64),
				"users_over_time": []Point{},
				"message":         "no users found",
			})
			svc.mu.Unlock()
			return
		}
		svc.mu.Unlock()

		// logs.Log("[api:users:metrics] Users over time: %v", svc.users_over_time)
		logs.Log("[api:users:metrics] Service: %+v", svc)

		c.JSON(http.StatusOK, gin.H{
			"total_users":     total_users,
			"total_roles":     total_roles,
			"users_over_time": users_over_time_points,
			"message":         "user metrics retrieved successfully",
		})
	}
}

func CountUsersByRole(store storage.Bucket, bucket string) (map[string]int64, error) {
	logs.Log("[api:users:metrics] counting users by role in bucket: %s", bucket)
	roleCounts := make(map[string]int64)

	// list all users from the specified bucket
	// note: should probably point this to the users bucket, this allows access to any bucket
	users, err := store.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	logs.Log("[api:users:metrics] found %d users in bucket: %s", len(users), bucket)
	// count each role across all users
	for _, raw := range users {
		user, ok := raw.(map[string]any)
		if !ok {
			logs.Log("skipping malformed user record: %v", raw)
			continue
		}
		rolesRaw, ok := user["roles"]
		if !ok {
			logs.Log("skipping user with no roles field: %v", user)
			continue
		}
		roles, ok := rolesRaw.(primitive.A)
		if !ok {
			logs.Log("skipping user with non-list roles: %v %T", rolesRaw, roles)
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
