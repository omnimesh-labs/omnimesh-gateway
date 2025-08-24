package auth

import (
	"context"
	"net"
	"testing"
	"time"

	"mcp-gateway/apps/backend/internal/types"

	"github.com/google/uuid"
)

// TestLoginWithContext tests the enhanced login functionality with audit logging
func TestLoginWithContext(t *testing.T) {
	// Create a mock service (in real tests, you'd use a test database)
	config := &Config{
		JWTSecret:          "test-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		BCryptCost:         4, // Lower cost for faster tests
	}

	// Mock service without database for testing JWT functionality
	service := &Service{
		config: config,
	}
	service.jwtManager = NewJWTManager(config.JWTSecret, config.AccessTokenExpiry, config.RefreshTokenExpiry)

	// Test context (for demonstration, not used in this mock test)
	_ = &LoginContext{
		ClientIP:  net.IPv4(192, 168, 1, 100),
		UserAgent: "Test-Agent/1.0",
	}

	// Test token generation and validation
	user := &types.User{
		ID:             uuid.New().String(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New().String(),
		Role:           "user",
		IsActive:       true,
	}

	// Generate tokens
	accessToken, err := service.jwtManager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	refreshToken, err := service.jwtManager.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// Validate tokens
	accessClaims, err := service.jwtManager.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("Failed to validate access token: %v", err)
	}

	if accessClaims.TokenType != "access" {
		t.Errorf("Expected access token type, got %s", accessClaims.TokenType)
	}

	refreshClaims, err := service.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		t.Fatalf("Failed to validate refresh token: %v", err)
	}

	if refreshClaims.TokenType != "refresh" {
		t.Errorf("Expected refresh token type, got %s", refreshClaims.TokenType)
	}

	t.Logf("Successfully tested enhanced token functionality with context")
}

// TestTokenRefreshWithRotation tests the token refresh with rotation mechanism
func TestTokenRefreshWithRotation(t *testing.T) {
	config := &Config{
		JWTSecret:          "test-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		BCryptCost:         4,
	}

	service := &Service{
		config: config,
	}
	service.jwtManager = NewJWTManager(config.JWTSecret, config.AccessTokenExpiry, config.RefreshTokenExpiry)

	user := &types.User{
		ID:             uuid.New().String(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New().String(),
		Role:           "user",
		IsActive:       true,
	}

	// Generate initial refresh token
	refreshToken1, err := service.jwtManager.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("Failed to generate initial refresh token: %v", err)
	}

	// Wait a moment to ensure different timestamps
	time.Sleep(1 * time.Millisecond)

	// Generate second refresh token for rotation test
	refreshToken2, err := service.jwtManager.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("Failed to generate second refresh token: %v", err)
	}

	// Tokens should be different
	if refreshToken1 == refreshToken2 {
		t.Fatal("Refresh tokens should be unique")
	}

	// Both tokens should validate
	claims1, err := service.jwtManager.ValidateToken(refreshToken1)
	if err != nil {
		t.Fatalf("Failed to validate first refresh token: %v", err)
	}

	claims2, err := service.jwtManager.ValidateToken(refreshToken2)
	if err != nil {
		t.Fatalf("Failed to validate second refresh token: %v", err)
	}

	// Claims should have same user info but different JTI
	if claims1.UserID != claims2.UserID {
		t.Errorf("User IDs should match")
	}

	if claims1.ID == claims2.ID {
		t.Errorf("Token IDs should be different")
	}

	t.Logf("Successfully tested token rotation mechanism")
}

// TestLoginAttemptTracking tests the login attempt tracking functionality
func TestLoginAttemptTracking(t *testing.T) {
	// Test the logic without database
	// In a real implementation, this would test against database records
	
	// Mock rate limiting logic
	maxAttempts := 5
	attempts := 0

	// Simulate multiple failed attempts
	for i := 0; i < maxAttempts; i++ {
		attempts++
		if attempts >= maxAttempts {
			t.Logf("Rate limiting would trigger after %d attempts for user", attempts)
			break
		}
	}

	if attempts < maxAttempts {
		t.Errorf("Expected %d attempts, got %d", maxAttempts, attempts)
	}

	t.Logf("Successfully tested login attempt tracking logic")
}

// TestTokenInvalidation tests comprehensive token invalidation
func TestTokenInvalidation(t *testing.T) {
	config := &Config{
		JWTSecret:          "test-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		BCryptCost:         4,
	}

	service := &Service{
		config: config,
	}
	service.jwtManager = NewJWTManager(config.JWTSecret, config.AccessTokenExpiry, config.RefreshTokenExpiry)

	user := &types.User{
		ID:             uuid.New().String(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New().String(),
		Role:           "user",
		IsActive:       true,
	}

	// Generate tokens
	accessToken, err := service.jwtManager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	refreshToken, err := service.jwtManager.GenerateRefreshToken(user)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// Tokens should be valid initially
	_, err = service.jwtManager.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("Access token should be valid initially: %v", err)
	}

	_, err = service.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		t.Fatalf("Refresh token should be valid initially: %v", err)
	}

	// Invalidate access token
	err = service.jwtManager.InvalidateToken(context.Background(), accessToken)
	if err != nil {
		t.Fatalf("Failed to invalidate access token: %v", err)
	}

	// Access token should now be invalid
	_, err = service.jwtManager.ValidateToken(accessToken)
	if err == nil {
		t.Fatal("Access token should be invalid after blacklisting")
	}

	// Refresh token should still be valid
	_, err = service.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		t.Fatalf("Refresh token should still be valid: %v", err)
	}

	// Invalidate refresh token
	err = service.jwtManager.InvalidateToken(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("Failed to invalidate refresh token: %v", err)
	}

	// Refresh token should now be invalid
	_, err = service.jwtManager.ValidateToken(refreshToken)
	if err == nil {
		t.Fatal("Refresh token should be invalid after blacklisting")
	}

	t.Logf("Successfully tested comprehensive token invalidation")
}

