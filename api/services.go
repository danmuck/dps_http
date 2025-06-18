package api

import (
	"github.com/danmuck/dps_http/api/logs"
	"github.com/danmuck/dps_http/api/services"
	"github.com/danmuck/dps_http/api/services/metrics"
	"github.com/danmuck/dps_http/storage"
	"github.com/gin-gonic/gin"
)

type Service struct {
	name     string // the unique name of the service
	version  string // the version identifier of the service
	endpoint string // the network address or url where the service can be accessed
	bucket   string // the storage bucket or logical grouping associated with the service
	running  bool   // indicates whether the service is currently running (true) or stopped (false)
}

// Services must register their own routes with a gin router
// service registration interface: services must implement register, start, and stop methods
type ServiceReg interface {
	Register(r *gin.Engine) // register service routes with gin engine
	Start() error           // start the service
	Stop() error            // stop the service
}

// registry holds references to all registered services
type Registry struct {
	UserMetrics *metrics.UserMetricsService
	Health      *services.HealthService
}

func (r *Registry) registerServices(store storage.Client, router *gin.Engine) {
	logs.Info("registering services")

	// User Metrics Service
	// //
	r.UserMetrics = metrics.NewUserMetricsService(
		store.ConnectOrCreateBucket("users"),
		store.ConnectOrCreateBucket("metrics"),
	)
	r.UserMetrics.Register(router)
	r.UserMetrics.Start()
	defer r.UserMetrics.Stop()
}
