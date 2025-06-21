package auth

import (
	"net/http"
	"time"

	"github.com/danmuck/dps_http/api/types"
	"github.com/danmuck/dps_http/lib/logs"
	"github.com/danmuck/dps_http/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RegisterHandler registers a new user, ensuring unique username and email,
// hashes the password, assigns roles, stores the user, and returns a JWT cookie.
func RegisterHandler(store storage.Client, jwtSecret string) gin.HandlerFunc {
	logs.Init("RegisterHandler() initializing with JWT secret: %s", jwtSecret)
	logs.Init("using storage type: %s", store.Type())
	logs.Init("using storage name: %s", store.Name())
	logs.Init("using storage: %T", store)
	return func(c *gin.Context) {

		var in registerPayload
		if err := c.ShouldBindJSON(&in); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if in.Password != in.Confirm {
			c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
			return
		}

		// uniqueness checks
		// could extend these
		if _, exists := store.Lookup("users", bson.M{"username": in.Username}); exists {
			c.JSON(http.StatusConflict, gin.H{"error": "username already in use"})
			return
		}
		if _, exists := store.Lookup("users", bson.M{"email": in.Email}); exists {
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}

		hash, err := HashPassword(in.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		roles := []string{"user"}
		if in.Username == "admin" || in.Username == "dirtpig" {
			logs.Log("assigning admin role to user: %s", in.Username)
			roles = append(roles, "admin")
		}
		user := types.User{
			ID:           primitive.NewObjectID(),
			Username:     in.Username,
			Email:        in.Email,
			PasswordHash: hash,
			Roles:        roles,
			Bio:          "Welcome to my office!",
			AvatarURL:    "",
			CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
			UpdatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		}
		logs.Log("creating user: %s", user.Username)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username": user.Username,
			"sub":      user.ID.Hex(),
			"roles":    user.Roles,
			"exp":      time.Now().Add(24 * time.Hour).Unix(),
		})
		logs.Log("signing token for user: %s", user.Username)
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}
		logs.Log("token signed successfully for user: %s \n  %v", user.Username, tokenString)

		if err := store.Store("users", user.ID.Hex(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		c.SetCookie("jwt", tokenString, 3600*24, "/", "localhost", false, true) // not secure for localhost
		c.SetCookie("username", user.Username, 3600*24, "/", "localhost", false, false)
		// c.SetCookie("sub", user.ID.Hex(), 3600*24, "/", "localhost", false, false)

		c.JSON(http.StatusCreated, gin.H{
			"status":   "ok",
			"username": user.Username,
		})
	}
}
