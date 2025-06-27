package main

// import (
// 	"log"
// 	"net/http"
// 	"time"

// 	"github.com/danmuck/dps_http/v1/admin"
// 	"github.com/danmuck/dps_http/v1/auth"
// 	"github.com/danmuck/dps_http/v1/configs"
// 	"github.com/danmuck/dps_http/v1/lib/logs"
// 	"github.com/danmuck/dps_http/v1/metrics"
// 	"github.com/danmuck/dps_http/v1/middleware"
// 	"github.com/danmuck/dps_http/v1/users"
// 	api "github.com/danmuck/dps_http/v1/v1"
// 	"github.com/gin-contrib/cors"
// 	"github.com/gin-gonic/gin"
// 	"github.com/joho/godotenv"
// )

// type WebServer struct {
// 	cfg      *configs.Config
// 	services map[string]api.Service
// 	router   *gin.Engine
// }

// func NewWebServer(cfg *configs.Config) *WebServer {
// 	r := gin.Default()
// 	r.SetTrustedProxies([]string{"127.0.0.1", cfg.Domain, "localhost:3031"})
// 	r.Use(gin.Logger(), gin.Recovery())

// 	r.Use(cors.New(cors.Config{
// 		AllowOrigins:     []string{"http://localhost:3000", cfg.Domain},
// 		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
// 		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
// 		ExposeHeaders:    []string{"Content-Length"},
// 		AllowCredentials: true,
// 		MaxAge:           12 * time.Hour,
// 	}))

// 	r.Use(func(c *gin.Context) {
// 		logs.Dev("Incoming request origin: %s", c.Request.Header.Get("Origin"))
// 		c.Next()
// 	})
// 	r.Use(func(c *gin.Context) {
// 		logs.Dev("Incoming request: %s %s (origin: %s)", c.Request.Method, c.Request.URL.Path, c.Request.Header.Get("Origin"))
// 		c.Next()
// 	})
// 	ws := &WebServer{
// 		cfg:      cfg,
// 		router:   r,
// 		services: make(map[string]api.Service),
// 	}
// 	return ws
// }

// func (ws *WebServer) registerServices() {
// 	log.Println("[api] Registering services")
// 	root := ws.router.Group("/")
// 	{
// 		root.GET("/", func(c *gin.Context) {
// 			c.JSON(http.StatusMovedPermanently, gin.H{"message": "Welcome to DPS backend"})
// 		})
// 	}

// 	ws.services["auth"] = auth.NewAuthService("auth")
// 	ws.services["users"] = users.NewUserService("users")
// 	ws.services["metrics"] = metrics.NewUserMetricsService("metrics")
// 	ws.services["admin"] = admin.NewAdminService("admin")

// 	v1 := root.Group(api.VERSION)
// 	var buf []api.Service = make([]api.Service, 0)
// 	for key, svc := range ws.services {
// 		if svc.DependsOn() != nil {
// 			buf = append(buf, svc)
// 			continue
// 		}
// 		if key == "admin" {
// 			svc.Up(v1)
// 			continue
// 		}
// 		svc.Up(v1)
// 	}
// }

// func (ws *WebServer) registerRoutes() {
// 	dev := ws.router.Group("/dev")
// 	dev.Use(middleware.JWTMiddleware(), middleware.AuthorizeByRoles("dev"))

// 	// note: register all services
// 	ws.registerServices()

// }

// func init() {
// 	// tries to load .env, but won’t crash if it’s missing
// 	if err := godotenv.Load(); err != nil {
// 		log.Println("Warning: no .env file found, relying on environment variables")
// 	}
// }

// func main() {

// 	cfg, err := configs.LoadConfig()
// 	if err != nil {
// 		logs.Fatal(err.Error())
// 	}

// 	server := NewWebServer(cfg)
// 	r := server.router

// 	server.registerRoutes()
// 	defer func() {
// 		for _, svc := range server.services {
// 			svc.Down()
// 		}
// 	}()
// 	r.Run(":" + cfg.Port)
// }
