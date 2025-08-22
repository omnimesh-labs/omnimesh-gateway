package auth

import (
	"testing"
	"time"
)

func TestService_HashAndValidatePassword(t *testing.T) {
	// Setup service with test config
	config := &Config{
		JWTSecret:          "test-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		BCryptCost:         10, // Lower cost for faster tests
	}

	service := &Service{
		config: config,
	}

	// Test password hashing
	password := "testpassword123"
	hash, err := service.hashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hash == "" {
		t.Fatal("Hash should not be empty")
	}

	if hash == password {
		t.Fatal("Hash should not equal original password")
	}

	// Test password validation - correct password
	if !service.validatePassword(password, hash) {
		t.Fatal("Should validate correct password")
	}

	// Test password validation - incorrect password
	if service.validatePassword("wrongpassword", hash) {
		t.Fatal("Should reject incorrect password")
	}

	// Test password validation - empty password
	if service.validatePassword("", hash) {
		t.Fatal("Should reject empty password")
	}
}

func TestService_HashPassword_DifferentPasswords(t *testing.T) {
	config := &Config{
		BCryptCost: 10,
	}

	service := &Service{
		config: config,
	}

	password1 := "password1"
	password2 := "password2"

	hash1, err := service.hashPassword(password1)
	if err != nil {
		t.Fatalf("Failed to hash password1: %v", err)
	}

	hash2, err := service.hashPassword(password2)
	if err != nil {
		t.Fatalf("Failed to hash password2: %v", err)
	}

	// Different passwords should produce different hashes
	if hash1 == hash2 {
		t.Fatal("Different passwords should produce different hashes")
	}

	// Even same password should produce different hashes due to salt
	hash1b, err := service.hashPassword(password1)
	if err != nil {
		t.Fatalf("Failed to hash password1 again: %v", err)
	}

	if hash1 == hash1b {
		t.Fatal("Same password should produce different hashes due to salt")
	}

	// But both hashes should validate the same password
	if !service.validatePassword(password1, hash1) {
		t.Fatal("hash1 should validate password1")
	}

	if !service.validatePassword(password1, hash1b) {
		t.Fatal("hash1b should validate password1")
	}
}