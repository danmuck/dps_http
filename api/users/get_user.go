package users

import (
	"net/http"

	api "github.com/danmuck/dps_http/api/v1"
	"github.com/danmuck/dps_http/lib/logs"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		logs.Init("GetUser getting %s", c.Param("username"))
		key := c.Param("username")

		// retrieve the raw map from storage
		raw, ok := service.storage.Lookup(service.userDB, bson.M{"username": key})
		if !ok {
			logs.Log("not found: %s", key)
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		logs.Log("raw user map: %v", raw["username"])

		rawBSON, _ := bson.Marshal(raw)
		var user api.User
		if err := bson.Unmarshal(rawBSON, &user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user"})
			return
		}

		// return the User struct
		logs.Log("got user %s", user.String())
		c.JSON(http.StatusOK, user)
	}
}
