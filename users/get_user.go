package users

import (
	"context"
	"net/http"

	"github.com/danmuck/dps_http/mongo_client"
	"github.com/danmuck/dps_lib/logs"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func WrapResponse(payload any, success bool, msg, err string) gin.H {
	response := gin.H{
		"success": success,
		"message": msg,
		"error":   err,
		"data":    payload,
	}
	return response
}

type GetUserPayload struct {
	Username string `json:"username"`
	ID       string `json:"id"`
}

func GetUser(client *mongo_client.MongoClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("username")
		// var req loginPayload
		// if err := c.ShouldBindJSON(&req); err != nil {
		// 	logs.Log("bind error: %v", err)
		// 	c.JSON(400, gin.H{"error": err.Error()})
		// 	return
		// }
		// username := c.Param("username")
		// if username == "" {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		// 	return
		// }
		logs.Debug("received request to get user: %s", username)
		user := &User{}
		err := service.storage.Collection("users").FindOne(context.Background(), bson.M{"username": username}).Decode(user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				logs.Err("user %s not found", username)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
				return
			}
			logs.Err("failed to find user %s: %v", username, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		logs.Debug("got user %s", user.String())

		c.JSON(http.StatusOK, WrapResponse(user, true, "user found", ""))
	}
}
