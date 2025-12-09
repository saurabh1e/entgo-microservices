package jwt

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	goredis "github.com/redis/go-redis/v9"
	"github.com/saurabh/entgo-microservices/pkg/logger"
	"github.com/saurabh/entgo-microservices/pkg/redis"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	secretKey         string
	expiryHours       int
	refreshExpiryDays int
	tokenService      *redis.TokenService
}

type Claims struct {
	UserID    int    `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	TokenType string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

func NewService(secretKey string, expiryHours int, redisClient *goredis.Client, serviceName string) *Service {
	return &Service{
		secretKey:         secretKey,
		expiryHours:       expiryHours,
		refreshExpiryDays: 30, // Refresh tokens expire in 30 days
		tokenService:      redis.NewTokenService(redisClient, serviceName),
	}
}

// generateTokenID creates a unique token ID
func (j *Service) generateTokenID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateTokenPair creates both access and refresh tokens with Redis tracking
func (j *Service) GenerateTokenPair(ctx context.Context, userID int, username, email string) (accessToken, refreshToken string, err error) {
	// Generate access token
	accessToken, accessTokenID, err := j.generateToken(userID, username, email, "access", time.Duration(j.expiryHours)*time.Hour)
	if err != nil {
		return "", "", err
	}

	// Generate refresh token
	refreshToken, refreshTokenID, err := j.generateToken(userID, username, email, "refresh", time.Duration(j.refreshExpiryDays)*24*time.Hour)
	if err != nil {
		return "", "", err
	}

	// Add both tokens to whitelist
	if err := j.tokenService.AddToWhitelist(ctx, accessTokenID, time.Duration(j.expiryHours)*time.Hour); err != nil {
		logger.WithError(err).Error("Failed to whitelist access token")
		return "", "", err
	}

	if err := j.tokenService.AddToWhitelist(ctx, refreshTokenID, time.Duration(j.refreshExpiryDays)*24*time.Hour); err != nil {
		logger.WithError(err).Error("Failed to whitelist refresh token")
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// generateToken creates a new JWT token with unique ID
func (j *Service) generateToken(userID int, username, email, tokenType string, expiry time.Duration) (string, string, error) {
	tokenID, err := j.generateTokenID()
	if err != nil {
		return "", "", err
	}

	claims := &Claims{
		UserID:    userID,
		Username:  username,
		Email:     email,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "auth",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secretKey))
	return tokenString, tokenID, err
}

// ValidateToken validates and parses a JWT token with Redis checks
func (j *Service) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check token validity in Redis
	isValid, err := j.tokenService.IsTokenValid(ctx, claims.ID)
	if err != nil {
		logger.WithError(err).Error("Failed to check token validity in Redis")
		return nil, errors.New("token validation failed")
	}

	if !isValid {
		return nil, errors.New("token has been revoked or is not valid")
	}

	return claims, nil
}

// RefreshAccessToken generates a new access token using a refresh token
func (j *Service) RefreshAccessToken(ctx context.Context, refreshToken string) (string, string, error) {
	claims, err := j.ValidateToken(ctx, refreshToken)
	if err != nil {
		return "", "", err
	}

	if claims.TokenType != "refresh" {
		return "", "", errors.New("invalid token type")
	}

	// Generate new token pair
	return j.GenerateTokenPair(ctx, claims.UserID, claims.Username, claims.Email)
}

// RevokeToken adds a token to the blacklist
func (j *Service) RevokeToken(ctx context.Context, tokenString string) error {
	claims, err := j.ValidateToken(ctx, tokenString)
	if err != nil {
		// Even if validation fails, try to extract the token ID for blacklisting
		token, parseErr := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(j.secretKey), nil
		})

		if parseErr != nil {
			return err // Return original validation error
		}

		if claims, ok := token.Claims.(*Claims); ok && claims.ID != "" {
			// Calculate remaining TTL based on expiry
			ttl := time.Until(claims.ExpiresAt.Time)
			if ttl > 0 {
				return j.tokenService.RevokeToken(ctx, claims.ID, ttl)
			}
		}
		return err
	}

	// Calculate remaining TTL
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		return nil // Token already expired
	}

	return j.tokenService.RevokeToken(ctx, claims.ID, ttl)
}

// RevokeAllUserTokens revokes all tokens for a specific user (useful for logout from all devices)
func (j *Service) RevokeAllUserTokens(ctx context.Context, userID int) error {
	logger.WithField("user_id", userID).Info("Revoking all user tokens - implement pattern matching if needed")
	// This is a placeholder - you can implement pattern matching for user tokens
	// or maintain a separate Redis set for each user's active tokens
	return nil
}

// HashPassword hashes a plain text password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword compares a plain text password with a hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
