package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"mommy-ai-api/auth"
)

// RegisterRequest body.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest body.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse returned on register/login.
type AuthResponse struct {
	Token string     `json:"token"`
	User  UserResponse `json:"user"`
}

// UserResponse public user info.
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// OnboardingRequest body for PUT /api/me/onboarding.
type OnboardingRequest struct {
	PregnancyWeek int `json:"pregnancy_week" binding:"required,min=1,max=42"`
	Feelings      int `json:"feelings" binding:"required,min=1,max=5"`
}

// MeResponse for GET /api/me (user + profile).
type MeResponse struct {
	User    UserResponse    `json:"user"`
	Profile *ProfileResponse `json:"profile,omitempty"`
}

// ProfileResponse onboarding data.
type ProfileResponse struct {
	PregnancyWeek int `json:"pregnancy_week"`
	Feelings     int `json:"feelings"`
}

// Register creates a user and returns JWT.
func Register(store *auth.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверные данные: email и пароль (мин. 6 символов)"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка сервера"})
			return
		}
		u, err := store.CreateUser(req.Email, string(hash))
		if err != nil {
			if err == auth.ErrEmailExists {
				c.JSON(http.StatusConflict, gin.H{"error": "такой email уже зарегистрирован"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		token, err := auth.SignToken(u.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка сервера"})
			return
		}
		c.JSON(http.StatusOK, AuthResponse{
			Token: token,
			User:  UserResponse{ID: u.ID, Email: u.Email},
		})
	}
}

// Login checks password and returns JWT.
func Login(store *auth.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "укажите email и пароль"})
			return
		}
		u, err := store.UserByEmail(req.Email)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный email или пароль"})
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "неверный email или пароль"})
			return
		}
		token, err := auth.SignToken(u.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка сервера"})
			return
		}
		c.JSON(http.StatusOK, AuthResponse{
			Token: token,
			User:  UserResponse{ID: u.ID, Email: u.Email},
		})
	}
}

// GetMe returns current user and profile (if onboarding done).
func GetMe(store *auth.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get(auth.ContextUserIDKey)
		id := userID.(string)
		u, err := store.UserByID(id)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "пользователь не найден"})
			return
		}
		res := MeResponse{User: UserResponse{ID: u.ID, Email: u.Email}}
		if p := store.GetProfile(id); p != nil {
			res.Profile = &ProfileResponse{PregnancyWeek: p.PregnancyWeek, Feelings: p.Feelings}
		}
		c.JSON(http.StatusOK, res)
	}
}

// PutOnboarding saves pregnancy week and feelings.
func PutOnboarding(store *auth.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get(auth.ContextUserIDKey)
		id := userID.(string)
		var req OnboardingRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неделя беременности 1–42, самочувствие 1–5"})
			return
		}
		store.SetProfile(id, auth.Profile{
			PregnancyWeek: req.PregnancyWeek,
			Feelings:      req.Feelings,
		})
		c.JSON(http.StatusOK, ProfileResponse{
			PregnancyWeek: req.PregnancyWeek,
			Feelings:      req.Feelings,
		})
	}
}
