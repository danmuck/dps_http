package auth

import (
	"golang.org/x/crypto/bcrypt"
)

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

// // GenerateToken produces a new random token string.
// func GenerateToken(nBytes int) (string, error) {
// 	b := make([]byte, nBytes)
// 	if _, err := rand.Read(b); err != nil {
// 		return "", fmt.Errorf("cannot generate random token: %w", err)
// 	}
// 	// Base64-URL encode with no padding
// 	return base64.RawURLEncoding.EncodeToString(b), nil
// }

// // HashToken returns a SHA-512 digest of the raw token, Base64-URL encoded.
// func HashToken(raw string) string {
// 	sum := sha512.Sum512([]byte(raw))
// 	// Use RawURLEncoding so there’s no “=” padding
// 	return base64.RawURLEncoding.EncodeToString(sum[:])
// }

// // StoreTokenForUser hashes the raw token and saves it in the database.
// // (Replace with your own DB logic.)
// func StoreTokenForUser(userID string, rawToken string) error {
// 	hashed := HashToken(rawToken)
// 	// e.g. UPDATE users SET token_hash = hashed WHERE id = userID
// 	return saveUserTokenHashToDB(userID, hashed)
// }

// // ValidateTokenForUser checks whether the presented raw token matches
// // the stored hash for that user. Returns nil on success.
// func ValidateTokenForUser(userID, presentedToken string) error {
// 	storedHash, err := loadUserTokenHashFromDB(userID)
// 	if err != nil {
// 		return fmt.Errorf("cannot load stored token hash: %w", err)
// 	}
// 	if storedHash != HashToken(presentedToken) {
// 		return errors.New("invalid token")
// 	}
// 	return nil
// }
