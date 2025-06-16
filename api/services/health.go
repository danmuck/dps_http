package services

import (
	"context"
	"net/http"
	"time"

	"github.com/danmuck/dps_http/storage"
	"github.com/gin-gonic/gin"
)

type HealthService struct {
	version string
	routes  []string
}

func NewHealthService(version string, routes []string) *HealthService {
	return &HealthService{
		version: version,
		routes:  routes,
	}
}

func (s *HealthService) Register(r *gin.Engine) {
	rg := r.Group("/health")
	{
		rg.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status": "ok",
			})
		})
	}
}

// Handler returns 200 OK only if Mongo is reachable within timeout.
func ServerHealthHandler(store storage.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Ping Mongo
		if err := store.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "unreachable",
				"error":    err.Error(),
			})
			return
		}

		// All good
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "reachable",
		})
	}
}
