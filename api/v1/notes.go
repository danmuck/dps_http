package v1

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Note interface {
	// Category() string
	// Content() string
	String() string
	Note() *note
}
type note struct {
	ID       primitive.ObjectID `bson:"id" json:"id"`
	Category string             `bson:"category" json:"category"`
	Content  string             `bson:"content" json:"content"`
	Date     primitive.DateTime `bson:"date" json:"date"`
}

type QuickTextNote struct {
	Title string `bson:"title" json:"title"`
	Body  string `bson:"body" json:"body"`
	Topic string `bson:"topic" json:"topic"`
}

type LinkNote struct {
	Title string `bson:"title" json:"title"`
	URL   string `bson:"url" json:"url"`
	Topic string `bson:"topic" json:"topic"`
}

type BlogNote struct {
	Title string `bson:"title" json:"title"`
	Topic string `bson:"topic" json:"topic"`
	Body  string `bson:"body" json:"body"`
}

func (qt *QuickTextNote) String() string {
	return fmt.Sprintf("Title: %s, Body: %s", qt.Title, qt.Body)
}
func (qt *QuickTextNote) Note() *note {
	return &note{
		Category: qt.Title,
		Content:  qt.Body,
	}
}

// new quick note handlerFunc
func NewQuickTextNoteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req QuickTextNote
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, req.Note())
	}
}

func (l *LinkNote) String() string {
	return fmt.Sprintf("Title: %s, URL: %s", l.Title, l.URL)
}
func (l *LinkNote) Note() *note {
	return &note{
		Category: l.Topic,
		Content:  l.String(),
	}
}

func (b *BlogNote) String() string {
	return fmt.Sprintf("Title: %s, Topic: %s, Body: %s", b.Title, b.Topic, b.Body)
}
func (b *BlogNote) Note() *note {
	return &note{
		Category: b.Topic,
		Content:  b.String(),
	}
}
