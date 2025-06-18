package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/danmuck/dps_http/api/auth"
	"github.com/danmuck/dps_http/api/services/metrics"
	"github.com/danmuck/dps_http/api/users"
	"github.com/danmuck/dps_http/configs"
	"github.com/danmuck/dps_http/middleware"
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
	ums := metrics.NewUserMetricsService(
		ws.store.ConnectOrCreateBucket("users"),
		ws.store.ConnectOrCreateBucket("metrics"),
	)
	ums.Register(ws.router)
	ums.Start()
}

func (ws *WebServer) registerRoutes() {
	rg := ws.router.Group("/auth")
	{
		rg.POST("/login", auth.LoginHandler(ws.store, ws.cfg.Auth.JWTSecret))
		rg.POST("/logout", auth.LogoutHandler())
		rg.POST("/register", auth.RegisterHandler(ws.store, ws.cfg.Auth.JWTSecret))
	}
	ug := ws.router.Group("/users")
	ug.Use(middleware.JWTMiddleware([]byte(ws.cfg.Auth.JWTSecret)), middleware.RoleMiddleware("user")) // Apply JWT and role middleware to all routes in this group
	{
		ug.GET("/", users.ListUsers(ws.store))
		ug.GET("/:username", users.GetUser(ws.store))
		ug.PUT("/:id", users.UpdateUser(ws.store))    // Update user by ID
		ug.DELETE("/:id", users.DeleteUser(ws.store)) // Delete user by ID
		// TODO: dev route should be moved and locked away
		ug.POST("/:id", middleware.RoleMiddleware("admin"), users.CreateUser(ws.store))
	}
	admin := ws.router.Group("/metrics")
	admin.Use(middleware.JWTMiddleware([]byte(ws.cfg.Auth.JWTSecret)), middleware.RoleMiddleware("admin"))

	dev := ws.router.Group("/dev")
	dev.Use(middleware.JWTMiddleware([]byte(ws.cfg.Auth.JWTSecret)), middleware.RoleMiddleware("dev"))

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
	uri := os.Getenv("MONGO_URI")
	jwt := os.Getenv("JWT_SECRET")
	domain := os.Getenv("DOMAIN")
	if domain == "" {
		domain = "127.0.0.1"
	}
	log.Println("Environment: { uri: ", uri, ", jwt: ", jwt, " }")
	if uri == "" || jwt == "" {
		log.Fatal("MONGO_URI and JWT_SECRET must be set in environment variables or .env file")
	}

	cfg := &configs.Config{
		Domain: domain,
		Port:   ":8080",
		DB: configs.Storage{
			T:        "mongo",
			Name:     "dps_http",
			MongoURI: os.Getenv("MONGO_URI"),
		},
		Auth: configs.Auth{
			JWTSecret: os.Getenv("JWT_SECRET"),
		},
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
	r.Run(cfg.Port)
}
