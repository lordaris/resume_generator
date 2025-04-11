package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/lordaris/resume_generator/internal/domain"
	"github.com/lordaris/resume_generator/internal/repository"
)

// ResumeHandler handles resume-related requests
type ResumeHandler struct {
	resumeRepo domain.ResumeRepository
}

// NewResumeHandler creates a new resume handler
func NewResumeHandler(resumeRepo domain.ResumeRepository) *ResumeHandler {
	return &ResumeHandler{
		resumeRepo: resumeRepo,
	}
}

// GetResumeHandler handles fetching a single resume
func (h *ResumeHandler) GetResumeHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	// Get user ID from claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	// Get resume ID from path parameter
	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	// Parse resume ID
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	// Get resume from repository
	resume, err := h.resumeRepo.GetCompleteResume(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	// Check if the resume belongs to the user
	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to access this resume", "FORBIDDEN")
		return
	}

	RespondWithJSON(w, http.StatusOK, resume)
}

// CreateResumeHandler handles creating a new resume
func (h *ResumeHandler) CreateResumeHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	// Get user ID from claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	// Create a new resume
	resume, err := h.resumeRepo.CreateResume(userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to create resume", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusCreated, resume)
}

// DeleteResumeHandler handles deleting a resume
func (h *ResumeHandler) DeleteResumeHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	// Get user ID from claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	// Get resume ID from path parameter
	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	// Parse resume ID
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	// Get resume to check ownership
	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	// Check if the resume belongs to the user
	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to delete this resume", "FORBIDDEN")
		return
	}

	// Delete the resume
	if err := h.resumeRepo.DeleteResume(resumeUUID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete resume", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Resume deleted successfully",
	})
}

// GetResumeListHandler handles fetching all resumes for a user
func (h *ResumeHandler) GetResumeListHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	// Get user ID from claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	// Get resumes from repository
	resumes, err := h.resumeRepo.GetResumesByUserID(userID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resumes", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, resumes)
}
