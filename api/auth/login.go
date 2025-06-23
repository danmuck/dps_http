package auth

import (
	"net/http"
	"time"

	api "github.com/danmuck/dps_http/api/v1"
	"github.com/danmuck/dps_http/lib/logs"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

// LoginHandler handles user login.
// It validates the input, looks up the user by username,
// checks the password against the stored hash, and signs a JWT token.
func LoginHandler() gin.HandlerFunc {
	logs.Init("LoginHandler() initializing with JWT secret: %s", service.secret)

	return func(c *gin.Context) {
		var in loginPayload
		if err := c.ShouldBindJSON(&in); err != nil {
			logs.Log("bind error: %v", err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// lookup user by username
		logs.Log("received login request for user: %s", in.Username)
		raw, found := service.storage.Lookup(service.userDB, bson.M{"username": in.Username})
		if !found {
			logs.Log("user not found: %s", in.Username)
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}
		logs.Log("user found: %s", in.Username)
		var user api.User
		data, _ := bson.Marshal(raw)
		if err := bson.Unmarshal(data, &user); err != nil {
			logs.Log("unmarshal error: %v", err)
			c.JSON(500, gin.H{"error": "server error"})
			return
		}

		// validate password against stored hash
		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)) != nil {
			logs.Log("invalid password for user: %s", in.Username)
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}

		// sign jwt token with secret loaded from config
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": user.Username,
			"sub":      user.ID.Hex(),
			"roles":    user.Roles,
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		})
		signed, err := token.SignedString([]byte(service.secret))
		if err != nil {
			logs.Log("failed to sign token for user %s: %v", user.Username, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}

		logs.Log("token signed successfully for user: %s \n  ...%v with hash: %s",
			user.Username, signed[len(signed)-20:], service.secret)

		c.SetCookie("jwt", signed, 3600*24, "/", "localhost", false, true)
		c.SetCookie("username", user.Username, 3600*24, "/", "localhost", false, false)
		// c.SetCookie("sub", user.ID.Hex(), 3600*24, "/", "localhost", false, false)

		c.JSON(http.StatusOK, gin.H{
			"username": user.Username,
		})
	}
}