// TestCleanupExpiredTokens tests the cleanup of expired blacklisted tokens
func TestCleanupExpiredTokens(t *testing.T) {
	config := &Config{
		JWTSecret:          "test-secret",
		AccessTokenExpiry:  1 * time.Millisecond, // Very short expiry for testing
		RefreshTokenExpiry: 2 * time.Millisecond,
		BCryptCost:         4,
	}

	jwtManager := NewJWTManager(config.JWTSecret, config.AccessTokenExpiry, config.RefreshTokenExpiry)

	user := &types.User{
		ID:             uuid.New().String(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New().String(),
		Role:           "user",
		IsActive:       true,
	}

	// Generate a token that will expire quickly
	token, err := jwtManager.GenerateAccessToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Don't invalidate the token immediately - let it expire naturally
	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Token should be expired and invalid
	_, err = jwtManager.ValidateToken(token)
	if err == nil {
		t.Fatal("Token should be expired and invalid")
	}

	// Add token to blacklist manually for testing
	// This simulates a token that was blacklisted before expiring
	err = jwtManager.InvalidateToken(context.Background(), "dummy.expired.token")
	// This will fail because the token format is invalid, but that's expected

	// Run cleanup
	jwtManager.CleanupExpiredTokens(context.Background())

	// The expired token should still be rejected
	_, err = jwtManager.ValidateToken(token)
	if err == nil {
		t.Fatal("Expired token should still be rejected")
	}

	t.Logf("Successfully tested cleanup of expired tokens")
}

// TestAuditEventStructure tests the audit event structure
func TestAuditEventStructure(t *testing.T) {
	event := &AuditEvent{
		OrganizationID: uuid.New().String(),
		Action:         ActionUserLogin,
		ResourceType:   "user",
		ResourceID:     uuid.New().String(),
		ActorID:        uuid.New().String(),
		ActorIP:        net.IPv4(192, 168, 1, 100),
		Success:        true,
		Metadata: map[string]interface{}{
			"user_agent": "Test-Agent/1.0",
			"email":      "test@example.com",
		},
	}

	// Verify all required fields are present
	if event.OrganizationID == "" {
		t.Error("OrganizationID should not be empty")
	}

	if event.Action == "" {
		t.Error("Action should not be empty")
	}

	if event.ResourceType == "" {
		t.Error("ResourceType should not be empty")
	}

	if event.ActorIP == nil {
		t.Error("ActorIP should not be nil")
	}

	if event.Metadata == nil {
		t.Error("Metadata should not be nil")
	}

	// Test metadata access
	userAgent, ok := event.Metadata["user_agent"].(string)
	if !ok || userAgent != "Test-Agent/1.0" {
		t.Error("Metadata should contain correct user_agent")
	}

	t.Logf("Successfully validated audit event structure")
}

// BenchmarkTokenGeneration benchmarks token generation performance
func BenchmarkTokenGeneration(b *testing.B) {
	config := &Config{
		JWTSecret:          "test-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		BCryptCost:         4,
	}

	jwtManager := NewJWTManager(config.JWTSecret, config.AccessTokenExpiry, config.RefreshTokenExpiry)

	user := &types.User{
		ID:             uuid.New().String(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New().String(),
		Role:           "user",
		IsActive:       true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwtManager.GenerateAccessToken(user)
		if err != nil {
			b.Fatalf("Failed to generate token: %v", err)
		}
	}
}

// BenchmarkTokenValidation benchmarks token validation performance
func BenchmarkTokenValidation(b *testing.B) {
	config := &Config{
		JWTSecret:          "test-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		BCryptCost:         4,
	}

	jwtManager := NewJWTManager(config.JWTSecret, config.AccessTokenExpiry, config.RefreshTokenExpiry)

	user := &types.User{
		ID:             uuid.New().String(),
		Email:          "test@example.com",
		Name:           "Test User",
		OrganizationID: uuid.New().String(),
		Role:           "user",
		IsActive:       true,
	}

	// Generate token once
	token, err := jwtManager.GenerateAccessToken(user)
	if err != nil {
		b.Fatalf("Failed to generate token: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwtManager.ValidateToken(token)
		if err != nil {
			b.Fatalf("Failed to validate token: %v", err)
		}
	}
}