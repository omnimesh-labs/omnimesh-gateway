package auth

import (
	"context"
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
	cache              TokenCache
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	// Use memory cache as fallback
	cache := NewMemoryTokenCache()
	
	return &JWTManager{
		secret:             []byte(secret),
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
		cache:              cache,
	}
}

// NewJWTManagerWithCache creates a new JWT manager with custom cache
func NewJWTManagerWithCache(secret string, accessExpiry, refreshExpiry time.Duration, cache TokenCache) *JWTManager {
	return &JWTManager{
		secret:             []byte(secret),
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
		cache:              cache,
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
	if isBlacklisted, err := j.isTokenBlacklisted(context.Background(), tokenString); err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	} else if isBlacklisted {
		return nil, errors.New("token has been revoked")
	}

	return claims, nil
}

// InvalidateToken adds token to blacklist
func (j *JWTManager) InvalidateToken(ctx context.Context, tokenString string) error {
	// Parse token to get expiration time, allowing expired tokens for blacklisting
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return j.secret, nil
	}, jwt.WithoutClaimsValidation())

	if err != nil {
		return fmt.Errorf("failed to parse token for blacklisting: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return errors.New("invalid token claims")
	}

	// Calculate time until expiration
	expiration := time.Until(claims.ExpiresAt.Time)
	if expiration <= 0 {
		// Token already expired, no need to blacklist
		return nil
	}

	// Store token in cache until it expires
	if err := j.cache.Set(ctx, tokenString, expiration); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	return nil
}

// isTokenBlacklisted checks if token is in blacklist
func (j *JWTManager) isTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	return j.cache.IsBlacklisted(ctx, tokenString)
}

// CleanupExpiredTokens removes expired tokens from blacklist
// Should be called periodically by a background job
func (j *JWTManager) CleanupExpiredTokens(ctx context.Context) error {
	return j.cache.Cleanup(ctx)
}

// Close closes the cache connection
func (j *JWTManager) Close() error {
	return j.cache.Close()
}
