package metrics

import (
	"fmt"
	"sync"
	"time"

	"github.com/danmuck/dps_http/configs"
	"github.com/danmuck/dps_http/lib/logs"
	"github.com/danmuck/dps_http/lib/middleware"
	"github.com/danmuck/dps_http/lib/storage"
	"github.com/danmuck/dps_http/lib/storage/mongo"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var service *UserMetricsService

// UserMetricsService is a service that provides user metrics.
type UserMetricsService struct {
	version  string
	endpoint string

	total_users     int64
	users_over_time map[string]int64
	total_roles     map[string]int64

	running   bool
	userDB    string
	metricsDB string
	storage   storage.Client // mongo client

	mu sync.Mutex
}

//	Service interface implementation
//
// //
func (svc *UserMetricsService) Up(rg *gin.RouterGroup) {
	logs.Init("Register %s", service.metricsDB)
	admin := rg.Group("/metrics")
	admin.Use(middleware.JWTMiddleware(), middleware.AuthorizeByRoles("admin"))
	admin.GET("/users", UserGrowth(svc))
	svc.start()
}

func (svc *UserMetricsService) Down() error {
	svc.mu.Lock()
	svc.running = false
	svc.mu.Unlock()
	logs.Log("stopped")
	return nil
}

func (svc *UserMetricsService) Version() string {
	return svc.version
}

func (svc *UserMetricsService) DependsOn() []string {
	return nil
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
		len(svc.users_over_time), svc.storage.Name())
}

func NewUserMetricsService(endpoint string) *UserMetricsService {
	cfg, err := configs.LoadConfig()
	if err != nil {
		logs.Fatal(err.Error())
	}
	m, err := mongo.NewMongoStore(cfg.DB.MongoURI, cfg.DB.Name)
	if err != nil {
		logs.Log("failed to create mongo store: %v", err)
		return nil
	}
	logs.Dev("initialized mongo store %s from %s", m.Name(), cfg.String())
	version := "v1"

	service = &UserMetricsService{
		endpoint: endpoint,
		version:  version,

		storage:   m,
		userDB:    "users" + version,
		metricsDB: endpoint + version,
		running:   false,

		total_users:     0,
		users_over_time: make(map[string]int64),
		total_roles:     make(map[string]int64),
	}
	return service
}

func (svc *UserMetricsService) start() {
	logs.Init("starting ...")
	total_users, err := svc.storage.ConnectOrCreateBucket(service.userDB).Count()
	if err != nil {
		logs.Err("failed to retrieve user count: %v", err)
		return
	}
	logs.Log("loaded %d users", total_users)
	svc.total_users = total_users
	users_over_time, err := svc.storage.ConnectOrCreateBucket(service.metricsDB).ListItems()
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

	go backgroundService()
}

func (svc *UserMetricsService) UserCountByRole() (map[string]int64, error) {
	logs.Init("UserCountByRole")
	roleCounts := make(map[string]int64)

	store := svc.storage.ConnectOrCreateBucket(service.userDB)
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
func backgroundService() {
	if !service.running {
		logs.Warn("is not running, exiting handler")
		return
	}

	logs.Info("starting %s cycle", configs.METRICS_delay.String())
	defer service.Down()

	for service.running {
		logs.Log("processing metrics...")
		total_users, err := service.storage.ConnectOrCreateBucket(service.userDB).Count()
		if err != nil {
			logs.Err("failed to retrieve user count: %v", err)
			return
		}

		roles, err := service.UserCountByRole()
		if err != nil {
			logs.Err("failed to retrieve user roles: %v", err)
			return
		}
		service.mu.Lock()
		service.users_over_time[time.Now().Format(time.Stamp)] = total_users
		service.total_roles = roles
		service.total_users = total_users
		service.mu.Unlock()

		collection := service.storage.ConnectOrCreateBucket(service.metricsDB)
		for timestamp, users := range service.users_over_time {
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
