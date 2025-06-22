package admin

import (
	"net/http"

	"github.com/danmuck/dps_http/lib/logs"
	"github.com/gin-gonic/gin"
)

func DeleteUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := service.storage.Delete(service.userDB, c.Param("id")); err != nil {
			logs.Err("DeleteUser: failed to delete user %s: %v", c.Param("id"), err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to delete user",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "user deleted successfully",
		})
	}
}
