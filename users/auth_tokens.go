package users

import (
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"github.com/danmuck/dps_lib/logs"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

func SetAuthCookies(c *gin.Context, jwtToken string, csrfToken string, username string) {
	host := strings.Split(c.Request.Host, ":")[0] // strips port
	if jwtToken != "" {
		c.SetCookie("jwt", jwtToken, 3600*24, "/", host, false, true)
	}
	if csrfToken != "" {
		c.SetCookie("csrfToken", csrfToken, 3600*24, "/", host, false, false)
	}
	c.SetCookie("username", username, 3600*24, "/", host, false, false)
}

func GenAndSetAuthTokens(c *gin.Context, user *User, secret string) (string, string, error) {
	jwtToken, err := GenAndSetJWT(c, user, secret)
	if err != nil {
		return "", "", err
	}

	csrfToken, err := GenerateCSRFToken()
	if err != nil {
		return "", "", err
	}
	user.JWTToken = jwtToken
	user.CSRFToken = csrfToken
	SetAuthCookies(c, jwtToken, csrfToken, user.Username)

	return jwtToken, csrfToken, nil
}

type CustomClaims struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func GenAndSetJWT(c *gin.Context, user *User, secret string) (string, error) {
	claims := CustomClaims{
		Username: user.Username,
		Roles:    user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.Username,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	logs.Debug("Generated JWT for user %s: %s", user.Username, signed)
	user.JWTToken = signed // Update user JWT token
	SetAuthCookies(c, signed, user.CSRFToken, user.Username)
	return signed, nil
}

func GenerateCSRFToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
