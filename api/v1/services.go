package v1

import (
	"github.com/gin-gonic/gin"
)

const VERSION = "api/v1"

// Services must register their own routes with a gin router
// service registration interface: services must implement register, start, and stop methods
type Service interface {
	Up(rg *gin.RouterGroup) // register service routes with gin engine
	Down() error            // stop the service
	DependsOn() []string    // returns a list of services this service depends on
	Version() string        // returns the version of the service
}

// registry holds references to all registered services
type Registry struct {
	// UserMetrics *metrics.UserMetricsService
	// Health      *services.HealthService
}
