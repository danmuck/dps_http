package main

import (
	"log"
	"net/http"
	"time"

	"github.com/danmuck/dps_http/api/admin"
	"github.com/danmuck/dps_http/api/auth"
	"github.com/danmuck/dps_http/api/metrics"
	"github.com/danmuck/dps_http/api/users"
	api "github.com/danmuck/dps_http/api/v1"
	"github.com/danmuck/dps_http/configs"
	"github.com/danmuck/dps_http/middleware"
	logs "github.com/danmuck/dps_lib/logs"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type WebServer struct {
	cfg      *configs.Config
	services map[string]api.Service
	router   *gin.Engine
}

func NewWebServer(cfg *configs.Config) *WebServer {
	router := gin.Default()
	// r.SetTrustedProxies([]string{"127.0.0.1", cfg.Domain})
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3031", cfg.Domain},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(gin.Logger(), gin.Recovery())
	router.Use(func(c *gin.Context) {
		logs.Dev("Incoming request: %s %s (origin: %s)",
			c.Request.Method, c.Request.URL.Path, c.Request.Header.Get("Origin"))
		logs.Debug("[DEBUG]> Headers: %v", c.Request.Header)
		logs.Debug("[DEBUG]> Body: %s", c.Request.Body)
		c.Next()
	})
	ws := &WebServer{
		cfg:      cfg,
		router:   router,
		services: make(map[string]api.Service),
	}
	return ws
}

func (ws *WebServer) registerServices() {
	log.Println("[api] Registering services")
	root := ws.router.Group("/")
	{
		root.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusMovedPermanently, gin.H{"message": "Welcome to DPS backend"})
		})
	}

	ws.services["auth"] = auth.NewAuthService("auth")
	ws.services["users"] = users.NewUserService("users")
	ws.services["metrics"] = metrics.NewUserMetricsService("metrics")
	ws.services["admin"] = admin.NewAdminService("admin")

	v1 := root.Group(api.VERSION)
	var buf []api.Service = make([]api.Service, 0)
	for key, svc := range ws.services {
		if svc.DependsOn() != nil {
			buf = append(buf, svc)
			continue
		}
		if key == "admin" {
			svc.Up(v1)
			continue
		}
		svc.Up(v1)
	}
}

func (ws *WebServer) registerRoutes() {
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

	server.registerRoutes()
	defer func() {
		for _, svc := range server.services {
			svc.Down()
		}
	}()
	r.Run(":" + cfg.Port)
}
