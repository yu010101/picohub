package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/yu01/picohub/internal/middleware"
	"github.com/yu01/picohub/internal/model"
	"github.com/yu01/picohub/internal/repository"
	"github.com/yu01/picohub/internal/scanner"
	"github.com/yu01/picohub/internal/service"
)

type SkillHandler struct {
	skillRepo  *repository.SkillRepository
	storage    *service.StorageService
	scanner    scanner.Scanner
}

func NewSkillHandler(skillRepo *repository.SkillRepository, storage *service.StorageService, sc scanner.Scanner) *SkillHandler {
	return &SkillHandler{skillRepo: skillRepo, storage: storage, scanner: sc}
}

func (h *SkillHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 {
		perPage = 12
	}

	params := model.SkillListParams{
		Query:    r.URL.Query().Get("q"),
		Category: r.URL.Query().Get("category"),
		Sort:     r.URL.Query().Get("sort"),
		Page:     page,
		PerPage:  perPage,
	}

	skills, total, err := h.skillRepo.List(params)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "failed to list skills"})
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	writeJSON(w, http.StatusOK, model.PaginatedResponse{
		Data:       skills,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	})
}

func (h *SkillHandler) Featured(w http.ResponseWriter, r *http.Request) {
	skills, err := h.skillRepo.Featured()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get featured skills"})
		return
	}
	writeJSON(w, http.StatusOK, skills)
}

func (h *SkillHandler) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	skill, err := h.skillRepo.FindBySlug(slug)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "skill not found"})
		return
	}
	writeJSON(w, http.StatusOK, skill)
}

func (h *SkillHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, model.ErrorResponse{Error: "unauthorized"})
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "file too large or invalid form"})
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "file is required"})
		return
	}
	defer file.Close()

	filePath, fileHash, manifest, err := h.storage.SaveFile(file, "upload.zip")
	if err != nil {
		if errors.Is(err, service.ErrFileTooLarge) {
			writeJSON(w, http.StatusRequestEntityTooLarge, model.ErrorResponse{Error: "file exceeds 10MB limit"})
			return
		}
		if errors.Is(err, service.ErrNoManifest) || errors.Is(err, service.ErrInvalidPackage) {
			writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
			return
		}
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "failed to save file"})
		return
	}

	// Scan
	scanResult, err := h.scanner.Scan(filePath)
	scanStatus := "clean"
	if err != nil || !scanResult.Clean {
		scanStatus = "flagged"
		h.storage.Delete(filePath)
		writeJSON(w, http.StatusBadRequest, model.ErrorResponse{Error: "file flagged by security scan"})
		return
	}

	tagsJSON, _ := json.Marshal(manifest.Tags)

	skill := &model.Skill{
		Slug:        manifest.Slug,
		Name:        manifest.Name,
		Description: manifest.Description,
		Version:     manifest.Version,
		Category:    manifest.Category,
		AuthorID:    userID,
		FilePath:    filePath,
		FileHash:    fileHash,
		ScanStatus:  scanStatus,
		Tags:        string(tagsJSON),
	}

	if err := h.skillRepo.Create(skill); err != nil {
		h.storage.Delete(filePath)
		writeJSON(w, http.StatusConflict, model.ErrorResponse{Error: "skill with this slug already exists"})
		return
	}

	writeJSON(w, http.StatusCreated, skill)
}

func (h *SkillHandler) Download(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	skill, err := h.skillRepo.FindBySlug(slug)
	if err != nil {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "skill not found"})
		return
	}

	if skill.FilePath == "" {
		writeJSON(w, http.StatusNotFound, model.ErrorResponse{Error: "no file available for download"})
		return
	}

	h.skillRepo.IncrementDownload(slug)

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", "attachment; filename="+skill.Slug+"-"+skill.Version+".zip")
	http.ServeFile(w, r, skill.FilePath)
}

func (h *SkillHandler) Categories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.skillRepo.Categories()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, model.ErrorResponse{Error: "failed to get categories"})
		return
	}
	writeJSON(w, http.StatusOK, cats)
}
