package server

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/danmuck/dps_http/mongo_client"
	"github.com/danmuck/dps_lib/logs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	// tries to load .env, but won’t crash if it’s missing
	if err := godotenv.Load(); err != nil {
		logs.Info("Warning: no .env file found, relying on environment variables")
	}

	if os.Getenv("MONGO_URI") == "" {
		logs.Fatal("MONGO_URI environment variable must be set")
	}
	if os.Getenv("MONGO_DB") == "" {
		logs.Fatal(" MONGO_DB environment variable must be set")
	}
	if os.Getenv("VERSION") == "" {
		logs.Fatal("VERSION environment variable must be set")
	}
	if os.Getenv("DOMAIN") == "" {
		logs.Fatal("DOMAIN environment variable must be set")
	}
	if os.Getenv("PORT") == "" {
		logs.Fatal("PORT environment variable must be set")
	}
	if os.Getenv("CLIENT") == "" {
		logs.Fatal("CLIENT environment variable must be set")
	}
	if os.Getenv("CLIENT_PORT") == "" {
		logs.Fatal("CLIENT_PORT environment variable must be set")
	}
	logs.Info("Environment variables loaded successfully")
	logs.Dev("[DEV]> .env > \n %s \n %s \n %s \n %s \n %s \n %s",
		os.Getenv("VERSION"), os.Getenv("MONGO_URI"), os.Getenv("MONGO_DB"),
		os.Getenv("DOMAIN"), os.Getenv("PORT"), os.Getenv("CLIENT_IP"))
}

type HTTPServer struct {
	Version  string
	Domain   string
	Port     string
	ClientIP string // ClientIP is the IP address of the client making the request

	Mongo  *mongo_client.MongoClient
	router *gin.Engine
}

func (s *HTTPServer) Start() error {
	address := ":" + s.Port
	logs.Info("Starting HTTP server on %s", address)
	return s.router.Run(address)
}

func (s *HTTPServer) Stop() error {
	logs.Info("Stopping HTTP server on %s", s.Domain+":"+s.Port)
	if err := s.Mongo.Client().Disconnect(context.Background()); err != nil {
		logs.Err("Failed to disconnect from MongoDB: %v", err)
		return err
	}
	return nil
}

func (s *HTTPServer) Router() *gin.RouterGroup {
	rootPath := fmt.Sprintf("api/v%s", s.Version)
	rg := s.router.Group(rootPath)
	return rg
}

func NewHTTPServer() *HTTPServer {
	mongo, err := mongo_client.NewMongoStore(
		os.Getenv("MONGO_URI"), // Replace with your MongoDB URI
		os.Getenv("MONGO_DB"),  // Database name for user storage
	)
	if err != nil {
		logs.Fatal("Failed to connect to MongoDB: %v", err)
	}
	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3031", os.Getenv("CLIENT") + ":" + os.Getenv("CLIENT_PORT")},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// router.SetTrustedProxies([]string{os.Getenv("CLIENT")})
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(func(c *gin.Context) {
		logs.Dev("Incoming request: %s %s (origin: %s)",
			c.Request.Method, c.Request.URL.Path, c.Request.Header.Get("Origin"))
		logs.Debug("[DEBUG]> Headers: %v", c.Request.Header)
		logs.Debug("[DEBUG]> Body: %s", c.Request.Body)
		c.Next()
	})

	logs.Info("HTTP server initialized with version %s, domain %s, port %s, client IP %s",
		os.Getenv("VERSION"), os.Getenv("DOMAIN"), os.Getenv("PORT"), os.Getenv("CLIENT")+" "+os.Getenv("CLIENT_PORT"))
	return &HTTPServer{
		Version:  os.Getenv("VERSION"),
		Domain:   os.Getenv("DOMAIN"),
		Port:     os.Getenv("PORT"),
		ClientIP: os.Getenv("CLIENT") + ":" + os.Getenv("CLIENT_PORT"),

		Mongo:  mongo,
		router: router,
	}
}
