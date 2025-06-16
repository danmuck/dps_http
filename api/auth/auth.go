package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/danmuck/dps_http/api/logs"
	"github.com/danmuck/dps_http/api/users"
	"github.com/danmuck/dps_http/api/utils"
	"github.com/danmuck/dps_http/storage"
)

// registerPayload is the expected input for user registration.
// It includes username, email, password, and a confirmation field for password matching.
// It is used to validate incoming registration requests.
type registerPayload struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Confirm  string `json:"confirm"  binding:"required,eqfield=Password"` // confirm password must match
}

// loginPayload is the expected input for user login.
// It includes username and password fields, both required for authentication.
// It is used to validate incoming login requests.
type loginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterHandler handles user registration.
// It validates the input, checks for uniqueness of username and email,
// hashes the password, assigns roles, and stores the user in the database.
// It also signs a JWT token for the user and sets it as a secure cookie.
// If successful, it returns a 201 Created response with the username.
// If there are validation errors or uniqueness conflicts, it returns appropriate error responses.
// It uses the provided storage interface to interact with the user data.
// The JWT secret is used to sign the token, and it should be kept secure.
func RegisterHandler(store storage.Client, jwtSecret string) gin.HandlerFunc {
	logs.Log("[api:auth] RegisterHandler() initializing with JWT secret: %s", jwtSecret)
	logs.Log("[api:auth] using storage type: %s", store.Type())
	logs.Log("[api:auth] using storage name: %s", store.Name())
	logs.Log("[api:auth] using storage: %T", store)
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

		hash, err := utils.HashPassword(in.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		roles := []string{"user"}
		if in.Username == "admin" || in.Username == "dirtpig" {
			logs.Log("[api:auth]: assigning admin role to user: %s", in.Username)
			roles = append(roles, "admin")
		}
		user := users.User{
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
		logs.Log("[api:auth]: creating user: %s", user.Username)

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":   user.ID.Hex(),
			"roles": user.Roles,
			"exp":   time.Now().Add(24 * time.Hour).Unix(),
		})
		logs.Log("[api:auth]: signing token for user: %s", user.Username)
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}
		logs.Log("[api:auth]: token signed successfully for user: %s \n  %v", user.Username, tokenString)

		if err := store.Store("users", user.ID.Hex(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		c.SetCookie("jwt", tokenString, 3600*24, "/", "localhost", true, true) // secure cookie
		c.JSON(http.StatusCreated, gin.H{
			"status":   "ok",
			"username": user.Username,
		})
	}
}

// LoginHandler handles user login.
// It validates the input, looks up the user by username,
// checks the password against the stored hash, and signs a JWT token.
func LoginHandler(store storage.Client, jwtSecret string) gin.HandlerFunc {
	logs.Log("[api:auth] LoginHandler() initializing with JWT secret: %s", jwtSecret)
	logs.Log("[api:auth] using storage type: %s", store.Type())
	logs.Log("[api:auth] using storage name: %s", store.Name())
	return func(c *gin.Context) {
		var in loginPayload
		if err := c.ShouldBindJSON(&in); err != nil {
			logs.Log("[api:auth]: bind error: %v", err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// lookup user by username
		logs.Log("[api:auth]: received login request for user: %s", in.Username)
		raw, found := store.Lookup("users", bson.M{"username": in.Username})
		if !found {
			logs.Log("[api:auth]: user not found: %s", in.Username)
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}
		logs.Log("[api:auth]: user found: %s", in.Username)
		var user users.User
		data, _ := bson.Marshal(raw)
		if err := bson.Unmarshal(data, &user); err != nil {
			logs.Log("[api:auth]: unmarshal error: %v", err)
			c.JSON(500, gin.H{"error": "server error"})
			return
		}

		// validate password against stored hash
		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)) != nil {
			logs.Log("[api:auth]: invalid password for user: %s", in.Username)
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
		signed, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			logs.Log("[api:auth]: failed to sign token for user %s: %v", user.Username, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}

		logs.Log("[api:auth]: token signed successfully for user: %s \n  ...%v with hash: %s",
			user.Username, signed[len(signed)-20:], jwtSecret)

		c.SetCookie("jwt", signed, 3600*24, "/", "", true, true)
		c.SetCookie("username", user.Username, 3600*24, "/", "", true, false)
		c.JSON(http.StatusOK, gin.H{
			"username": user.Username,
		})
	}
}

// LogoutHandler clears the JWT cookie to log the user out.
// note: It does not invalidate the JWT on the server side, but removes it from the client.
func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		logs.Log("[api:auth] LogoutHandler() clearing JWT cookie")
		c.SetCookie("jwt", "", -1, "/", "", true, true)
		c.SetCookie("username", "", -1, "/", "", true, false)
		c.JSON(http.StatusOK, gin.H{"status": "logged out"})
	}
}
