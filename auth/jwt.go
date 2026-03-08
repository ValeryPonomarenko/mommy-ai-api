package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const jwtSecret = "mommy-ai-prototype-secret-change-in-production"
const jwtExpiry = 30 * 24 * time.Hour // 30 days

type claims struct {
	UserID string `json:"uid"`
	jwt.RegisteredClaims
}

// SignToken creates a JWT for the user ID.
func SignToken(userID string) (string, error) {
	c := claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(jwtExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return t.SignedString([]byte(jwtSecret))
}

// ValidateToken returns the user ID from a JWT, or error.
func ValidateToken(tokenString string) (string, error) {
	t, err := jwt.ParseWithClaims(tokenString, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return "", err
	}
	c, ok := t.Claims.(*claims)
	if !ok || !t.Valid {
		return "", fmt.Errorf("invalid token")
	}
	return c.UserID, nil
}
