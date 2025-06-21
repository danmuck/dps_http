package users

import (
	"net/http"

	"github.com/danmuck/dps_http/lib/logs"
	"github.com/gin-gonic/gin"
)

func ListUsers() gin.HandlerFunc {
	logs.Init("ListUsers from storage: %s", service.storage.Name())
	return func(c *gin.Context) {
		users, err := service.storage.List(service.userDB)
		if err != nil {
			logs.Err("failed to retrieve users: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to retrieve users",
			})
			return
		}

		logs.Log("listed %d users", len(users))
		c.JSON(http.StatusOK, users)
	}
}
