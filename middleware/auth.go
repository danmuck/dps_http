package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	UserID string
	Roles  string
}

func (c *Claims) Valid() error {
	return fmt.Errorf("not implemented, filler struct implements jwt.Claims")
}

// Secret used to sign JWTs
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// JWTMiddleware validates the JWT in the Authorization header.
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		tkn := strings.TrimPrefix(auth, "Bearer ")

		token, err := jwt.ParseWithClaims(tkn, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims := token.Claims.(*Claims)
		c.Set("user_id", claims.UserID)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

// RoleMiddleware ensures the user has at least one required role.
func RoleMiddleware(required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		r, _ := c.Get("roles")
		roles := r.([]string)

		for _, req := range required {
			for _, have := range roles {
				if have == req {
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}
