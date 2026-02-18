//go:build fts5

package service_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/yu01/picohub/internal/config"
	"github.com/yu01/picohub/internal/database"
	"github.com/yu01/picohub/internal/model"
	"github.com/yu01/picohub/internal/repository"
	"github.com/yu01/picohub/internal/service"
)

const testJWTSecret = "test-secret-for-auth-service"

// setupAuthService creates an in-memory database, repositories, and returns
// a configured AuthService for unit testing.
func setupAuthService(t *testing.T) *service.AuthService {
	t.Helper()

	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	cfg := &config.Config{
		JWTSecret:  testJWTSecret,
		JWTExpiry:  1 * time.Hour,
		BcryptCost: 4, // low cost for fast tests
	}

	userRepo := repository.NewUserRepository(db)
	return service.NewAuthService(userRepo, cfg)
}

// --- Register ---

func TestRegisterSuccess(t *testing.T) {
	authService := setupAuthService(t)

	resp, err := authService.Register(model.RegisterRequest{
		Username:    "testuser",
		Email:       "testuser@example.com",
		Password:    "securepassword",
		DisplayName: "Test User",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Username != "testuser" {
		t.Errorf("expected username=testuser, got %q", resp.User.Username)
	}
	if resp.User.Email != "testuser@example.com" {
		t.Errorf("expected email=testuser@example.com, got %q", resp.User.Email)
	}
	if resp.User.DisplayName != "Test User" {
		t.Errorf("expected display_name='Test User', got %q", resp.User.DisplayName)
	}
	if resp.User.ID == 0 {
		t.Error("expected non-zero user ID")
	}
}

func TestRegisterDefaultDisplayName(t *testing.T) {
	authService := setupAuthService(t)

	resp, err := authService.Register(model.RegisterRequest{
		Username: "noname",
		Email:    "noname@example.com",
		Password: "securepassword",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// When display_name is empty, it should default to username
	if resp.User.DisplayName != "noname" {
		t.Errorf("expected display_name=noname (defaulted from username), got %q", resp.User.DisplayName)
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	authService := setupAuthService(t)

	_, err := authService.Register(model.RegisterRequest{
		Username: "first",
		Email:    "dup@example.com",
		Password: "securepassword",
	})
	if err != nil {
		t.Fatalf("first register: %v", err)
	}

	_, err = authService.Register(model.RegisterRequest{
		Username: "second",
		Email:    "dup@example.com",
		Password: "securepassword",
	})
	if err == nil {
		t.Fatal("expected error for duplicate email, got nil")
	}
	if err != service.ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	authService := setupAuthService(t)

	_, err := authService.Register(model.RegisterRequest{
		Username: "sameuser",
		Email:    "first@example.com",
		Password: "securepassword",
	})
	if err != nil {
		t.Fatalf("first register: %v", err)
	}

	_, err = authService.Register(model.RegisterRequest{
		Username: "sameuser",
		Email:    "second@example.com",
		Password: "securepassword",
	})
	if err == nil {
		t.Fatal("expected error for duplicate username, got nil")
	}
	if err != service.ErrUserExists {
		t.Errorf("expected ErrUserExists, got %v", err)
	}
}

// --- Login ---

func TestLoginSuccess(t *testing.T) {
	authService := setupAuthService(t)

	// Register first
	_, err := authService.Register(model.RegisterRequest{
		Username: "logintest",
		Email:    "logintest@example.com",
		Password: "mypassword123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Login
	resp, err := authService.Login(model.LoginRequest{
		Email:    "logintest@example.com",
		Password: "mypassword123",
	})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Username != "logintest" {
		t.Errorf("expected username=logintest, got %q", resp.User.Username)
	}
	if resp.User.Email != "logintest@example.com" {
		t.Errorf("expected email=logintest@example.com, got %q", resp.User.Email)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	authService := setupAuthService(t)

	_, err := authService.Register(model.RegisterRequest{
		Username: "wrongpw",
		Email:    "wrongpw@example.com",
		Password: "correctpassword",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = authService.Login(model.LoginRequest{
		Email:    "wrongpw@example.com",
		Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if err != service.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginNonexistentUser(t *testing.T) {
	authService := setupAuthService(t)

	_, err := authService.Login(model.LoginRequest{
		Email:    "ghost@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected error for nonexistent user, got nil")
	}
	if err != service.ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

// --- Token generation and validation ---

func TestTokenGenerationAndValidation(t *testing.T) {
	authService := setupAuthService(t)

	resp, err := authService.Register(model.RegisterRequest{
		Username: "tokentest",
		Email:    "tokentest@example.com",
		Password: "securepassword",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Parse the token
	token, err := jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Fatalf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(testJWTSecret), nil
	})
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if !token.Valid {
		t.Fatal("expected token to be valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("expected MapClaims")
	}

	// Verify claims
	userID := int64(claims["user_id"].(float64))
	if userID != resp.User.ID {
		t.Errorf("expected user_id=%d, got %d", resp.User.ID, userID)
	}

	username := claims["username"].(string)
	if username != "tokentest" {
		t.Errorf("expected username=tokentest, got %q", username)
	}

	isAdmin := claims["is_admin"].(bool)
	if isAdmin != false {
		t.Errorf("expected is_admin=false, got %v", isAdmin)
	}

	// Verify expiry is set in the future
	exp := int64(claims["exp"].(float64))
	if exp <= time.Now().Unix() {
		t.Error("expected exp to be in the future")
	}

	// Verify iat is set
	iat := int64(claims["iat"].(float64))
	if iat == 0 {
		t.Error("expected iat to be set")
	}
}

func TestTokenWithWrongSecretFails(t *testing.T) {
	authService := setupAuthService(t)

	resp, err := authService.Register(model.RegisterRequest{
		Username: "wrongsecret",
		Email:    "wrongsecret@example.com",
		Password: "securepassword",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// Try to parse with wrong secret
	_, err = jwt.Parse(resp.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte("wrong-secret"), nil
	})
	if err == nil {
		t.Fatal("expected error when parsing with wrong secret")
	}
}

func TestGetUserAfterRegister(t *testing.T) {
	authService := setupAuthService(t)

	resp, err := authService.Register(model.RegisterRequest{
		Username:    "getme",
		Email:       "getme@example.com",
		Password:    "securepassword",
		DisplayName: "Get Me",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	user, err := authService.GetUser(resp.User.ID)
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	if user.Username != "getme" {
		t.Errorf("expected username=getme, got %q", user.Username)
	}
	if user.Email != "getme@example.com" {
		t.Errorf("expected email=getme@example.com, got %q", user.Email)
	}
	if user.DisplayName != "Get Me" {
		t.Errorf("expected display_name='Get Me', got %q", user.DisplayName)
	}
}

func TestGetUserNotFound(t *testing.T) {
	authService := setupAuthService(t)

	_, err := authService.GetUser(99999)
	if err == nil {
		t.Fatal("expected error for nonexistent user ID, got nil")
	}
}
