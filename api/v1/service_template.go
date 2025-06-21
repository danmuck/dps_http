package v1

import (
	"fmt"

	"github.com/danmuck/dps_http/configs"
	"github.com/danmuck/dps_http/lib/logs"
	"github.com/danmuck/dps_http/lib/middleware"
	"github.com/danmuck/dps_http/lib/storage/mongo"

	"github.com/gin-gonic/gin"
)

// TemplateService implements the service interface
// route handlers are defined in their corresponding files
// (note: there are none for this template service)
// //
var service *TemplateService

type TemplateService struct {
	endpoint string // endpoint for the service; canonically /api/v_/<service_name>
	version  string // api version; the version this template resides in

	// service specific structures
	userDB  string
	storage *mongo.MongoClient
}

// assign routes for the service and initialize any resources
// routes are structured `api/v1/<service_name>/<your_endpoints>`
func (svc *TemplateService) Up(rg *gin.RouterGroup) {
	svc_g := rg.Group(Path(service.endpoint))
	svc_g.Use(middleware.JWTMiddleware()) // Apply middleware from lib/ if needed
	// svc_g.GET("/", handlerFunc())

}

// bring the service down gracefully and release all resources
func (svc *TemplateService) Down() error {
	logs.Dev("auth service Down() not yet implemented")
	return fmt.Errorf("not yet implemented")
}

// returns the API version this depends on
func (svc *TemplateService) Version() string {
	return svc.version
}

// any other services this depends on
func (svc *TemplateService) DependsOn() []string {
	return nil
}

// returns a pointer to the server instance
// expects its state to be initialized and ready for Up()
func NewUserService(endpoint string) *TemplateService {
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
	service = &TemplateService{
		endpoint: endpoint,
		version:  version,
		userDB:   endpoint + version,
		storage:  m,
	}
	return service
}
