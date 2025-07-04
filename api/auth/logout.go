package auth

import (
	"net/http"

	logs "github.com/danmuck/dps_lib/logs"

	"github.com/gin-gonic/gin"
)

// LogoutHandler clears the JWT cookie to log the user out.
// note: It does not invalidate the JWT on the server side, but removes it from the client.
func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logs.Log("LogoutHandler() clearing JWT cookie")
		c.SetCookie("jwt", "", -1, "/", "", true, true)
		c.SetCookie("username", "", -1, "/", "", true, false)
		c.JSON(http.StatusOK, gin.H{"status": "logged out"})
	}
}
