// backend/internal/services/authentication/utils.go

package authentication

import (
	"errors"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes the plaintext password using bcrypt.
// It returns the hashed password or an error if hashing fails.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares a plaintext password with its hashed version.
// Returns true if they match, false otherwise.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Claims defines the structure of JWT claims.
// It includes the user ID and standard JWT claims.
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a JWT token for a given user ID.
// Parameters:
// - userID: The unique identifier of the user.
// - secret: The secret key used to sign the token.
// - expiry: The duration after which the token expires.
// Returns:
// - A signed JWT token string.
// - An error if token generation fails.
func GenerateJWT(userID string, secret string, expiry time.Duration) (string, error) {
	expirationTime := time.Now().Add(expiry)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token string and returns the claims if valid.
// Parameters:
// - tokenString: The JWT token string to validate.
// - secret: The secret key used to sign the token.
// Returns:
// - The Claims extracted from the token.
// - An error if validation fails.
func ValidateJWT(tokenString string, secret string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Ensure the token's signing method is HMAC.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
