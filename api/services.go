package api

import (
	"github.com/danmuck/dps_http/api/logs"
	"github.com/danmuck/dps_http/api/services"
	"github.com/danmuck/dps_http/api/services/metrics"
	"github.com/danmuck/dps_http/storage"
	"github.com/gin-gonic/gin"
)

type Service struct {
	Name     string
	Version  string
	Endpoint string
	Bucket   string
	Running  bool
}

// Services must register their own routes with a gin router
type ServiceReg interface {
	Register(r *gin.Engine)
	Start() error
	Stop() error
}

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
