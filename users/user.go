package users

import (
	"context"
	"fmt"
	"time"

	"github.com/danmuck/dps_http/mongo_client"
	"github.com/danmuck/dps_lib/logs"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var MongoClient *mongo_client.MongoClient = nil
var Database = "users"                       // default database endpoint, can be overridden
var DefaultURI = "mongodb://localhost:27017" // Default MongoDB URI, can be overridden

// must run config if you would like to override the default MongoDB URI or database
func ConfigureT(client *mongo_client.MongoClient, endpoint string) {
	if client == nil {
		logs.Err("MongoClient is nil, initializing with default URI and database")
		var err error
		client, err = mongo_client.NewMongoStore(
			DefaultURI,
			Database,
		)
		if err != nil {
			panic(fmt.Sprintf("Failed to connect to MongoDB: %v", err))
		}
	}
	MongoClient = client
	Database = endpoint
}

// User represents an authenticated user
type User struct {
	ID           primitive.ObjectID `bson:"_id" json:"id"`
	Username     string             `bson:"username" json:"username"`
	Email        string             `bson:"email" json:"email"`
	PasswordHash string             `bson:"password" json:"-"`
	JWTToken     string             `bson:"token" json:"-"`
	CSRFToken    string             `bson:"csrf_token" json:"-"` // CSRF token for additional security; empty atm
	Roles        []string           `bson:"roles" json:"roles"`

	Bio       string             `bson:"bio,omitempty" json:"bio,omitempty"`
	AvatarURL string             `bson:"avatar_url,omitempty" json:"avatar_url,omitempty"`
	CreatedAt primitive.DateTime `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at,omitempty"`

	ContactInfo ContactInfo   `bson:"contact_info,omitempty" json:"contact_info,omitempty"`
	Projects    []ProjectInfo `bson:"projects,omitempty" json:"projects,omitempty"`
	ProfileInfo ProfileInfo   `bson:"profile_info,omitempty" json:"profile_info,omitempty"`
	CareerInfo  CareerInfo    `bson:"career_info,omitempty" json:"career_info"`
}

func (u *User) String() string {
	var token string = "[no token]"
	if len(u.JWTToken) > 20 {
		token = u.JWTToken[:10] + " ... " + u.JWTToken[len(u.JWTToken)-10:]
	}
	return fmt.Sprintf(`
	ID: %s
	User: %s
	Password: %s
	Email: %s
	Roles: %v
	Token: [%s]

	CreatedAt: %s
	UpdatedAt: %s
	Bio: %s
	AvatarURL: %s

	ContactInfo: %s
	Projects: %+v
	ProfileInfo: %s
	CareerInfo: %s
`, u.ID.Hex(),
		u.Username, u.PasswordHash, u.Email, u.Roles, token,
		u.CreatedAt.Time(),
		u.UpdatedAt.Time(), u.Bio, u.AvatarURL,
		u.ContactInfo.String(),
		u.Projects,
		u.ProfileInfo.String(),
		u.CareerInfo.String())

}

func (u *User) UpdateJWTToken(token string) {
	u.JWTToken = token
	u.UpdatedAt = primitive.NewDateTimeFromTime(time.Now())
	// Update the user in the database
	_, err := service.storage.Collection(Database).UpdateOne(
		context.Background(),
		bson.M{"id": u.ID},
		bson.M{"$set": bson.M{"token": u.JWTToken, "updated_at": u.UpdatedAt}},
	)
	if err != nil {
		fmt.Printf("Failed to update user token: %v\n", err)
	}
}
