package api

import (
	"github.com/gin-gonic/gin"
)

// type Service struct {
// 	name     string // the unique name of the service
// 	version  string // the version identifier of the service
// 	endpoint string // the network address or url where the service can be accessed
// 	bucket   string // the storage bucket or logical grouping associated with the service
// 	running  bool   // indicates whether the service is currently running (true) or stopped (false)
// }

// Services must register their own routes with a gin router
// service registration interface: services must implement register, start, and stop methods
type Service interface {
	Up(rg *gin.RouterGroup) // register service routes with gin engine
	Down() error            // stop the service
}

// registry holds references to all registered services
type Registry struct {
	// UserMetrics *metrics.UserMetricsService
	// Health      *services.HealthService
}
