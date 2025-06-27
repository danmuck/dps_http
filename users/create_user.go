package users

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/danmuck/dps_http/mongo_client"
	"github.com/danmuck/dps_lib/logs"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	Username string   `json:"username" binding:"required"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=8"`
	Bio      string   `json:"bio"`
	Roles    []string `json:"roles"`
}

func CreateUser(client *mongo_client.MongoClient) gin.HandlerFunc {
	logs.Dev("CreateOne from storage: %s", client.Name())
	return func(c *gin.Context) {
		logs.Dev("CreateOne called with body: %v", c.Request.Body)
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			logs.Err("Failed to bind JSON: %v", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
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
			Bio:          "I am a bot.",
			AvatarURL:    "",
			JWTToken:     "", // will be set after signing
			CSRFToken:    "", // will be set after signing
			CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
			UpdatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		}
		logs.Dev("creating user: %s", user.Username)

		if _, err := service.storage.Collection("users").InsertOne(context.Background(), user); err != nil {
			logs.Err("failed to insert user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		logs.Dev("user created successfully: %+v", user)
		c.JSON(http.StatusCreated, gin.H{"message": "user created successfully"})
	}
}
