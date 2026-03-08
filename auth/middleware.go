package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// ContextUserIDKey is the key for user ID in gin context.
const ContextUserIDKey = "userID"

// RequireAuth validates Bearer JWT and sets user ID in context. Returns 401 if missing/invalid.
func RequireAuth(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "требуется авторизация"})
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "неверный формат токена"})
			return
		}
		userID, err := ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "недействительный токен"})
			return
		}
		if _, err := store.UserByID(userID); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не найден"})
			return
		}
		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}

// OptionalAuth sets user ID in context when Bearer JWT is valid. Does not abort when missing/invalid.
func OptionalAuth(store *Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.Next()
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}
		userID, err := ValidateToken(parts[1])
		if err != nil {
			c.Next()
			return
		}
		if _, err := store.UserByID(userID); err != nil {
			c.Next()
			return
		}
		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}
