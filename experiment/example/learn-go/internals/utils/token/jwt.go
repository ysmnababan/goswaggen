package token

import (
	"fmt"
	"learn-go/config"
	"time"

	"github.com/golang-jwt/jwt"
)

// GenerateJWT generates a JWT token with a given UUID and expiration time.
func GenerateJWT(userID string, email string) (string, error) {
	jwtKey := config.Get().JWT.Key
	// Define token claims
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
	}

	// Create a new token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with the secret key
	signedToken, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}
