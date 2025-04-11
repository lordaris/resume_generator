package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/lordaris/resume_generator/internal/domain"
	"github.com/lordaris/resume_generator/internal/repository"
	"github.com/rs/zerolog/log"
)

// UserHandler handles user-related requests
type UserHandler struct {
	userRepo   domain.UserRepository
	resumeRepo domain.ResumeRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(userRepo domain.UserRepository, resumeRepo domain.ResumeRepository) *UserHandler {
	return &UserHandler{
		userRepo:   userRepo,
		resumeRepo: resumeRepo,
	}
}

// GetProfileHandler handles fetching the user profile
func (h *UserHandler) GetProfileHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get user from repository
	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "User not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get user profile", "INTERNAL_SERVER_ERROR")
		return
	}

	// Get user's resumes
	resumes, err := h.resumeRepo.GetResumesByUserID(userID)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID.String()).Msg("Failed to get user resumes")
		// Continue without resumes
	}

	// Create response with user profile and resume list
	type resumeInfo struct {
		ID        string `json:"id"`
		CreatedAt string `json:"created_at"`
	}

	type profileResponse struct {
		UserID    string       `json:"user_id"`
		Email     string       `json:"email"`
		Role      string       `json:"role"`
		CreatedAt string       `json:"created_at"`
		Resumes   []resumeInfo `json:"resumes,omitempty"`
	}

	response := profileResponse{
		UserID:    user.ID.String(),
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if resumes != nil {
		resumeList := make([]resumeInfo, len(resumes))
		for i, resume := range resumes {
			resumeList[i] = resumeInfo{
				ID:        resume.ID.String(),
				CreatedAt: resume.CreatedAt.Format("2006-01-02T15:04:05Z"),
			}
		}
		response.Resumes = resumeList
	}

	RespondWithJSON(w, http.StatusOK, response)
}
