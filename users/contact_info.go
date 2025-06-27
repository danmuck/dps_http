package users

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

func (c *ContactInfo) String() string {
	return fmt.Sprintf(`
	Email: %s
	Phone: %s
	Location: %s
	Website: %s
	LinkedIn: %s
	GitHub: %s`,
		c.Email, c.Phone, c.Location, c.Website, c.LinkedIn, c.GitHub)
}
func (p *ProjectInfo) String() string {
	return fmt.Sprintf(`
	Name: %s
	Description: %s
	Highlights: %s
	Role: %s
	URL: %s
	TechStack: %v`,
		p.Name, p.Description, p.Highlights, p.Role, p.URL, p.TechStack)
}
func (p *ProfileInfo) String() string {
	return fmt.Sprintf(`
	Bio: %s
	AvatarURL: %s
	CreatedAt: %s
	UpdatedAt: %s`,
		p.Bio, p.AvatarURL, p.CreatedAt.Time(), p.UpdatedAt.Time())
}
func (c *CareerInfo) String() string {
	return fmt.Sprintf(`
	CurrentRole: %s
	Skills: %v
	Experience: %s
	Education: %s
	Interests: %v`,
		c.CurrentRole, c.Skills, c.Experience, c.Education, c.Interests)
}
