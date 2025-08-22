package auth

import (
	"errors"
	"fmt"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secret             []byte
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secret:             []byte(secret),
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	Role           string `json:"role"`
	TokenType      string `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// GenerateAccessToken generates a new access token
func (j *JWTManager) GenerateAccessToken(user *types.User) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:         user.ID,
		OrganizationID: user.OrganizationID,
		Role:           user.Role,
		TokenType:      "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "mcp-gateway",
			Subject:   user.ID,
			ID:        fmt.Sprintf("%s-%d-%d", user.ID, now.Unix(), now.UnixNano()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a new refresh token
func (j *JWTManager) GenerateRefreshToken(user *types.User) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:         user.ID,
		OrganizationID: user.OrganizationID,
		Role:           user.Role,
		TokenType:      "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(j.refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "mcp-gateway",
			Subject:   user.ID,
			ID:        fmt.Sprintf("%s-refresh-%d-%d", user.ID, now.Unix(), now.UnixNano()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Check if token is blacklisted
	if j.isTokenBlacklisted(tokenString) {
		return nil, errors.New("token has been revoked")
	}

	return claims, nil
}

// tokenBlacklist stores revoked tokens (in production, use Redis)
var tokenBlacklist = make(map[string]time.Time)

// InvalidateToken adds token to blacklist
func (j *JWTManager) InvalidateToken(tokenString string) error {
	// Parse token to get expiration time
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return j.secret, nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token for blacklisting: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return errors.New("invalid token claims")
	}

	// Store token in blacklist until it expires
	tokenBlacklist[tokenString] = claims.ExpiresAt.Time

	// TODO: In production, implement with Redis:
	// j.redisClient.Set(ctx, "blacklist:"+tokenString, "revoked", time.Until(claims.ExpiresAt.Time))

	return nil
}

// isTokenBlacklisted checks if token is in blacklist
func (j *JWTManager) isTokenBlacklisted(tokenString string) bool {
	expiry, exists := tokenBlacklist[tokenString]
	if !exists {
		return false
	}

	// Remove expired tokens from blacklist
	if time.Now().After(expiry) {
		delete(tokenBlacklist, tokenString)
		return false
	}

	return true
}

// CleanupExpiredTokens removes expired tokens from blacklist
// Should be called periodically by a background job
func (j *JWTManager) CleanupExpiredTokens() {
	now := time.Now()
	for token, expiry := range tokenBlacklist {
		if now.After(expiry) {
			delete(tokenBlacklist, token)
		}
	}
}
