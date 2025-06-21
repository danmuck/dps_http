package v1

import (
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ContactInfo struct {
	Email    string `json:"email" bson:"email"`
	Phone    string `json:"phone,omitempty" bson:"phone,omitempty"`
	Location string `json:"location,omitempty" bson:"location,omitempty"`
	Website  string `json:"website,omitempty" bson:"website,omitempty"`
	LinkedIn string `json:"linkedin,omitempty" bson:"linkedin,omitempty"`
	GitHub   string `json:"github,omitempty" bson:"github,omitempty"`
}

type ProjectInfo struct {
	Name        string   `json:"name" bson:"name"`
	Description string   `json:"description,omitempty" bson:"description,omitempty"`
	Highlights  string   `json:"highlights,omitempty" bson:"highlights,omitempty"`
	Role        string   `json:"role,omitempty" bson:"role,omitempty"`
	URL         string   `json:"url,omitempty" bson:"url,omitempty"`
	TechStack   []string `json:"tech_stack,omitempty" bson:"tech_stack,omitempty"`
}

type AdminInfo struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Username     string             `bson:"username" json:"username"`
	PasswordHash string             `bson:"password_hash" json:"-"`
	Token        string             `bson:"token" json:"-"`
	Roles        []string           `bson:"roles" json:"roles"`
}

type ProfileInfo struct {
	Bio       string             `json:"bio,omitempty" bson:"bio,omitempty"`
	AvatarURL string             `json:"avatar_url,omitempty" bson:"avatar_url,omitempty"`
	CreatedAt primitive.DateTime `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt primitive.DateTime `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}

type CareerInfo struct {
	CurrentRole string   `json:"current_role,omitempty" bson:"current_role,omitempty"`
	Skills      []string `json:"skills,omitempty" bson:"skills,omitempty"`
	Experience  string   `json:"experience,omitempty" bson:"experience,omitempty"`
	Education   string   `json:"education,omitempty" bson:"education,omitempty"`
	Interests   []string `json:"interests,omitempty" bson:"interests,omitempty"`
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

func (u *User) String() string {
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
