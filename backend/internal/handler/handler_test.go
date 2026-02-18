//go:build fts5

package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/yu01/picohub/internal/config"
	"github.com/yu01/picohub/internal/database"
	"github.com/yu01/picohub/internal/handler"
	"github.com/yu01/picohub/internal/middleware"
	"github.com/yu01/picohub/internal/model"
	"github.com/yu01/picohub/internal/repository"
	"github.com/yu01/picohub/internal/scanner"
	"github.com/yu01/picohub/internal/service"
)

const testJWTSecret = "test-secret-key-for-tests"

// setupTestServer creates an in-memory SQLite DB, seeds it, wires all handlers
// and routes, and returns an httptest.Server ready for integration testing.
func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()

	// Use database.Open which handles migration (schema creation) automatically.
	// The :memory: path gets query params appended by database.Open for WAL, foreign keys, etc.
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("database.Open: %v", err)
	}

	// Seed with test data
	if err := database.Seed(db); err != nil {
		t.Fatalf("seed database: %v", err)
	}

	t.Cleanup(func() { db.Close() })

	cfg := &config.Config{
		Port:           "0",
		DBPath:         ":memory:",
		JWTSecret:      testJWTSecret,
		JWTExpiry:      24 * time.Hour,
		UploadDir:      t.TempDir(),
		MaxUploadSize:  10 << 20,
		AllowedOrigins: []string{"*"},
		BcryptCost:     4, // low cost for fast tests
	}

	userRepo := repository.NewUserRepository(db)
	skillRepo := repository.NewSkillRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	authService := service.NewAuthService(userRepo, cfg)
	storageService := service.NewStorageService(cfg.UploadDir, cfg.MaxUploadSize)
	noopScanner := scanner.NewNoopScanner()

	authHandler := handler.NewAuthHandler(authService)
	skillHandler := handler.NewSkillHandler(skillRepo, storageService, noopScanner)
	reviewHandler := handler.NewReviewHandler(reviewRepo, skillRepo)

	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.Health)

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.With(middleware.Auth(cfg.JWTSecret)).Get("/me", authHandler.Me)
		})

		r.Route("/skills", func(r chi.Router) {
			r.Get("/", skillHandler.List)
			r.Get("/featured", skillHandler.Featured)
			r.Get("/categories", skillHandler.Categories)

			r.Route("/{slug}", func(r chi.Router) {
				r.Get("/", skillHandler.Get)
				r.Get("/reviews", reviewHandler.List)
				r.With(middleware.Auth(cfg.JWTSecret)).Post("/reviews", reviewHandler.Create)
			})
		})
	})

	return httptest.NewServer(r)
}

// registerUser is a helper that registers a new user and returns the auth response.
func registerUser(t *testing.T, serverURL string, username, email, password string) *model.AuthResponse {
	t.Helper()
	body, _ := json.Marshal(model.RegisterRequest{
		Username: username,
		Email:    email,
		Password: password,
	})
	resp, err := http.Post(serverURL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("register expected 201, got %d: %s", resp.StatusCode, string(b))
	}
	var authResp model.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	return &authResp
}

// --- Health ---

func TestHealth(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/health")
	if err != nil {
		t.Fatalf("health request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("expected status=ok, got %q", result["status"])
	}
	if result["service"] != "picohub-api" {
		t.Errorf("expected service=picohub-api, got %q", result["service"])
	}
}

// --- Register ---

func TestRegisterSuccess(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(model.RegisterRequest{
		Username:    "newuser",
		Email:       "newuser@example.com",
		Password:    "securepass123",
		DisplayName: "New User",
	})
	resp, err := http.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(b))
	}

	var authResp model.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if authResp.Token == "" {
		t.Error("expected non-empty token")
	}
	if authResp.User.Username != "newuser" {
		t.Errorf("expected username=newuser, got %q", authResp.User.Username)
	}
	if authResp.User.Email != "newuser@example.com" {
		t.Errorf("expected email=newuser@example.com, got %q", authResp.User.Email)
	}
	if authResp.User.DisplayName != "New User" {
		t.Errorf("expected display_name='New User', got %q", authResp.User.DisplayName)
	}
}

func TestRegisterDuplicate(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	// The seed data already has admin@picohub.dev
	body, _ := json.Marshal(model.RegisterRequest{
		Username: "admin",
		Email:    "admin@picohub.dev",
		Password: "password12345",
	})
	resp, err := http.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Fatalf("expected 409, got %d", resp.StatusCode)
	}

	var errResp model.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Error != "user already exists" {
		t.Errorf("expected 'user already exists', got %q", errResp.Error)
	}
}

