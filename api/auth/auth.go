package auth

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/danmuck/dps_http/api/users"
	"github.com/danmuck/dps_http/api/utils"
	"github.com/danmuck/dps_http/storage"
)

type registerPayload struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Confirm  string `json:"confirm"  binding:"required,eqfield=Password"` // confirm password must match
}

type loginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func RegisterHandler(store storage.Storage, jwtSecret string) gin.HandlerFunc {
	log.Printf("RegisterHandler: initializing with JWT secret: %s", jwtSecret)
	log.Printf("RegisterHandler: using storage type: %s", store.Type())
	log.Printf("RegisterHandler: using storage name: %s", store.Name())
	log.Printf("RegisterHandler: using storage: %T", store)
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

		user := users.User{
			ID:           primitive.NewObjectID(),
			Username:     in.Username,
			Email:        in.Email,
			PasswordHash: hash,
			Roles:        []string{"user"},
			Bio:          "Welcome to my office!",
			AvatarURL:    "",
			CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
			UpdatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		}
		log.Printf("RegisterHandler: creating user: %s", user.Username)
		// sign jwt
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":   user.ID.Hex(),
			"roles": user.Roles,
			"exp":   time.Now().Add(24 * time.Hour).Unix(),
		})
		log.Printf("RegisterHandler: signing token for user: %s", user.Username)
		tokenString, err := token.SignedString([]byte(jwtSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}
		log.Printf("RegisterHandler: token signed successfully for user: %s \n  %v", user.Username, tokenString)
		// store under username
		// user.Token = tokenString
		if err := store.Store("users", user.ID.Hex(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		c.SetCookie("jwt", tokenString, 3600*24, "/", "localhost", true, true) // secure cookie
		c.JSON(http.StatusCreated, gin.H{
			"status":   "ok",
			"username": user.Username,
			// "token":    tokenString,
			// no longer returning token in response
			// it is set as a secure cookie
		})
	}
}

func LoginHandler(store storage.Storage, jwtSecret string) gin.HandlerFunc {
	log.Printf("LoginHandler: initializing with JWT secret: %s", jwtSecret)
	log.Printf("LoginHandler: using storage type: %s", store.Type())
	log.Printf("LoginHandler: using storage name: %s", store.Name())
	return func(c *gin.Context) {
		var in loginPayload
		if err := c.ShouldBindJSON(&in); err != nil {
			log.Printf("LoginHandler: bind error: %v", err)
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		// lookup user by username
		log.Printf("LoginHandler: received login request for user: %s", in.Username)
		raw, found := store.Lookup("users", bson.M{"username": in.Username})
		if !found {
			log.Printf("LoginHandler: user not found: %s", in.Username)
			c.JSON(401, gin.H{"error": "invalid credentials"})
			return
		}
		log.Printf("LoginHandler: user found: %s", in.Username)
		var user users.User
		data, _ := bson.Marshal(raw)
		if err := bson.Unmarshal(data, &user); err != nil {
			log.Printf("LoginHandler: unmarshal error: %v", err)
			c.JSON(500, gin.H{"error": "server error"})
			return
		}

		// validate password against stored hash
		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password)) != nil {
			log.Printf("LoginHandler: invalid password for user: %s", in.Username)
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
			log.Printf("LoginHandler: failed to sign token for user %s: %v", user.Username, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}

		log.Printf("LoginHandler: token signed successfully for user: %s \n  ...%v with hash: %s",
			user.Username, signed[len(signed)-20:], jwtSecret)

		c.SetCookie("jwt", signed, 3600*24, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{
			"username": user.Username,
		})
	}
}

func LogoutHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("LogoutHandler: clearing JWT cookie")
		c.SetCookie("jwt", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"status": "logged out"})
	}
}
