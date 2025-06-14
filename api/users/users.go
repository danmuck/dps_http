package users

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/danmuck/dps_http/api/utils"
	"github.com/danmuck/dps_http/storage"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Claims defines the JWT payload
type Claims struct {
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

// User represents an authenticated user
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Username     string             `bson:"username" json:"username"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	Token        string             `bson:"token" json:"-"`
	Roles        []string           `bson:"roles" json:"roles"`

	Bio       string             `bson:"bio,omitempty" json:"bio,omitempty"`
	AvatarURL string             `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	CreatedAt primitive.DateTime `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

func (u *User) string() string {
	var token string = "[no token]"
	if len(u.Token) > 20 {
		token = u.Token[:10] + " ... " + u.Token[len(u.Token)-10:]
	}
	return fmt.Sprintf(`
	User: %s
	Password: %s
	Email: %s
	Roles: %v
	Token: [%s]

	CreatedAt: %s
	UpdatedAt: %s
	Bio: %s
	AvatarURL: %s
`,
		u.Username, u.PasswordHash, u.Email, u.Roles, token,
		u.CreatedAt.Time(),
		u.UpdatedAt.Time(), u.Bio, u.AvatarURL)

}

type createUserPayload struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Confirm  string `json:"confirm" binding:"required,eqfield=Password"` // confirm password must match
	// Add more fields as needed
	// Roles    []string `json:"roles" binding:"required"`
}

func GetUser(store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("GetUser: getting user by username: %s", c.Param("username"))
		key := c.Param("username") // you’re using username as the key now

		// retrieve the raw map from storage
		raw, ok := store.Lookup("users", bson.M{"username": key})
		if !ok {
			log.Printf("GetUser: user not found: %s", key)
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Printf("GetUser: getting raw user map: %v", raw["username"])

		rawBSON, _ := bson.Marshal(raw)
		var user User
		if err := bson.Unmarshal(rawBSON, &user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user"})
			return
		}

		// return the User struct
		log.Printf("GetUser: got user: %s", user.string())
		c.JSON(http.StatusOK, user)
	}
}

type updateUserPayload struct {
	Email     *string `json:"email,omitempty"`
	Bio       *string `json:"bio,omitempty"`
	AvatarURL *string `json:"avatarURL,omitempty"`
}

func UpdateUser(store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Param("id")
		log.Printf("UpdateUser: updating user by username: %s", username)
		var patch updateUserPayload
		if err := c.ShouldBindJSON(&patch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		// retrieve the exact document by key
		rawValue, ok := store.Lookup("users", bson.M{"username": username})
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		log.Printf("UpdateUser: retrieved raw user map: %v", rawValue)
		// note: there must be a better way to do this
		// convert the rawValue (any) into User struct
		js, err := json.Marshal(rawValue)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "marshal failure"})
			return
		}
		var existing User
		if err := json.Unmarshal(js, &existing); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unmarshal failure"})
			return
		}
		log.Printf("UpdateUser: existing user: %s %v", existing.string(), existing.ID.Hex())

		if patch.Email != nil {
			existing.Email = *patch.Email
		}
		if patch.Bio != nil {
			existing.Bio = *patch.Bio
		}
		if patch.AvatarURL != nil {
			existing.AvatarURL = *patch.AvatarURL
		}
		existing.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())

		// write the full updated user object back under the same key
		if err := store.Update("users", existing.ID.Hex(), existing); err != nil {
			log.Printf("UpdateUser: failed to update user %s: %v \n  %v", username, err, existing)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
			return
		}

		c.JSON(http.StatusOK, existing)
	}
}

func DeleteUser(store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := store.Delete("users", c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to delete user",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"message": "user deleted successfully",
		})
	}
}

func ListUsers(store storage.Storage) gin.HandlerFunc {
	log.Printf("> Listing users from store: %s", store.Name())
	return func(c *gin.Context) {
		users, err := store.List("users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to retrieve users",
			})
			return
		}

		log.Printf("> Retrieved %d users", len(users))
		c.JSON(http.StatusOK, users)
	}
}

// NOTE:
// NEEDS TO BE REDIRECTED TO AUTH SERVICE AND LOCKED BEHIND ROLE
func CreateUser(store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload createUserPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			log.Printf("> bind error: %v", err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		if _, exists := store.Lookup("users", bson.M{"username": payload.Username}); exists {
			c.JSON(http.StatusConflict, gin.H{"error": "username already in use"})
			return
		}
		if _, exists := store.Lookup("users", bson.M{"email": payload.Email}); exists {
			c.JSON(http.StatusConflict, gin.H{"error": "email already in use"})
			return
		}
		log.Printf("> Creating user: %s", payload.Username)
		//
		// hash password and prepare user object
		if payload.Password != payload.Confirm {
			log.Printf("> Passwords do not match")
			c.JSON(http.StatusBadRequest, gin.H{
				"status": "error",
				"error":  "passwords do not match",
			})
			return
		}

		hash, err := utils.HashPassword(payload.Password)
		if err != nil {
			log.Printf("> Hashing error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to hash password",
			})
			return
		}
		log.Printf("> Hashed password: %s", hash)
		user := User{
			ID:           primitive.NewObjectID(),
			Username:     payload.Username,
			Email:        payload.Email,
			PasswordHash: hash,
			Roles:        []string{"user"},                          // default
			Token:        "REPLACE_WITH_JWT_TOKEN",                  // placeholder
			Bio:          "introductory things and stuff",           // optional
			AvatarURL:    "",                                        // optional
			CreatedAt:    primitive.NewDateTimeFromTime(time.Now()), // assuming you set this in middleware
			UpdatedAt:    primitive.NewDateTimeFromTime(time.Now()), // assuming you set this in middleware
		}

		log.Printf("> User object: %s", user.string())

		// Store the user in the database
		if err := store.Store("users", user.ID.Hex(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to create user",
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"status": "ok",
			"user":   user,
		})
	}
}
