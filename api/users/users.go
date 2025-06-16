package users

import (
	// "encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/danmuck/dps_http/api/logs"
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

func GetUser(store storage.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		logs.Log("GetUser: getting %s", c.Param("username"))
		key := c.Param("username")

		// retrieve the raw map from storage
		raw, ok := store.Lookup("users", bson.M{"username": key})
		if !ok {
			logs.Log("GetUser: not found: %s", key)
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		logs.Log("GetUser: raw user map: %v", raw["username"])

		rawBSON, _ := bson.Marshal(raw)
		var user User
		if err := bson.Unmarshal(rawBSON, &user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse user"})
			return
		}

		// return the User struct
		logs.Log("GetUser: got user %s", user.string())
		c.JSON(http.StatusOK, user)
	}
}

func UpdateUser(store storage.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		logs.Log("UpdateUser: updating user %s", id)

		// bind only the updatable fields
		var patch struct {
			Email     string   `json:"email,omitempty"`
			Bio       string   `json:"bio,omitempty"`
			AvatarURL string   `json:"avatarURL,omitempty"`
			Roles     []string `json:"roles,omitempty"` // if you want to allow role changes
		}
		if err := c.ShouldBindJSON(&patch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
			return
		}

		// build a map of just the changed fields
		updates := map[string]any{}
		if patch.Email != "" {
			logs.Log("UpdateUser: patching email to %s", patch.Email)
			updates["email"] = patch.Email
		}
		if patch.Bio != "" {
			logs.Log("UpdateUser: patching bio to %s", patch.Bio)
			updates["bio"] = patch.Bio
		}
		if patch.AvatarURL != "" {
			logs.Log("UpdateUser: patching avatarURL to %s", patch.AvatarURL)
			updates["avatarURL"] = patch.AvatarURL
		}
		if len(patch.Roles) > 0 {
			logs.Log("UpdateUser: patching roles to %v", patch.Roles)
			updates["roles"] = patch.Roles
		}
		if len(updates) == 0 {
			logs.Log("UpdateUser: nothing to update for user %s", id)
			c.JSON(http.StatusBadRequest, gin.H{"error": "nothing to update"})
			return
		}

		// apply the patch
		if err := store.Patch("users", id, updates); err != nil {
			logs.Log("UpdateUser: failed to update user %s: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update"})
			return
		}

		// return the updated document (you can re‐fetch with Retrieve)
		logs.Log("UpdateUser: retreiving updated user %s", id)
		updated, _ := store.Retrieve("users", id)
		logs.Log("UpdateUser: updated user %s: %v", id, updated)
		c.JSON(http.StatusOK, updated)
	}
}

func DeleteUser(store storage.Client) gin.HandlerFunc {
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

func ListUsers(store storage.Client) gin.HandlerFunc {
	logs.Log("> Listing users from store: %s", store.Name())
	return func(c *gin.Context) {
		users, err := store.List("users")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to retrieve users",
			})
			return
		}

		logs.Log("> Retrieved %d users", len(users))
		c.JSON(http.StatusOK, users)
	}
}

// creates a dummy user with random data
// NOTE:
// NEEDS TO BE REDIRECTED TO AUTH SERVICE AND LOCKED BEHIND ROLE
func CreateUser(store storage.Client) gin.HandlerFunc {
	logs.Log("[DEV]> CreateUser() initializing with storage: %s", store.Name())
	return func(c *gin.Context) {
		var email, username, password string
		username = c.PostForm("username")
		logs.Log("[DEV]> CreateUser: received username: %s", username)
		email = dummyString(8, "@dirtranch.io")
		password = dummyString(4, "crypt")

		if username == "" || username == "undefined" {
			logs.Log("[DEV]> No username provided, generating a random one")
			username = dummyString(4, "dps")
		}
		store.Lookup("users", bson.M{"username": username})
		if _, found := store.Lookup("users", bson.M{"username": username}); found {
			logs.Log("[DEV]> User %s already exists, generating a new one", username)
			username = dummyString(4, "dps")
		}

		logs.Log("[DEV]> Creating user: %s", username)

		hash, err := utils.HashPassword(password)
		if err != nil {
			logs.Log("[DEV]> Hashing error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to hash password",
			})
			return
		}

		logs.Log("[DEV]> Hashed password: %s", hash)
		user := User{
			ID:           primitive.NewObjectID(),
			Username:     username,
			Email:        email,
			PasswordHash: hash,
			Roles:        []string{"dummy"},
			Token:        "DEV_TOKEN",
			Bio:          "the dev. the creator.. ... ..i'm special",
			AvatarURL:    "/dm_logo.svg",
			CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
			UpdatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		}

		logs.Log("[DEV]> User object: %s", user.string())

		// Store the user in the database
		if err := store.Store("users", user.ID.Hex(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to create user",
			})
			return
		}

		c.Redirect(http.StatusFound, "/users/new")
	}
}

func dummyString(length int, postfix string) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ01"
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return fmt.Sprintf("%s%s", string(b), postfix)
}
