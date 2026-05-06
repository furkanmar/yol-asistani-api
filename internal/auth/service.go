package auth

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/furkan/yol-asistani-api/config"
)

var (
	ErrEmailTaken    = errors.New("email already in use")
	ErrInvalidCreds  = errors.New("invalid email or password")
	ErrInvalidToken  = errors.New("invalid or expired token")
)

type Service struct {
	db  *pgxpool.Pool
	cfg *config.Config
}

func NewService(db *pgxpool.Pool, cfg *config.Config) *Service {
	return &Service{db: db, cfg: cfg}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	// Email kontrolü
	var exists bool
	s.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, req.Email).Scan(&exists)
	if exists {
		return nil, ErrEmailTaken
	}

	// Hash
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Insert
	var user User
	err = s.db.QueryRow(ctx, `
		INSERT INTO users(email, password_hash, display_name)
		VALUES($1, $2, $3)
		RETURNING id, email, display_name, subscription_tier, created_at
	`, req.Email, string(hash), req.DisplayName).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.SubscriptionTier, &user.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	var user User
	var passwordHash string

	err := s.db.QueryRow(ctx, `
		SELECT id, email, display_name, subscription_tier, created_at, password_hash
		FROM users WHERE email=$1
	`, req.Email).Scan(
		&user.ID, &user.Email, &user.DisplayName, &user.SubscriptionTier, &user.CreatedAt,
		&passwordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidCreds
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCreds
	}

	return s.issueTokens(ctx, user)
}

func (s *Service) Refresh(ctx context.Context, rawRefreshToken string) (*RefreshResponse, error) {
	// JWT doğrula
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(rawRefreshToken, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(s.cfg.JWTRefreshSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	// DB'de token hash var mı kontrol et
	hash := hashToken(rawRefreshToken)
	var userID string
	err = s.db.QueryRow(ctx, `
		SELECT user_id FROM refresh_tokens
		WHERE token_hash=$1 AND expires_at > NOW()
	`, hash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("query refresh token: %w", err)
	}

	// Güncel tier'ı al
	var tier string
	s.db.QueryRow(ctx, `SELECT subscription_tier FROM users WHERE id=$1`, userID).Scan(&tier)

	// Yeni access token
	accessToken, err := s.generateAccessToken(userID, tier)
	if err != nil {
		return nil, err
	}

	return &RefreshResponse{AccessToken: accessToken}, nil
}

func (s *Service) Logout(ctx context.Context, rawRefreshToken string) error {
	hash := hashToken(rawRefreshToken)
	_, err := s.db.Exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash=$1`, hash)
	return err
}

func (s *Service) Me(ctx context.Context, userID string) (*User, error) {
	var user User
	err := s.db.QueryRow(ctx, `
		SELECT id, email, display_name, subscription_tier, created_at
		FROM users WHERE id=$1
	`, userID).Scan(&user.ID, &user.Email, &user.DisplayName, &user.SubscriptionTier, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, fmt.Errorf("query user: %w", err)
	}
	return &user, nil
}

// --- internal helpers ---

func (s *Service) issueTokens(ctx context.Context, user User) (*AuthResponse, error) {
	accessToken, err := s.generateAccessToken(user.ID, user.SubscriptionTier)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	// Refresh token'ı DB'ye kaydet
	hash := hashToken(refreshToken)
	_, err = s.db.Exec(ctx, `
		INSERT INTO refresh_tokens(user_id, token_hash, expires_at)
		VALUES($1, $2, NOW() + INTERVAL '30 days')
	`, user.ID, hash)
	if err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}

func (s *Service) generateAccessToken(userID, tier string) (string, error) {
	claims := Claims{
		UserID:           userID,
		SubscriptionTier: tier,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func (s *Service) generateRefreshToken(userID string) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        uuid.New().String(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWTRefreshSecret))
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", h)
}
