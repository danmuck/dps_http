package users

import (
	"fmt"
	"os"

	"github.com/danmuck/dps_http/mongo_client"
	"github.com/danmuck/dps_lib/logs"
	"github.com/gin-gonic/gin"
)

// interface impl
// //
var service *UsersConfig

type UsersConfig struct {
	endpoint string
	version  string

	secret  string
	userDB  string
	storage *mongo_client.MongoClient
}

func (svc *UsersConfig) Up(rg *gin.RouterGroup) {
	users := rg.Group("/users")
	// ug.Use(middleware.JWTMiddleware(), middleware.AuthorizeByRoles("user")) // Apply JWT and role middleware to all routes in this group

	users.GET("", ListUsers(svc.storage))
	users.POST("", CreateUser(svc.storage))
	users.GET("/:username/profile", GetUser(svc.storage))
	users.GET("/:username", func(c *gin.Context) {
		id := c.Param("username")
		logs.Dev("[DEV]> [NOIMPL] Get user by ID: %s", id)
	})
	auth := users.Group("/auth")
	auth.POST("/login", LoginHandler())
	auth.POST("/register", RegisterHandler())
	logs.Dev("login @ %s", auth.BasePath())

}

func (svc *UsersConfig) Down() error {
	logs.Dev("auth service Down() not yet implemented")
	return fmt.Errorf("not yet implemented")
}

func (svc *UsersConfig) DependsOn() []string {
	return nil
}

func Configure(client *mongo_client.MongoClient) *UsersConfig {
	service = &UsersConfig{
		endpoint: "users",
		version:  os.Getenv("VERSION"),
		userDB:   "users" + os.Getenv("VERSION"),
		secret:   os.Getenv("JWT_SECRET"),
		storage:  client,
	}
	return service
}
