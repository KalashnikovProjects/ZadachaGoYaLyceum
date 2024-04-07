package auth

import (
	"Zadacha/internal/my_errors"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"os"
	"testing"
)

func TestLoadUserIdFromToken(t *testing.T) {
	os.Setenv("HMAC", "my_secret_key")
	defer os.Unsetenv("HMAC")

	userId := 42
	token, err := GenerateTokenFromId(userId)
	if err != nil {
		t.Errorf("failed to generate token: %v", err)
	}

	loadedUserId, err := LoadUserIdFromToken(token)
	if err != nil {
		t.Errorf("failed to load user ID from token: %v", err)
	}

	if loadedUserId != userId {
		t.Errorf("expected user ID %d, got %d", userId, loadedUserId)
	}

	// Test error handling
	invalidToken := "invalid.token.string"
	_, err = LoadUserIdFromToken(invalidToken)
	if err == nil {
		t.Errorf("expected an error when loading user ID from invalid token")
	}

	os.Setenv("HMAC", "wrong_secret_key")
	defer os.Unsetenv("HMAC")
	_, err = LoadUserIdFromToken(token)
	if err == nil {
		t.Errorf("expected an error when loading user ID from token with wrong secret key")
	}
	if err == nil {
		t.Errorf("expected AuthenticationError, got %v", err)
	}

	tokenWithInvalidMethod := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6NDJ9.invalid_signature"
	_, err = LoadUserIdFromToken(tokenWithInvalidMethod)
	if err == nil {
		t.Errorf("expected AuthenticationError, got %v", err)
	}

	// Test error handling for claims conversion
	tokenWithInvalidClaims := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyX9.invalid_signature"
	_, err = LoadUserIdFromToken(tokenWithInvalidClaims)
	if err == nil {
		t.Errorf("expected AuthenticationError, got %v", err)
	}
}

func TestGenerateTokenFromId(t *testing.T) {
	os.Setenv("HMAC", "my_secret_key")
	defer os.Unsetenv("HMAC")

	userId := 42
	token, err := GenerateTokenFromId(userId)
	if err != nil {
		t.Errorf("failed to generate token: %v", err)
	}

	tokenFromString, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, my_errors.AuthenticationError
		}

		return []byte(os.Getenv("HMAC")), nil
	})

	if err != nil {
		t.Errorf("failed to parse token: %v", err)
	}

	if claims, ok := tokenFromString.Claims.(jwt.MapClaims); ok {
		if claims["id"].(float64) != float64(userId) {
			t.Errorf("expected user ID %d, got %d", userId, int(claims["id"].(float64)))
		}
	} else {
		t.Errorf("failed to get claims from token")
	}

}

func TestComparePasswordWithHashed(t *testing.T) {
	password := "mypassword"
	hashedPassword, err := GenerateHashedPassword(password)
	if err != nil {
		t.Errorf("failed to generate hashed password: %v", err)
	}

	if err := ComparePasswordWithHashed(password, hashedPassword); err != nil {
		t.Errorf("failed to compare password with hashed password: %v", err)
	}

	if err := ComparePasswordWithHashed("wrongpassword", hashedPassword); err == nil {
		t.Errorf("expected error when comparing wrong password")
	}

	// Test error handling
	if err := ComparePasswordWithHashed(password, "invalidhash"); err == nil {
		t.Errorf("expected an error when comparing password with invalid hash")
	}
}

func TestGenerateHashedPassword(t *testing.T) {
	password := "mypassword"
	hashedPassword1, err := GenerateHashedPassword(password)
	if err != nil {
		t.Errorf("failed to generate hashed password: %v", err)
	}

	hashedPassword2, err := GenerateHashedPassword(password)
	if err != nil {
		t.Errorf("failed to generate hashed password: %v", err)
	}

	if hashedPassword1 == hashedPassword2 {
		t.Errorf("hashed passwords should be different")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword1), []byte(password)); err != nil {
		t.Errorf("failed to compare password with hashed password: %v", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword2), []byte(password)); err != nil {
		t.Errorf("failed to compare password with hashed password: %v", err)
	}
}
