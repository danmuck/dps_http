package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/danmuck/dps_http/api/users"
	"github.com/danmuck/dps_http/middleware"
	"github.com/danmuck/dps_http/storage"
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
	store  storage.Storage
}

func NewWebServer(cfg *Config) *WebServer {
	store, err := storage.NewMongoStore(cfg.db.MongoURI, cfg.db.Name)
	if err != nil {
		log.Fatalf("failed to connect to MongoDB: %v", err)
	}

	// Initialize Gin --
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1", cfg.domain})
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
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

func (ws *WebServer) registerRoutes() {
	ug := ws.router.Group("/users")
	{
		ug.GET("/", users.ListUsers(ws.store))        // List all users
		ug.POST("/:id", users.CreateUser(ws.store))   // Create a new user
		ug.GET("/:username", users.GetUser(ws.store)) // Get user by username

		ug.PUT("/:id", users.UpdateUser(ws.store))    // Update user by ID
		ug.DELETE("/:id", users.DeleteUser(ws.store)) // Delete user by ID
	}
	dev := ws.router.Group("/dev")
	dev.Use(middleware.JWTMiddleware(), middleware.RoleMiddleware("dev"))
	admin := ws.router.Group("/admin")
	admin.Use(middleware.JWTMiddleware(), middleware.RoleMiddleware("admin"))

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
	// store := server.store

	root := r.Group("/")
	{
		root.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Welcome to DPS backend"})
		})
		// root.GET("/health", services.ServerHealthHandler(store))
	}
	// r.GET("/prom", gin.WrapH(promhttp.Handler()))
	//
	server.registerRoutes()

	// Start server
	r.Run(cfg.port)
}
