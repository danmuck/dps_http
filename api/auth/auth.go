package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// registerPayload defines the input for user registration.
type registerPayload struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Confirm  string `json:"confirm"  binding:"required,eqfield=Password"` // confirm password must match
}

// loginPayload represents the input for user login.
type loginPayload struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func VerifyPassword(hashed, password string) bool {
	check := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	if check != nil {
		return false
	}
	return true
}
