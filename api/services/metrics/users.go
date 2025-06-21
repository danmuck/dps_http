package metrics

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/danmuck/dps_http/configs"
	"github.com/danmuck/dps_http/lib/logs"
	"github.com/danmuck/dps_http/storage"
	"github.com/danmuck/dps_http/storage/mongo"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserMetricsService is a service that provides user metrics.
type UserMetricsService struct {
	version         string
	endpoint        string
	tracking        string
	total_users     int64
	users_over_time map[string]int64
	total_roles     map[string]int64
	running         bool
	bucket          storage.Client

	mu sync.Mutex
}

//	Service interface implementation
//
// //
func (svc *UserMetricsService) Up(rg *gin.RouterGroup) {
	logs.Init("Register %s", "metrics")
	rg.GET("/users", MetricsHandler(svc))
	svc.start()
	defer svc.Down()
}

func (svc *UserMetricsService) Down() {
	svc.mu.Lock()
	svc.running = false
	svc.mu.Unlock()
	logs.Log("stopped")
}

func (svc *UserMetricsService) String() string {
	return fmt.Sprintf(`
	UserMetricsService:
		Version: %s

		total users: %d
		growth data: %v
		role counts: %v
		bucket: %v
	`,
		svc.version,
		svc.total_users, svc.total_roles,
		len(svc.users_over_time), svc.bucket)
}

func NewUserMetricsService() *UserMetricsService {
	db, err := mongo.NewMongoStore("metrics", "users")
	if err != nil {
		logs.Err("failed to connect to database: %v", err)
		return nil
	}
	return &UserMetricsService{
		total_users: 0,
		running:     false,
		bucket:      db,

		users_over_time: make(map[string]int64),
		total_roles:     make(map[string]int64),
	}
}

func MetricsHandler(svc *UserMetricsService) gin.HandlerFunc {
	// note: this is a singleton service, so we can use a single instance
	// it needs to be initialized at the server
	logs.Init("initializing service handler [%s]", svc.String())
	return func(c *gin.Context) {
		svc.mu.Lock()
		total_users := svc.total_users
		total_roles := svc.total_roles
		users_over_time_points := MapTimestampToInt64Points(svc.users_over_time)
		logs.Info("total users: %d, total roles: %v", total_users, total_roles)
		if total_users == 0 {
			logs.Warn("no users found, returning empty metrics")
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

		logs.Debug("service: %+v", svc)

		c.JSON(http.StatusOK, gin.H{
			"total_users":     total_users,
			"total_roles":     total_roles,
			"users_over_time": users_over_time_points,
			"message":         "user metrics retrieved successfully",
		})
	}
}

func (svc *UserMetricsService) start() {
	logs.Init("starting ...")
	total_users, err := svc.bucket.ConnectOrCreateBucket("users").Count()
	if err != nil {
		logs.Err("failed to retrieve user count: %v", err)
		return
	}
	logs.Log("loaded %d users", total_users)
	svc.total_users = total_users
	users_over_time, err := svc.bucket.ConnectOrCreateBucket("metrics").ListItems()
	if err != nil {
		logs.Err("failed to retrieve user metrics: %v", err)
		return
	}
	total_over_time := make(map[string]int64)
	for idx, raw := range users_over_time {
		if idx%250 == 1 {
			logs.Debug("sanity check processing point %d/%d", idx, len(users_over_time))
		}
		timestamp, ok := raw["key"].(string)
		if !ok {
			logs.Warn("found malformed user metrics point without timestamp: %v", raw)
			continue
		}
		count, ok := raw["value"].(int64)
		if !ok {
			logs.Warn("found malformed user metrics point without count: %v", raw)
			continue
		}
		total_over_time[timestamp] = count
		if idx%500 == 1 {
			logs.Debug("sanity check processing %d/%d { %s : %d }", idx, len(users_over_time), timestamp, count)
		}
	}
	total_roles, err := svc.UserCountByRole()
	if err != nil {
		logs.Err("failed to retrieve user roles: %v", err)
		return
	}
	svc.mu.Lock()
	svc.running = true
	svc.users_over_time = total_over_time
	svc.total_users = total_users
	svc.total_roles = total_roles
	svc.mu.Unlock()

	logs.Info("initialized with %d users, roles: %v, users_over_time points: %v",
		total_users, total_roles, len(total_over_time))

	go backgroundService(svc)
}

func (svc *UserMetricsService) UserCountByRole() (map[string]int64, error) {
	logs.Init("UserCountByRole")
	roleCounts := make(map[string]int64)

	store := svc.bucket.ConnectOrCreateBucket("users")
	users, err := store.ListKeys()
	if err != nil {
		logs.Err("failed to list users: %v", err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	// count each role across all users
	for _, raw := range users {
		user, ok := raw.(map[string]any)
		if !ok {
			logs.Warn("skipping malformed user record: %v", raw)
			continue
		}
		rolesRaw, ok := user["roles"]
		if !ok {
			logs.Warn("skipping user with no roles field: %v", user)
			continue
		}
		roles, ok := rolesRaw.(primitive.A)
		if !ok {
			logs.Warn("skipping user with non-list roles: %v %T", rolesRaw, roles)
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

//	Private handler for service lifecycle management
//
// //
func backgroundService(svc *UserMetricsService) {
	if !svc.running {
		logs.Warn("is not running, exiting handler")
		return
	}

	logs.Info("starting %s cycle", configs.METRICS_delay.String())
	defer svc.Down()

	for svc.running {
		logs.Log("processing metrics...")
		total_users, err := svc.bucket.ConnectOrCreateBucket("users").Count()
		if err != nil {
			logs.Err("failed to retrieve user count: %v", err)
			return
		}

		roles, err := svc.UserCountByRole()
		if err != nil {
			logs.Err("failed to retrieve user roles: %v", err)
			return
		}
		svc.mu.Lock()
		svc.users_over_time[time.Now().Format(time.Stamp)] = total_users
		svc.total_roles = roles
		svc.total_users = total_users
		svc.mu.Unlock()

		collection := svc.bucket.ConnectOrCreateBucket("metrics")
		for timestamp, users := range svc.users_over_time {
			go func(ts string, us int64) {
				if err := collection.Store(ts, us); err != nil {
					logs.Err("failed to store user metrics: %v", err)
					return
				}
			}(timestamp, users)
		}

		time.Sleep(configs.METRICS_delay) // wait for the next cycle
	}
	logs.Log("handler exiting, service is no longer running")
}
