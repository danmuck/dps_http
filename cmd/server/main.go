package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/danmuck/dps_http/api/auth"
	"github.com/danmuck/dps_http/api/services"
	"github.com/danmuck/dps_http/api/users"
	"github.com/danmuck/dps_http/middleware"
	"github.com/danmuck/dps_http/storage"
	mongodb "github.com/danmuck/dps_http/storage/mongo"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	// "github.com/prometheus/client_golang/prometheus/promhttp"
	// ginprom "github.com/zsais/go-gin-prometheus"
)

type Config struct {
	domain string
	port   string
	db     Storage
	auth   Auth
}
type Storage struct {
	t        string // type for future expansion default: "mongo"
	MongoURI string
	Name     string // database name
}
type Auth struct {
	JWTSecret string
}
type WebServer struct {
	cfg    *Config
	router *gin.Engine
	store  storage.Client
}

func NewWebServer(cfg *Config) *WebServer {
	store, err := mongodb.NewMongoStore(cfg.db.MongoURI, cfg.db.Name)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	// Initialize Gin --
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1", cfg.domain})
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
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
	// Register services
	log.Println("[api] Registering services")
	// UserMetricsService
	ums := services.NewUserMetricsService(
		ws.store.ConnectOrCreateBucket("users"),
		ws.store.ConnectOrCreateBucket("metrics"),
	)
	ums.Register(ws.router)
	ums.Start()
}

func (ws *WebServer) registerRoutes() {
	rg := ws.router.Group("/auth")
	{
		rg.POST("/login", auth.LoginHandler(ws.store, ws.cfg.auth.JWTSecret))
		rg.POST("/logout", auth.LogoutHandler())
		rg.POST("/register", auth.RegisterHandler(ws.store, ws.cfg.auth.JWTSecret))
		// rg.GET("/health", storage.ServerHealthHandler(ws.store))
	}
	ug := ws.router.Group("/users")
	ug.Use(middleware.JWTMiddleware([]byte(ws.cfg.auth.JWTSecret)), middleware.RoleMiddleware("user")) // Apply JWT and role middleware to all routes in this group
	{
		ug.GET("/", users.ListUsers(ws.store))
		ug.GET("/:username", users.GetUser(ws.store))
		// TODO: dev route should be moved and locked away
		ug.POST("/:id", users.CreateUser(ws.store))

		ug.PUT("/:id", users.UpdateUser(ws.store))    // Update user by ID
		ug.DELETE("/:id", users.DeleteUser(ws.store)) // Delete user by ID
	}
	admin := ws.router.Group("/metrics")
	admin.Use(middleware.JWTMiddleware([]byte(ws.cfg.auth.JWTSecret)), middleware.RoleMiddleware("admin"))
	// {
	// 	umg := admin.Group("/users")
	// 	// umg.GET("/roles", users.GetUserByID(ws.store))      // Get user by ID
	// 	umg.GET("/", users.MetricsHandler(
	// 		ws.store.ConnectOrCreateBucket("users"),
	// 		ws.store.ConnectOrCreateBucket("metrics"))) // List all users
	// }
	dev := ws.router.Group("/dev")
	dev.Use(middleware.JWTMiddleware([]byte(ws.cfg.auth.JWTSecret)), middleware.RoleMiddleware("dev"))

	ws.registerServices() // Register all services

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
	log.Println("Environment: { uri: ", uri, ", jwt: ", jwt, " }")
	if uri == "" || jwt == "" {
		log.Fatal("MONGO_URI and JWT_SECRET must be set in environment variables or .env file")
	}

	// Load configuration
	cfg := &Config{
		domain: "127.0.0.1",
		port:   ":8080",
		db: Storage{
			// needs to be updated alongside the storage/ api
			t:        "mongo",
			Name:     "dps_http",
			MongoURI: os.Getenv("MONGO_URI"),
		},
		auth: Auth{
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
	// r.GET("/prom", gin.WrapH(promhttp.Handler()))
	//
	server.registerRoutes()

	// start server
	r.Run(cfg.port)
}