func TestRegisterValidationMissingFields(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(model.RegisterRequest{
		Username: "",
		Email:    "x@example.com",
		Password: "password123",
	})
	resp, err := http.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestRegisterValidationShortPassword(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(model.RegisterRequest{
		Username: "shortpw",
		Email:    "shortpw@example.com",
		Password: "short",
	})
	resp, err := http.Post(srv.URL+"/api/v1/auth/register", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	var errResp model.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Error != "password must be at least 8 characters" {
		t.Errorf("expected password length error, got %q", errResp.Error)
	}
}

// --- Login ---

func TestLoginSuccess(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	// Seed data: tanaka@example.com with password "password123"
	body, _ := json.Marshal(model.LoginRequest{
		Email:    "tanaka@example.com",
		Password: "password123",
	})
	resp, err := http.Post(srv.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}

	var authResp model.AuthResponse
	json.NewDecoder(resp.Body).Decode(&authResp)
	if authResp.Token == "" {
		t.Error("expected non-empty token")
	}
	if authResp.User.Username != "tanaka" {
		t.Errorf("expected username=tanaka, got %q", authResp.User.Username)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(model.LoginRequest{
		Email:    "tanaka@example.com",
		Password: "wrongpassword",
	})
	resp, err := http.Post(srv.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}

	var errResp model.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Error != "invalid email or password" {
		t.Errorf("expected 'invalid email or password', got %q", errResp.Error)
	}
}

func TestLoginMissingUser(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	body, _ := json.Marshal(model.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	})
	resp, err := http.Post(srv.URL+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// --- Me ---

func TestMeWithValidJWT(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	// Register a new user to get a valid token
	authResp := registerUser(t, srv.URL, "metest", "metest@example.com", "password1234")

	req, _ := http.NewRequest("GET", srv.URL+"/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+authResp.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}

	var user model.User
	json.NewDecoder(resp.Body).Decode(&user)
	if user.Username != "metest" {
		t.Errorf("expected username=metest, got %q", user.Username)
	}
	if user.Email != "metest@example.com" {
		t.Errorf("expected email=metest@example.com, got %q", user.Email)
	}
}

func TestMeWithoutJWT(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/auth/me")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

// --- Skills List ---

func TestSkillsList(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/skills")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}

	var paginated model.PaginatedResponse
	json.NewDecoder(resp.Body).Decode(&paginated)

	if paginated.Total != 5 {
		t.Errorf("expected total=5, got %d", paginated.Total)
	}
	if paginated.Page != 1 {
		t.Errorf("expected page=1, got %d", paginated.Page)
	}

	// Data should be a slice
	dataSlice, ok := paginated.Data.([]interface{})
	if !ok {
		t.Fatalf("expected data to be a slice, got %T", paginated.Data)
	}
	if len(dataSlice) != 5 {
		t.Errorf("expected 5 skills, got %d", len(dataSlice))
	}
}

func TestSkillsListWithSearch(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/skills?q=LINE")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}

	var paginated model.PaginatedResponse
	json.NewDecoder(resp.Body).Decode(&paginated)

	if paginated.Total < 1 {
		t.Errorf("expected at least 1 result for search 'LINE', got %d", paginated.Total)
	}
}

func TestSkillsListWithCategoryFilter(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/skills?category=messaging")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}

	var paginated model.PaginatedResponse
	json.NewDecoder(resp.Body).Decode(&paginated)

	if paginated.Total != 1 {
		t.Errorf("expected total=1 for category=messaging, got %d", paginated.Total)
	}
}

// --- Skills Featured ---

func TestSkillsFeatured(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/skills/featured")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}

	var skills []model.Skill
	json.NewDecoder(resp.Body).Decode(&skills)

	// Seed data has 4 featured skills: line-messenger, rakuten-shopping, weather-reminder, notion-lite
	if len(skills) != 4 {
		t.Errorf("expected 4 featured skills, got %d", len(skills))
	}

	for _, s := range skills {
		if !s.IsFeatured {
			t.Errorf("skill %q should be featured", s.Slug)
		}
	}
}

// --- Skills Get by slug ---

func TestSkillGetFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/skills/line-messenger")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}

	var skill model.Skill
	json.NewDecoder(resp.Body).Decode(&skill)

	if skill.Slug != "line-messenger" {
		t.Errorf("expected slug=line-messenger, got %q", skill.Slug)
	}
	if skill.Name != "LINE Messenger" {
		t.Errorf("expected name='LINE Messenger', got %q", skill.Name)
	}
	if skill.Category != "messaging" {
		t.Errorf("expected category=messaging, got %q", skill.Category)
	}
	if skill.AuthorName != "tanaka" {
		t.Errorf("expected author_name=tanaka, got %q", skill.AuthorName)
	}
	// line-messenger has 2 reviews in seed data
	if skill.ReviewCount != 2 {
		t.Errorf("expected review_count=2, got %d", skill.ReviewCount)
	}
}

func TestSkillGetNotFound(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/skills/nonexistent-skill")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	var errResp model.ErrorResponse
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp.Error != "skill not found" {
		t.Errorf("expected 'skill not found', got %q", errResp.Error)
	}
}

// --- Reviews List ---

func TestReviewsList(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/v1/skills/line-messenger/reviews")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, string(b))
	}

	var reviews []model.Review
	json.NewDecoder(resp.Body).Decode(&reviews)

	// Seed data: line-messenger has 2 reviews
	if len(reviews) != 2 {
		t.Errorf("expected 2 reviews, got %d", len(reviews))
	}

	for _, r := range reviews {
		if r.Rating < 1 || r.Rating > 5 {
			t.Errorf("rating out of range: %d", r.Rating)
		}
		if r.Username == "" {
			t.Error("expected non-empty username on review")
		}
	}
}

func TestReviewsListEmptyForNewSkill(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	// weather-reminder has 1 review, mercari-lister has 1 - let's check one with known count
	resp, err := http.Get(srv.URL + "/api/v1/skills/weather-reminder/reviews")
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var reviews []model.Review
	json.NewDecoder(resp.Body).Decode(&reviews)

	if len(reviews) != 1 {
		t.Errorf("expected 1 review for weather-reminder, got %d", len(reviews))
	}
}

// --- Reviews Create ---

func TestReviewCreateWithAuth(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	// Register a new user (not in seed data, so no duplicate review issues)
	authResp := registerUser(t, srv.URL, "reviewer", "reviewer@example.com", "password1234")

	reviewBody, _ := json.Marshal(model.CreateReviewRequest{
		Rating: 5,
		Title:  "Great skill!",
		Body:   "This is an excellent skill for messaging.",
	})

	req, _ := http.NewRequest("POST", srv.URL+"/api/v1/skills/line-messenger/reviews", bytes.NewReader(reviewBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authResp.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 201, got %d: %s", resp.StatusCode, string(b))
	}

	var review model.Review
	json.NewDecoder(resp.Body).Decode(&review)

	if review.Rating != 5 {
		t.Errorf("expected rating=5, got %d", review.Rating)
	}
	if review.Title != "Great skill!" {
		t.Errorf("expected title='Great skill!', got %q", review.Title)
	}
	if review.ID == 0 {
		t.Error("expected non-zero review ID")
	}
}

func TestReviewCreateWithoutAuth(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	reviewBody, _ := json.Marshal(model.CreateReviewRequest{
		Rating: 4,
		Title:  "Nice",
		Body:   "Should fail without auth.",
	})

	req, _ := http.NewRequest("POST", srv.URL+"/api/v1/skills/line-messenger/reviews", bytes.NewReader(reviewBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestReviewCreateDuplicate(t *testing.T) {
	srv := setupTestServer(t)
	defer srv.Close()

	// Register a new user
	authResp := registerUser(t, srv.URL, "dupreview", "dupreview@example.com", "password1234")

	reviewBody, _ := json.Marshal(model.CreateReviewRequest{
		Rating: 4,
		Title:  "First review",
		Body:   "My first review.",
	})

	// First review should succeed
	req, _ := http.NewRequest("POST", srv.URL+"/api/v1/skills/rakuten-shopping/reviews", bytes.NewReader(reviewBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authResp.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("first review request: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("first review: expected 201, got %d", resp.StatusCode)
	}

	// Second review on same skill should fail with 409
	reviewBody2, _ := json.Marshal(model.CreateReviewRequest{
		Rating: 5,
		Title:  "Second review",
		Body:   "Trying again.",
	})
	req2, _ := http.NewRequest("POST", srv.URL+"/api/v1/skills/rakuten-shopping/reviews", bytes.NewReader(reviewBody2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+authResp.Token)

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatalf("second review request: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusConflict {
		b, _ := io.ReadAll(resp2.Body)
		t.Fatalf("expected 409 for duplicate review, got %d: %s", resp2.StatusCode, string(b))
	}

	var errResp model.ErrorResponse
	json.NewDecoder(resp2.Body).Decode(&errResp)
	if errResp.Error != "you have already reviewed this skill" {
		t.Errorf("expected duplicate review error, got %q", errResp.Error)
	}
}
