package admin

import (
	"crypto/sha512"
	"net/http"
	"time"

	"github.com/danmuck/dps_http/api/auth"
	api "github.com/danmuck/dps_http/api/v1"
	"github.com/danmuck/dps_http/lib/logs"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func createDummyUser(username string) error {
	email := dummyString(8, "@dirtranch.io")
	password := dummyString(4, "crypt")
	if username == "" || username == "undefined" {
		logs.Log("[DEV]> No username provided, generating a random one")
		username = dummyString(4, "dps")
	}

	if _, found := service.storage.Lookup(service.userDB, bson.M{"username": username}); found {
		logs.Log("[DEV]> User %s already exists, generating a new one", username)
		username = dummyString(4, "dps")
	}

	logs.Log("[DEV]> Creating user: %s", username)

	hash, err := auth.HashPassword(password)
	if err != nil {
		logs.Log("[DEV]> Hashing error: %v", err)
		return err
	}
	t := sha512.Sum512([]byte(username))
	logs.Dev("token: %s", t)

	logs.Log("[DEV]> Hashed password: %s", hash)
	user := api.User{
		ID:           primitive.NewObjectID(),
		Username:     username,
		Email:        email,
		PasswordHash: hash,
		Roles:        []string{"dummy"},
		Token:        "NEW_USER_TOKEN",
		Bio:          "the dev. the creator.. ... ..i'm special",
		AvatarURL:    "banner.svg",
		CreatedAt:    primitive.NewDateTimeFromTime(time.Now()),
		UpdatedAt:    primitive.NewDateTimeFromTime(time.Now()),
	}

	logs.Log("[DEV]> User object: %s", user.String())

	// Store the user in the database
	if err := service.storage.Store(service.userDB, user.ID.Hex(), user); err != nil {
		return err
	}
	return nil
}

// creates a dummy user with random data
// note:
// NEEDS TO BE REDIRECTED TO AUTH SERVICE AND LOCKED BEHIND ROLE
func CreateUser() gin.HandlerFunc {
	logs.Init("[DEV]> CreateUser() initializing with storage: %s", service.storage.Name())
	return func(c *gin.Context) {
		var username string
		username = c.PostForm("username")
		logs.Log("[DEV]> CreateUser: received username: %s", username)
		// email = dummyString(8, "@dirtranch.io")
		// password = dummyString(4, "crypt")

		if username == "" || username == "undefined" {
			logs.Log("[DEV]> No username provided, generating a random one")
			username = dummyString(4, "dps")
		}

		if _, found := service.storage.Lookup(service.userDB, bson.M{"username": username}); found {
			logs.Log("[DEV]> User %s already exists, generating a new one", username)
			username = dummyString(4, "dps")
		}

		if err := createDummyUser(username); err != nil {
			// Store the user in the database
			// if err := service.storage.Store(service.userDB, user.ID.Hex(), user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": "error",
				"error":  "failed to create user",
			})
			return
		}

		c.Redirect(http.StatusFound, "/admin/new")
	}
}
