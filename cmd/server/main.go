package main

import (
	"log"
	"net/http"
	"time"

	"github.com/danmuck/dps_http/api/auth"
	"github.com/danmuck/dps_http/api/services/metrics"
	"github.com/danmuck/dps_http/api/users"
	"github.com/danmuck/dps_http/configs"
	"github.com/danmuck/dps_http/lib/logs"
	"github.com/danmuck/dps_http/lib/middleware"
	"github.com/danmuck/dps_http/storage"
	mongodb "github.com/danmuck/dps_http/storage/mongo"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type WebServer struct {
	cfg    *configs.Config
	router *gin.Engine
	store  storage.Client
}

func NewWebServer(cfg *configs.Config) *WebServer {
	store, err := mongodb.NewMongoStore(cfg.DB.MongoURI, cfg.DB.Name)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1", cfg.Domain})
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", cfg.Domain},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	ws := &WebServer{
		cfg:    cfg,
		store:  store,
		router: r,
	}
	return ws
}

func (ws *WebServer) registerServices() {
	log.Println("[api] Registering services")
	metrics.NewUserMetricsService().Up(ws.router.Group("metrics"))
}

func (ws *WebServer) registerRoutes() {
	middleware.InitAuthMiddleware(ws.cfg.Auth.JWTSecret)
	rg := ws.router.Group("/auth")
	{
		rg.POST("/login", auth.LoginHandler(ws.store, ws.cfg.Auth.JWTSecret))
		rg.POST("/logout", auth.LogoutHandler())
		rg.POST("/register", auth.RegisterHandler(ws.store, ws.cfg.Auth.JWTSecret))
	}
	ug := ws.router.Group("/users")
	ug.Use(middleware.JWTMiddleware(), middleware.AuthorizeByRoles("user")) // Apply JWT and role middleware to all routes in this group
	{
		// @NOTE -- locked to admin for development purposes
		ug.GET("/", middleware.AuthorizeByRoles("admin"), users.ListUsers(ws.store))
		ug.GET("/:username", users.GetUser(ws.store))
		uog := ug.Group("/")
		uog.Use(middleware.AuthorizeResourceAccess())
		{
			uog.GET("/r/:username", users.GetUser(ws.store)) // Get user by ID
			uog.PUT("/:id", users.UpdateUser(ws.store))      // Update user by ID
			uog.DELETE("/:id", users.DeleteUser(ws.store))   // Delete user by ID
		}
		// TODO: dev route should be moved and locked away
		ug.POST("/:id", middleware.AuthorizeByRoles("admin"), users.CreateUser(ws.store))
	}
	admin := ws.router.Group("/metrics")
	admin.Use(middleware.JWTMiddleware(), middleware.AuthorizeByRoles("admin"))

	dev := ws.router.Group("/dev")
	dev.Use(middleware.JWTMiddleware(), middleware.AuthorizeByRoles("dev"))

	// note: register all services
	ws.registerServices()

}

func init() {
	// tries to load .env, but won’t crash if it’s missing
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: no .env file found, relying on environment variables")
	}
}

func main() {

	cfg, err := configs.LoadConfig()
	if err != nil {
		logs.Fatal(err.Error())
	}

	server := NewWebServer(cfg)
	r := server.router

	root := r.Group("/")
	{
		root.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusMovedPermanently, gin.H{"message": "Welcome to DPS backend"})
		})
	}

	server.registerRoutes()
	r.Run(":" + cfg.Port)
}
