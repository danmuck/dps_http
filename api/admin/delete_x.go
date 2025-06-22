package admin

import (
	"net/http"
	"slices"
	"strconv"

	"github.com/danmuck/dps_http/lib/logs"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func DeleteUsersX() gin.HandlerFunc {
	return func(c *gin.Context) {
		X := c.PostForm("num_users")
		N, err := strconv.Atoi(X)
		if err != nil {
			logs.Log("[DEV]> DeleteUsersX: invalid number of users: %s", X)
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "invalid number of users",
			})
			return
		}

		bucket := service.storage.ConnectOrCreateBucket(service.userDB)
		items, err := bucket.ListItems()
		if err != nil {
			logs.Err("Could not list user items: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to list users",
			})
			return
		}

		deleted := 0
		for _, doc := range items {
			if deleted >= N {
				break
			}

			// Get key (string)
			keyStr, ok := doc["key"].(string)
			if !ok {
				logs.Err("Skipping document with invalid key: %+v", doc["key"])
				continue
			}

			// Get value map
			val, ok := doc["value"].(map[string]any)
			if !ok {
				logs.Err("Skipping document with invalid value: %+v", doc["value"])
				continue
			}

			// Check roles
			rolesAny, ok := val["roles"].(primitive.A)
			if !ok {
				logs.Err("Skipping user %s: roles field not primitive.A", keyStr)
				continue
			}

			var roles []string
			for _, r := range rolesAny {
				if s, ok := r.(string); ok {
					roles = append(roles, s)
				}
			}

			if !slices.Contains(roles, "dummy") {
				continue
			}

			// Delete user
			err := bucket.Delete(keyStr)
			if err != nil {
				logs.Err("Failed to delete user %s: %v", keyStr, err)
				continue
			}

			logs.Dev("Deleted dummy user: %s", keyStr)
			deleted++
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"deleted": deleted,
		})
	}
}
