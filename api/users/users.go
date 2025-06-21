package users

import (
	// "encoding/json"

	"fmt"
	"math/rand"

	"github.com/danmuck/dps_http/configs"
	"github.com/danmuck/dps_http/lib/logs"
	"github.com/danmuck/dps_http/lib/middleware"
	"github.com/danmuck/dps_http/storage/mongo"
	"github.com/gin-gonic/gin"
)

var service *UserService

type UserService struct {
	endpoint string
	version  string
	userDB   string
	storage  *mongo.MongoClient
}

func (us *UserService) Up(rg *gin.RouterGroup) {
	rg.Use(middleware.JWTMiddleware(), middleware.AuthorizeByRoles("user")) // Apply JWT and role middleware to all routes in this group
	{
		// @NOTE -- locked to admin for development purposes
		rg.GET("/", middleware.AuthorizeByRoles("admin"), ListUsers())
		rg.GET("/:username", GetUser())
		uog := rg.Group("/")
		uog.Use(middleware.AuthorizeResourceAccess())
		{
			uog.GET("/r/:username", GetUser()) // Get user by ID
			uog.PUT("/:id", UpdateUser())      // Update user by ID
			uog.DELETE("/:id", DeleteUser())   // Delete user by ID
		}
		// TODO: dev route should be moved and locked away
		rg.POST("/:id", middleware.AuthorizeByRoles("admin"), CreateUser())
	}
}

func NewUserService(endpoint, version string) *UserService {
	cfg, err := configs.LoadConfig()
	if err != nil {
		logs.Fatal(err.Error())
	}
	m, err := mongo.NewMongoStore(cfg.DB.MongoURI, cfg.DB.Name)
	if err != nil {
		logs.Log("failed to create mongo store: %v", err)
		return nil
	}
	service = &UserService{
		endpoint: endpoint,
		version:  version,
		userDB:   endpoint + version,
		storage:  m,
	}
	return service
}

func dummyString(length int, postfix string) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ01"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s%s", string(b), postfix)
}
