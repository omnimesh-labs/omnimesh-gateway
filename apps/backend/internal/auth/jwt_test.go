package auth

import (
	"context"
	"testing"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

func TestJWTManager_GenerateAndValidateAccessToken(t *testing.T) {
	// Setup
	jwtManager := NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
	
	user := &types.User{
		ID:             uuid.New().String(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New().String(),
		Role:           "user",
		IsActive:       true,
	}

	// Test access token generation
	token, err := jwtManager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Test token validation
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Verify claims
	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}

	if claims.OrganizationID != user.OrganizationID {
		t.Errorf("Expected organization ID %s, got %s", user.OrganizationID, claims.OrganizationID)
	}

	if claims.Role != user.Role {
		t.Errorf("Expected role %s, got %s", user.Role, claims.Role)
	}

	if claims.TokenType != "access" {
		t.Errorf("Expected token type 'access', got %s", claims.TokenType)
	}
}

func TestJWTManager_GenerateAndValidateRefreshToken(t *testing.T) {
	// Setup
	jwtManager := NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
	
	user := &types.User{
		ID:             uuid.New().String(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New().String(),
		Role:           "user",
		IsActive:       true,
	}

	// Test refresh token generation
	token, err := jwtManager.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	// Test token validation
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Verify claims
	if claims.TokenType != "refresh" {
		t.Errorf("Expected token type 'refresh', got %s", claims.TokenType)
	}
}

func TestJWTManager_InvalidateToken(t *testing.T) {
	// Setup
	jwtManager := NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)
	
	user := &types.User{
		ID:             uuid.New().String(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New().String(),
		Role:           "user",
		IsActive:       true,
	}

	// Generate token
	token, err := jwtManager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	// Validate token works initially
	_, err = jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Token should be valid initially: %v", err)
	}

	// Invalidate token
	err = jwtManager.InvalidateToken(context.Background(), token)
	if err != nil {
		t.Fatalf("Failed to invalidate token: %v", err)
	}

	// Token should now be invalid
	_, err = jwtManager.ValidateToken(token)
	if err == nil {
		t.Fatal("Token should be invalid after blacklisting")
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	jwtManager := NewJWTManager("test-secret-key", 15*time.Minute, 7*24*time.Hour)

	// Test with completely invalid token
	_, err := jwtManager.ValidateToken("invalid.token.here")
	if err == nil {
		t.Fatal("Should reject invalid token")
	}

	// Test with empty token
	_, err = jwtManager.ValidateToken("")
	if err == nil {
		t.Fatal("Should reject empty token")
	}
}

func TestJWTManager_WrongSecret(t *testing.T) {
	// Generate token with one secret
	jwtManager1 := NewJWTManager("secret-1", 15*time.Minute, 7*24*time.Hour)
	
	user := &types.User{
		ID:             uuid.New().String(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New().String(),
		Role:           "user",
		IsActive:       true,
	}

	token, err := jwtManager1.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to validate with different secret
	jwtManager2 := NewJWTManager("secret-2", 15*time.Minute, 7*24*time.Hour)
	_, err = jwtManager2.ValidateToken(token)
	if err == nil {
		t.Fatal("Should reject token signed with different secret")
	}
}