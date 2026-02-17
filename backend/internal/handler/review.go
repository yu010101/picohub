package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/yu01/picohub/internal/middleware"
	"github.com/yu01/picohub/internal/model"
	"github.com/yu01/picohub/internal/repository"
)

type ReviewHandler struct {
	reviewRepo *repository.ReviewRepository
	skillRepo  *repository.SkillRepository
}

func NewReviewHandler(reviewRepo *repository.ReviewRepository, skillRepo *repository.SkillRepository) *ReviewHandler {
	return &ReviewHandler{reviewRepo: reviewRepo, skillRepo: skillRepo}
}

func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	skill, err := h.skillRepo.FindBySlug(slug)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "skill not found"})
		return
	}

	reviews, err := h.reviewRepo.ListBySkillID(skill.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "failed to list reviews"})
		return
	}

	if reviews == nil {
		reviews = []model.Review{}
	}

	writeJSON(w, http.StatusOK, reviews)
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	slug := chi.URLParam(r, "slug")
	skill, err := h.skillRepo.FindBySlug(slug)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "skill not found"})
		return
	}

	var req model.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "rating must be between 1 and 5"})
		return
	}

	review := &model.Review{
		SkillID: skill.ID,
		UserID:  userID,
		Rating:  req.Rating,
		Title:   req.Title,
		Body:    req.Body,
	}

	if err := h.reviewRepo.Create(review); err != nil {
		writeJSON(w, http.StatusConflict, model.ErrorResponse{Error: "you have already reviewed this skill"})
		return
	}

	writeJSON(w, http.StatusCreated, review)
}
