package admin

import (
	"net/http"
	"strconv"

	"github.com/danmuck/dps_http/lib/logs"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func CreateUsersX() gin.HandlerFunc {
	return func(c *gin.Context) {

		X := c.PostForm("num_users")
		N, err := strconv.Atoi(X)

		if err != nil {
			logs.Err("[DEV]> CreateUser: received invalid number of users: %s", X)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid number of users",
			})
			return
		}
		go func() {
			logs.Info("[DEV]> CreateUser: creating ++(%d) dummy users", N)
			if err := CreateXUsers(N); err != nil {
				logs.Err("[DEV]> CreateUser: failed to create %d dummy users: %v", N, err)
			} else {
				logs.Info("[DEV]> CreateUser: successfully created ++(%d) dummy users", N)
			}
		}()
		// err = CreateXUsers(N)
		// if err != nil {
		// 	logs.Err("[DEV]> CreateUser: failed to create %d dummy users: %v", N, err)
		// 	c.JSON(http.StatusInternalServerError, gin.H{
		// 		"status": "error",
		// 		"error":  "failed to create users",
		// 	})
		// 	return
		// }
		c.Redirect(http.StatusFound, "/admin/new")
	}
}

func CreateXUsers(x int) error {
	logs.Info("[DEV]> Creating ++(%d) dummy users", x)
	for range x {
		username := dummyString(8, "dps")
		if _, found := service.storage.Lookup(service.userDB, bson.M{"username": username}); found {
			logs.Log("[DEV]> User %s already exists, generating a new one", username)
			username = dummyString(8, "dps")
		}

		if err := createDummyUser(username); err != nil {
			logs.Warn("[DEV]> CreateUser: error: %s, user not created", err)
			continue
		}
	}
	logs.Warn("[DEV]> Created ++(%d) dummy users", x)

	return nil
}
