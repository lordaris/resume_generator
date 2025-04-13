package handler

import (
	"encoding/json"
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

	RespondWithJSON(w, http.StatusOK, map[string]any{
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

// SavePersonalInfoHandler stores personal information
func (h *ResumeHandler) SavePersonalInfoHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	// Obtain UserID through claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	// Get id from path
	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	// transform resumeID string to uuid
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to update this resume", "FORBIDDEN")
		return
	}

	var personalInfo domain.PersonalInfo
	if err := json.NewDecoder(r.Body).Decode(&personalInfo); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	if err := personalInfo.Validate(); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	personalInfo.BeforeSave()

	if err := h.resumeRepo.SavePersonalInfo(resumeUUID, &personalInfo); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to save personal info", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"message": "Personal info saved successfully",
	})
}

func (h *ResumeHandler) AddEducationHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to update this resume", "FORBIDDEN")
		return
	}

	var education domain.Education
	if err := json.NewDecoder(r.Body).Decode(&education); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	if err := education.Validate(); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	education.BeforeSave()

	educationID, err := h.resumeRepo.AddEducation(resumeUUID, &education)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to add education", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]any{
		"id":      educationID,
		"message": "Education added successfully",
	})
}

func (h *ResumeHandler) GetEducationHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to access this resume", "FORBIDDEN")
		return
	}

	education, err := h.resumeRepo.GetEducationByResume(resumeUUID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get education entries", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, education)
}

func (h *ResumeHandler) DeleteEducationHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	educationID := r.PathValue("educationId")

	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	if educationID == "" {
		RespondWithError(w, http.StatusBadRequest, "Education ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	educationUUID, err := uuid.Parse(educationID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid education ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to update this resume", "FORBIDDEN")
		return
	}

	if err := h.resumeRepo.DeleteEducation(educationUUID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Education entry not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete education entry", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"message": "Education entry deleted successfully",
	})
}

func (h *ResumeHandler) AddExperienceHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to update this resume", "FORBIDDEN")
		return
	}

	var experience domain.Experience
	if err := json.NewDecoder(r.Body).Decode(&experience); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	if err := experience.Validate(); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	experience.BeforeSave()

	experienceID, err := h.resumeRepo.AddExperience(resumeUUID, &experience)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to add experience", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]any{
		"id":      experienceID,
		"message": "Experience added successfully",
	})
}

func (h *ResumeHandler) GetExperienceHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to access this resume", "FORBIDDEN")
		return
	}

	experience, err := h.resumeRepo.GetExperienceByResume(resumeUUID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get experience entries", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, experience)
}

func (h *ResumeHandler) DeleteExperienceHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	experienceID := r.PathValue("experienceId")

	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	if experienceID == "" {
		RespondWithError(w, http.StatusBadRequest, "Experience ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	experienceUUID, err := uuid.Parse(experienceID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid experience ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to update this resume", "FORBIDDEN")
		return
	}

	if err := h.resumeRepo.DeleteExperience(experienceUUID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Experience entry not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete experience entry", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"message": "Experience entry deleted successfully",
	})
}

func (h *ResumeHandler) AddSkillHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to update this resume", "FORBIDDEN")
		return
	}

	var skill domain.Skill
	if err := json.NewDecoder(r.Body).Decode(&skill); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	if err := skill.Validate(); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	skill.BeforeSave()

	skillID, err := h.resumeRepo.AddSkill(resumeUUID, &skill)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to add skill", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]any{
		"id":      skillID,
		"message": "Skill added successfully",
	})
}

func (h *ResumeHandler) GetSkillsHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to access this resume", "FORBIDDEN")
		return
	}

	skills, err := h.resumeRepo.GetSkillsByResume(resumeUUID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get skills", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, skills)
}

func (h *ResumeHandler) DeleteSkillHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	skillID := r.PathValue("skillId")

	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	if skillID == "" {
		RespondWithError(w, http.StatusBadRequest, "Skill ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	skillUUID, err := uuid.Parse(skillID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid skill ID", "INVALID_REQUEST")

		return

	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to update this resume", "FORBIDDEN")
		return
	}

	if err := h.resumeRepo.DeleteSkill(skillUUID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Skill not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete skill", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"message": "Skill deleted successfully",
	})
}

func (h *ResumeHandler) AddProjectHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to update this resume", "FORBIDDEN")
		return
	}

	var project domain.Project
	if err := json.NewDecoder(r.Body).Decode(&project); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	if err := project.Validate(); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	project.BeforeSave()

	projectID, err := h.resumeRepo.AddProject(resumeUUID, &project)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to add project", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]any{
		"id":      projectID,
		"message": "Project added successfully",
	})
}

func (h *ResumeHandler) GetProjectsHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to access this resume", "FORBIDDEN")
		return
	}

	projects, err := h.resumeRepo.GetProjectsByResume(resumeUUID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get projects", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, projects)
}

func (h *ResumeHandler) DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	projectID := r.PathValue("projectId")

	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	if projectID == "" {
		RespondWithError(w, http.StatusBadRequest, "Project ID is required", "INVALID_REQUEST")
		return
	}
	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid project ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to update this resume", "FORBIDDEN")
		return
	}

	if err := h.resumeRepo.DeleteProject(projectUUID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Project not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete project", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"message": "Project deleted successfully",
	})
}

func (h *ResumeHandler) AddCertificationHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to update this resume", "FORBIDDEN")
		return
	}

	var certification domain.Certification
	if err := json.NewDecoder(r.Body).Decode(&certification); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	if err := certification.Validate(); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error(), "VALIDATION_ERROR")
		return
	}

	certification.BeforeSave()

	certificationID, err := h.resumeRepo.AddCertification(resumeUUID, &certification)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to add certification", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusCreated, map[string]any{
		"id":      certificationID,
		"message": "Certification added successfully",
	})
}

func (h *ResumeHandler) GetCertificationsHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to access this resume", "FORBIDDEN")
		return
	}

	certifications, err := h.resumeRepo.GetCertificationsByResume(resumeUUID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get certifications", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, certifications)
}

func (h *ResumeHandler) DeleteCertificationHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	certificationID := r.PathValue("certificationId")

	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	if certificationID == "" {
		RespondWithError(w, http.StatusBadRequest, "Certification ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	certificationUUID, err := uuid.Parse(certificationID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid certification ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to update this resume", "FORBIDDEN")
		return
	}

	if err := h.resumeRepo.DeleteCertification(certificationUUID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Certification not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete certification", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"message": "Certification deleted successfully",
	})
}

func (h *ResumeHandler) GetPersonalInfoHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Invalid user ID", "INTERNAL_SERVER_ERROR")
		return
	}

	resumeID := r.PathValue("id")
	if resumeID == "" {
		RespondWithError(w, http.StatusBadRequest, "Resume ID is required", "INVALID_REQUEST")
		return
	}

	resumeUUID, err := uuid.Parse(resumeID)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid resume ID", "INVALID_REQUEST")
		return
	}

	resume, err := h.resumeRepo.GetResumeByID(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithError(w, http.StatusNotFound, "Resume not found", "NOT_FOUND")
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get resume", "INTERNAL_SERVER_ERROR")
		return
	}

	if resume.UserID != userID && claims.Role != "admin" {
		RespondWithError(w, http.StatusForbidden, "You don't have permission to access this resume", "FORBIDDEN")
		return
	}

	personalInfo, err := h.resumeRepo.GetPersonalInfo(resumeUUID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			RespondWithJSON(w, http.StatusOK, nil)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get personal info", "INTERNAL_SERVER_ERROR")
		return
	}

	RespondWithJSON(w, http.StatusOK, personalInfo)
}
