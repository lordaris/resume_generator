package handler

import (
	"net/http"

	"github.com/lordaris/resume_generator/internal/domain"
)

// AdminHandler handles admin-related requests
type AdminHandler struct {
	userRepo domain.UserRepository
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(userRepo domain.UserRepository) *AdminHandler {
	return &AdminHandler{
		userRepo: userRepo,
	}
}

// GetUsersHandler handles fetching all users (admin only)
func (h *AdminHandler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := GetClaimsFromContext(r.Context())
	if err != nil {
		RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	RespondWithJSON(w, http.StatusOK, map[string]any{
		"message":  "Admin route accessed successfully",
		"admin_id": claims.UserID,
		"email":    claims.Email,
	})
}
