package metrics

import (
	"net/http"

	"github.com/danmuck/dps_http/lib/logs"
	"github.com/gin-gonic/gin"
)

func UserGrowth(svc *UserMetricsService) gin.HandlerFunc {
	// note: this is a singleton service, so we can use a single instance
	// it needs to be initialized at the server
	logs.Init("initializing service handler [%s.%s]", svc.endpoint, svc.version)
	return func(c *gin.Context) {
		err := svc.UpdateTotalUsers()
		if err != nil {
			logs.Err("failed to get user count: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get user count",
			})
			return
		}
		err = svc.UpdateRoleCounts()
		if err != nil {
			logs.Err("failed to get user count by role: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get user count by role",
			})
			return
		}
		service.AddGrowthData()
		service.WriteMetrics()
		svc.mu.Lock()
		defer svc.mu.Unlock()

		users_over_time_points := MapTimestampToInt64Points(svc.users_over_time)
		logs.Info("total users: %d, total roles: %v", svc.total_users, svc.total_roles)
		if svc.total_users == 0 {
			logs.Warn("no users found, returning empty metrics")
			c.JSON(http.StatusOK, gin.H{
				"total_users":     0,
				"total_roles":     make(map[string]int64),
				"users_over_time": []Point{},
				"message":         "no users found",
			})
			return
		}

		logs.Debug("service: %+v", svc)

		c.JSON(http.StatusOK, gin.H{
			"total_users":     svc.total_users,
			"total_roles":     svc.total_roles,
			"users_over_time": users_over_time_points,
			"message":         "user metrics retrieved successfully",
		})
	}
}
