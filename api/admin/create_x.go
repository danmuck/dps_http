package admin

import (
	"net/http"
	"strconv"

	"github.com/danmuck/dps_http/lib/logs"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateUsersX() gin.HandlerFunc {
	logs.Init("[DEV]> CreateUser() initializing with storage: %s", service.storage.Name())
	return func(c *gin.Context) {

		X := c.PostForm("num_users")
		N, err := strconv.Atoi(X)
		if err != nil {
			logs.Log("[DEV]> CreateUser: received invalid number of users: %s", X)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid number of users",
			})
			return
		}

		logs.Log("[DEV]> CreateUser: received username: %s", X)
		for range N {
			username := dummyString(8, "dps")
			if _, found := service.storage.Lookup(service.userDB, bson.M{"username": username}); found {
				logs.Log("[DEV]> User %s already exists, generating a new one", username)
				username = dummyString(8, "dps")
			}

			if err := createDummyUser(username); err != nil {
				logs.Dev("[DEV]> CreateUser: error: %s", err)
				continue
			}
		}
		c.Redirect(http.StatusFound, "/admin/new")
	}
}
