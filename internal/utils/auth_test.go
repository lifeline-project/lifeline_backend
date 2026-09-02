package utils

import (
	"testing"
)

func TestHashPasswordAndCompare(t *testing.T) {
	password := "supersecret123"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Expected no error when hashing password, got %v", err)
	}

	if hashed == password {
		t.Errorf("Expected hashed password to not match plain password")
	}

	if !CheckPasswordHash(password, hashed) {
		t.Errorf("Expected CheckPasswordHash to return true for matching passwords")
	}

	if CheckPasswordHash("wrongpassword", hashed) {
		t.Errorf("Expected CheckPasswordHash to return false for mismatching password")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	userID := uint(42)
	role := "PATIENT"
	secret := "myjwtsecretkey"

	token, err := GenerateToken(userID, role, secret)
	if err != nil {
		t.Fatalf("Expected no error when generating token, got %v", err)
	}

	claims, err := ParseToken(token, secret)
	if err != nil {
		t.Fatalf("Expected no error when parsing valid token, got %v", err)
	}

	claimsMap := *claims
	retrievedUserID, ok1 := claimsMap["user_id"].(float64) // JSON numbers are parsed as float64
	retrievedRole, ok2 := claimsMap["role"].(string)

	if !ok1 || uint(retrievedUserID) != userID {
		t.Errorf("Expected user_id to be %d, got %v", userID, claimsMap["user_id"])
	}

	if !ok2 || retrievedRole != role {
		t.Errorf("Expected role to be %s, got %v", role, claimsMap["role"])
	}

	// Test invalid secret
	_, err = ParseToken(token, "wrongsecret")
	if err == nil {
		t.Errorf("Expected error when parsing token with wrong secret")
	}
}
