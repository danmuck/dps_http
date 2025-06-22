package admin

import (
	"fmt"
	"math/rand"

	"github.com/danmuck/dps_http/configs"
	"github.com/danmuck/dps_http/lib/logs"
	"github.com/danmuck/dps_http/lib/middleware"
	"github.com/danmuck/dps_http/lib/storage/mongo"

	"github.com/gin-gonic/gin"
)

// AdminService implements the service interface
// route handlers are defined in their corresponding files
// (note: there are none for this template service)
// //
var service *AdminService

type AdminService struct {
	endpoint string // endpoint for the service; canonically /api/v_/<service_name>
	version  string // api version; the version this template resides in

	// service specific structures
	userDB  string
	storage *mongo.MongoClient
}

// assign routes for the service and initialize any resources
// routes are structured `api/v1/<service_name>/<your_endpoints>`
func (svc *AdminService) Up(root *gin.RouterGroup) {

	rg := root.Group("/")
	rg.Use(middleware.JWTMiddleware(), middleware.AuthorizeByRoles("admin"))

	ug := rg.Group("/users")
	ug.POST("/:id", CreateUser())
	admin := rg.Group("/admin")
	admin.POST("/gcx", CreateUsersX())

	ug.DELETE("/:id", DeleteUser()) // Delete user by ID

	logs.Dev("[AdminService] up at %s and %s", root.BasePath(), ug.BasePath())
}

// bring the service down gracefully and release all resources
func (svc *AdminService) Down() error {
	logs.Dev("auth service Down() not yet implemented")
	return fmt.Errorf("not yet implemented")
}

// returns the API version this depends on
func (svc *AdminService) Version() string {
	return svc.version
}

// any other services this depends on
func (svc *AdminService) DependsOn() []string {
	return nil
}

// returns a pointer to the server instance
// expects its state to be initialized and ready for Up()
func NewAdminService(endpoint string) *AdminService {
	logs.Dev("[NewAdminService]")
	cfg, err := configs.LoadConfig()
	if err != nil {
		logs.Fatal(err.Error())
	}
	m, err := mongo.NewMongoStore(cfg.DB.MongoURI, cfg.DB.Name)
	if err != nil {
		logs.Log("failed to create mongo store: %v", err)
		return nil
	}
	version := "v1"
	service = &AdminService{
		endpoint: endpoint,
		version:  version,
		userDB:   "users" + version,
		storage:  m,
	}
	return service
}

// for development purposes
func dummyString(length int, postfix string) string {
	const letters = "abcdefghijklmnopqrstuvwxyz01"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s%s", string(b), postfix)
}
