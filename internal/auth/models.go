package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	DisplayName      string     `json:"display_name"`
	SubscriptionTier string     `json:"subscription_tier"`
	CreatedAt        time.Time  `json:"created_at"`
}

type Claims struct {
	UserID           string `json:"sub"`
	SubscriptionTier string `json:"tier"`
	jwt.RegisteredClaims
}

// Request types
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Response types
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
}
