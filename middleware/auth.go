package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	Subject  string `json:"sub"`      // user ID
	Username string `json:"username"` // username for convenience
	// Roles is a list of roles the user has, e.g. ["admin", "user"]
	Roles []string `json:"roles"`
}

// JWTMiddleware validates the JWT in the Authorization header.
func JWTMiddleware(jwtSecret []byte) gin.HandlerFunc {
	log.Printf("JWTMiddleware: running ... ")
	return func(c *gin.Context) {

		tokenString, err := c.Cookie("jwt")
		if err != nil {
			log.Printf("JWTMiddleware: no token found in cookie")

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		log.Printf("JWTMiddleware: cookie token: ...%s", tokenString[len(tokenString)-20:])
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
			log.Printf("JWTMiddleware: parsing token with claims: %v", token.Claims)
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			log.Printf("JWTMiddleware: invalid token: (%v) %v", err, token)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		log.Printf("JWTMiddleware: token is valid, claims: %v", token.Claims)
		claims := token.Claims.(*Claims)
		c.Set("user_id", claims.Subject)
		c.Set("username", claims.RegisteredClaims.Subject)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

// RoleMiddleware ensures the user has at least one required role.
func RoleMiddleware(required ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, _ := c.Get("roles")
		roles, ok := raw.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid roles"})
			return
		}
		for _, want := range required {
			for _, have := range roles {
				if have == want {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "insufficient permissions"})
	}
}
