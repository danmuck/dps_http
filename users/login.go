package users

import (
	"context"
	"net/http"

	"github.com/danmuck/dps_lib/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
)

// loginPayload represents the input for user login.
type loginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginHandler handles user login.
// It validates the input, looks up the user by username,
// checks the password against the stored hash, and signs a JWT token.
func LoginHandler() gin.HandlerFunc {
	logs.Init("LoginHandler() initializing with JWT secret: %s", service.secret)

	return func(c *gin.Context) {
		var req loginPayload
		if err := c.ShouldBindJSON(&req); err != nil {
			logs.Log("bind error: %v", err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		user := &User{}
		err := service.storage.Collection("users").FindOne(context.Background(), bson.M{"username": req.Username}).Decode(user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				logs.Err("user %s not found", req.Username)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
				return
			}
			logs.Err("failed to find user %s: %v", req.Username, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		logs.Debug("got user %s", user.String())
		logs.Debug("user found: %s", req.Username)
		// validate password against stored hash
		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
			logs.Err("invalid password for user: %s", req.Username)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
			return
		}

		GenAndSetJWT(c, user, service.secret)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "login successful",
			"data": gin.H{
				"username": user.Username,
				"success":  true,
			},
		})
	}
}
