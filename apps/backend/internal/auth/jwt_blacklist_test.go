package auth

import (
	"context"
	"testing"
	"time"

	"mcp-gateway/apps/backend/internal/types"
)

func TestJWTBlacklisting(t *testing.T) {
	// Create JWT manager with memory cache
	cache := NewMemoryTokenCache()
	jwtManager := NewJWTManagerWithCache("test-secret", time.Hour, time.Hour*24, cache)
	
	// Create a test user
	user := &types.User{
		ID:             "test-user-id",
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: "test-org-id",
		Role:           "user",
	}
	
	// Generate access token
	token, err := jwtManager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}
	
	// Validate token (should succeed)
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	
	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}
	
	// Blacklist the token
	ctx := context.Background()
	err = jwtManager.InvalidateToken(ctx, token)
	if err != nil {
		t.Fatalf("Failed to blacklist token: %v", err)
	}
	
	// Try to validate token again (should fail)
	_, err = jwtManager.ValidateToken(token)
	if err == nil {
		t.Fatal("Expected validation to fail for blacklisted token")
	}
	
	if err.Error() != "token has been revoked" {
		t.Errorf("Expected 'token has been revoked' error, got: %v", err)
	}
}

func TestJWTBlacklistingWithRedis(t *testing.T) {
	// Test Redis cache if available (skip if Redis not available)
	cache, err := NewRedisTokenCache("localhost:6379", "", 0)
	if err != nil {
		t.Skip("Redis not available, skipping Redis blacklist test")
		return
	}
	defer cache.Close()
	
	jwtManager := NewJWTManagerWithCache("test-secret", time.Hour, time.Hour*24, cache)
	
	// Create a test user
	user := &types.User{
		ID:             "test-user-redis-id",
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: "test-org-id",
		Role:           "user",
	}
	
	// Generate access token
	token, err := jwtManager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}
	
	// Validate token (should succeed)
	claims, err := jwtManager.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}
	
	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}
	
	// Blacklist the token
	ctx := context.Background()
	err = jwtManager.InvalidateToken(ctx, token)
	if err != nil {
		t.Fatalf("Failed to blacklist token: %v", err)
	}
	
	// Try to validate token again (should fail)
	_, err = jwtManager.ValidateToken(token)
	if err == nil {
		t.Fatal("Expected validation to fail for blacklisted token")
	}
	
	if err.Error() != "token has been revoked" {
		t.Errorf("Expected 'token has been revoked' error, got: %v", err)
	}
}

func TestJWTBlacklistExpiration(t *testing.T) {
	cache := NewMemoryTokenCache()
	jwtManager := NewJWTManagerWithCache("test-secret", time.Second*5, time.Hour*24, cache) // Use longer expiry
	
	// Create a test user
	user := &types.User{
		ID:             "test-user-expiry-id",
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: "test-org-id",
		Role:           "user",
	}
	
	// Generate access token
	token, err := jwtManager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}
	
	// Blacklist the token
	ctx := context.Background()
	err = jwtManager.InvalidateToken(ctx, token)
	if err != nil {
		t.Fatalf("Failed to blacklist token: %v", err)
	}
	
	// Verify token is blacklisted
	_, err = jwtManager.ValidateToken(token)
	if err == nil {
		t.Fatal("Expected validation to fail for blacklisted token")
	}
	
	if err.Error() != "token has been revoked" {
		t.Errorf("Expected 'token has been revoked' error, got: %v", err)
	}
}

func TestMemoryCacheCleanup(t *testing.T) {
	cache := NewMemoryTokenCache()
	ctx := context.Background()
	
	// Add some tokens to cache
	err := cache.Set(ctx, "token1", time.Millisecond*50)
	if err != nil {
		t.Fatalf("Failed to set token in cache: %v", err)
	}
	
	err = cache.Set(ctx, "token2", time.Hour) // Long expiry
	if err != nil {
		t.Fatalf("Failed to set token in cache: %v", err)
	}
	
	// Check both tokens are blacklisted
	blacklisted, err := cache.IsBlacklisted(ctx, "token1")
	if err != nil || !blacklisted {
		t.Error("Expected token1 to be blacklisted")
	}
	
	blacklisted, err = cache.IsBlacklisted(ctx, "token2")
	if err != nil || !blacklisted {
		t.Error("Expected token2 to be blacklisted")
	}
	
	// Wait for token1 to expire
	time.Sleep(time.Millisecond * 60)
	
	// token1 should no longer be blacklisted (expired and cleaned up)
	blacklisted, err = cache.IsBlacklisted(ctx, "token1")
	if err != nil || blacklisted {
		t.Error("Expected token1 to be automatically cleaned up")
	}
	
	// token2 should still be blacklisted
	blacklisted, err = cache.IsBlacklisted(ctx, "token2")
	if err != nil || !blacklisted {
		t.Error("Expected token2 to still be blacklisted")
	}
	
	// Run cleanup
	err = cache.Cleanup(ctx)
	if err != nil {
		t.Fatalf("Failed to run cleanup: %v", err)
	}
}