package users

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/danmuck/dps_lib/logs"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// registerPayload defines the input for user registration.
type registerPayload struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Confirm  string `json:"confirm"  binding:"required,eqfield=Password"` // confirm password must match
}

func RegisterHandler() gin.HandlerFunc {
	logs.Init("register storage: %s", service.storage.Name())
	return func(c *gin.Context) {

		var req registerPayload
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Password != req.Confirm {
			c.JSON(http.StatusBadRequest, gin.H{"error": "passwords do not match"})
			return
		}

		if exists := VerifyUsername(req.Username); exists {
			logs.Err("Username already exists: %s", req.Username)
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("username %s already in use", req.Username)})
			return
		}
		if exists := VerifyEmail(req.Email); exists {
			logs.Err("Email already exists: %s", req.Email)
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("email %s already in use", req.Email)})
			return
		}

		hash, err := HashPassword(req.Password)
		if err != nil {
			logs.Err("Failed to hash password: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		roles := []string{"user"}
		if req.Username == "admin" || req.Username == "dirtpig" || req.Username == "danmuck" {
			logs.Dev("assigning admin role to user: %s", req.Username)
			roles = append(roles, "admin")
		}
		user := &User{
			ID:           primitive.NewObjectID(),
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: hash,
			Roles:        roles,
			Bio: fmt.Sprintf(`Welcome to my office! I am %s and %s is what I do.`,
				req.Username, strings.Join(roles, ", ")),
			AvatarURL: "",
			JWTToken:  "", // will be set after signing
			CSRFToken: "", // will be set after signing
			CreatedAt: primitive.NewDateTimeFromTime(time.Now()),
			UpdatedAt: primitive.NewDateTimeFromTime(time.Now()),
		}
		logs.Dev("creating user: %s", user.Username)

		if _, err := service.storage.Collection("users").InsertOne(context.Background(), user); err != nil {
			logs.Err("failed to insert user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		logs.Debug("user created successfully: %+v", user)
		GenAndSetAuthTokens(c, user, service.secret)

		c.JSON(http.StatusCreated, gin.H{
			"status": "ok",
			"error":  err,
			"data": gin.H{
				"success":  true,
				"username": user.Username,
			},
		})
	}
}
