package middleware

import (
	"net/http"
	"slices"

	"github.com/danmuck/dps_http/lib/logs"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

type Claims struct {
	jwt.RegisteredClaims
	Subject  string `json:"sub"`      // user ID
	Username string `json:"username"` // username for convenience
	// roles is a list of roles the user has, e.g. ["admin", "user"]
	Roles []string `json:"roles"`
}

var JWT_SECRET []byte = []byte("your_jwt_secret_here")

func InitAuthMiddleware(jwtSecret string) error {
	JWT_SECRET = []byte(jwtSecret)
	return nil
}

// JWTMiddleware validates the JWT in the Authorization header.
// secret is not passed elsewhere
func JWTMiddleware() gin.HandlerFunc {
	logs.Init("JWTMiddleware")
	return func(c *gin.Context) {

		tokenString, err := c.Cookie("jwt")
		if err != nil {
			logs.Err("no token found in cookie")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}
		logs.Debug("cookie token: ...%s", tokenString[len(tokenString)-20:])
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
			logs.Debug("parsing token with claims: %v", token.Claims)
			return JWT_SECRET, nil
		})
		if err != nil || !token.Valid {
			logs.Err("invalid token: (%v) %v", err, token)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		logs.Debug("token is valid, claims: [%v]", token.Claims)
		claims := token.Claims.(*Claims)
		logs.Debug("SUBJECT: %s USER: %s", claims.Subject, claims.Username)

		c.Set("username", claims.Username)
		c.Set("user_id", claims.Subject)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

func AuthorizeResourceAccess() gin.HandlerFunc {
	return func(c *gin.Context) {

		// parse & verify the JWT from the cookie/header
		tokenStr, err := c.Cookie("jwt")
		if err != nil {
			logs.Err("missing token, not authorized")
			c.AbortWithStatusJSON(401, gin.H{"error": "missing token, not authorized"})
			return
		}
		token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
			return JWT_SECRET, nil
		})
		if err != nil || !token.Valid {
			logs.Err("invalid token, not authorized")
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token, not authorized"})
			return
		}

		// extract username from claims to match url param
		claims := token.Claims.(*Claims)
		username := claims.Username
		owner := c.Param("username")
		owner_id := c.Param("id")

		raw, _ := c.Get("roles")
		have, ok := raw.([]string)
		if !ok {
			logs.Err("roles on ctx are not a string slice, aborting")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "roles invalid, try logging in and out"})
			return
		}

		// @NOTE -- not sure how secure this is, need to clean it up, i think there is a lot of unnecessary
		// variables flying around in its current state
		if (claims.Username != owner && claims.Subject != owner_id) && !CheckForRole("admin", have...) {
			// logs.Dev("cuser: %s owner: %s csub: %s owner_id%s", claims.Username, owner, claims.Subject, owner_id)
			logs.Err("Not authorized: user: %s owner: %s owner_id: %s", username, owner, owner_id)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"auth": "not owner, not authorized"})
			return
		}
		logs.Debug("Authorized: user: %s owner: %s owner_id: %s", username, owner, owner_id)
		// otherwise store it in context
		c.Set("username", owner)
		c.Set("user_id", claims.Subject)
		c.Set("roles", claims.Roles)
		c.Next()
	}
}

func CheckForRole(want string, have ...string) bool {
	if have == nil {
		logs.Err("CheckRole role is empty")
		return false
	}
	if slices.Contains(have, want) {
		return true
	}
	return false
}

// AuthorizeByRoles ensures the user has **at least** one required role.
func AuthorizeByRoles(required ...string) gin.HandlerFunc {
	logs.Init("RoleMiddleware required: %v", required)
	return func(c *gin.Context) {
		raw, _ := c.Get("roles")
		have, ok := raw.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid roles"})
			return
		}
		for _, want := range required {
			if CheckForRole(want, have...) {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "insufficient permissions"})
	}
}
