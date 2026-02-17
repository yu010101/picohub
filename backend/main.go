package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/yu01/picohub/internal/config"
	"github.com/yu01/picohub/internal/database"
	"github.com/yu01/picohub/internal/handler"
	"github.com/yu01/picohub/internal/middleware"
	"github.com/yu01/picohub/internal/repository"
	"github.com/yu01/picohub/internal/scanner"
	"github.com/yu01/picohub/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := database.Seed(db); err != nil {
		log.Printf("Warning: seed failed: %v", err)
	}

	// Repositories
	userRepo := repository.NewUserRepository(db)
	skillRepo := repository.NewSkillRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	// Services
	authService := service.NewAuthService(userRepo, cfg)
	storageService := service.NewStorageService(cfg.UploadDir, cfg.MaxUploadSize)
	noopScanner := scanner.NewNoopScanner()

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	skillHandler := handler.NewSkillHandler(skillRepo, storageService, noopScanner)
	reviewHandler := handler.NewReviewHandler(reviewRepo, skillRepo)

	// Router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.Logger)
	r.Use(middleware.CORS(cfg.AllowedOrigins))

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", handler.Health)

		// Auth routes
		r.Route("/auth", func(r chi.Router) {
			r.With(middleware.RateLimit(3, time.Hour)).Post("/register", authHandler.Register)
			r.With(middleware.RateLimit(5, time.Minute)).Post("/login", authHandler.Login)
			r.With(middleware.Auth(cfg.JWTSecret)).Get("/me", authHandler.Me)
		})

		// Skill routes
		r.Route("/skills", func(r chi.Router) {
			r.With(middleware.RateLimit(100, time.Minute)).Get("/", skillHandler.List)
			r.Get("/featured", skillHandler.Featured)
			r.Get("/categories", skillHandler.Categories)
			r.With(middleware.Auth(cfg.JWTSecret)).Post("/", skillHandler.Create)

			r.Route("/{slug}", func(r chi.Router) {
				r.Get("/", skillHandler.Get)
				r.Get("/download", skillHandler.Download)
				r.Get("/reviews", reviewHandler.List)
				r.With(middleware.Auth(cfg.JWTSecret)).Post("/reviews", reviewHandler.Create)
			})
		})
	})

	log.Printf("PicoHub API server starting on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
