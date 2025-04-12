package main

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/lordaris/resume_generator/internal/handler"
	"github.com/lordaris/resume_generator/internal/repository"
	"github.com/lordaris/resume_generator/internal/service"
	"github.com/lordaris/resume_generator/pkg/auth"
	"github.com/lordaris/resume_generator/pkg/security"
	"github.com/redis/go-redis/v9"
)

// setupRoutes configures and returns the router with all routes
func setupRoutes(db *sqlx.DB, redisClient *redis.Client, jwtConfig auth.JWTConfig) http.Handler {
	// Create router
	mux := http.NewServeMux()

	// Create repositories
	userRepo := repository.NewPostgresUserRepository(db)
	resumeRepo := repository.NewPostgresResumeRepository(db)

	// Create JWT handler
	jwtHandler := auth.NewJWT(jwtConfig)

	// Create services
	authServiceConfig := service.AuthServiceConfig{
		AccessTokenExpiry:  jwtConfig.AccessTokenExpiry,
		RefreshTokenExpiry: jwtConfig.RefreshTokenExpiry,
		ResetTokenExpiry:   jwtConfig.ResetTokenExpiry,
	}
	authService := service.NewAuthService(userRepo, jwtHandler, authServiceConfig)

	// Create middleware
	authMiddleware := handler.NewAuthMiddleware(authService)
	sessionLogger := handler.NewSessionLogger()

	// Create handlers
	authHandler := handler.NewAuthHandler(authService, redisClient)
	userHandler := handler.NewUserHandler(userRepo, resumeRepo)
	resumeHandler := handler.NewResumeHandler(resumeRepo)
	adminHandler := handler.NewAdminHandler(userRepo)

	// Public routes
	mux.HandleFunc("POST /api/v1/register", authHandler.RegisterHandler)
	mux.HandleFunc("POST /api/v1/login", authHandler.LoginHandler)
	mux.HandleFunc("POST /api/v1/refresh-token", authHandler.RefreshTokenHandler)
	mux.HandleFunc("POST /api/v1/logout", authHandler.LogoutHandler)
	mux.HandleFunc("POST /api/v1/request-password-reset", authHandler.RequestPasswordResetHandler)
	mux.HandleFunc("POST /api/v1/reset-password", authHandler.ResetPasswordHandler)

	// User profile route
	mux.Handle("GET /api/v1/user/profile", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(userHandler.GetProfileHandler))))

	// Admin route
	mux.Handle("GET /api/v1/admin/users", sessionLogger.LogActivity(authMiddleware.AuthRequired(authMiddleware.RequireRole("admin")(http.HandlerFunc(adminHandler.GetUsersHandler)))))

	// Resume routes
	mux.Handle("GET /api/v1/resumes", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.GetResumeListHandler))))
	mux.Handle("GET /api/v1/resumes/{id}", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.GetResumeHandler))))
	mux.Handle("POST /api/v1/resumes", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.CreateResumeHandler))))
	mux.Handle("DELETE /api/v1/resumes/{id}", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.DeleteResumeHandler))))

	corsMiddleware := security.CORSMiddleware(security.DefaultCORSConfig())

	// Wrap the entire router with CORS middleware
	handlerWithCORS := corsMiddleware(mux)

	return handlerWithCORS
}

// Example protected handler
func userProfileHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := handler.GetClaimsFromContext(r.Context())
	if err != nil {
		handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	handler.RespondWithJSON(w, http.StatusOK, map[string]any{
		"message": "Protected route",

		"user_id": claims.UserID,
		"email":   claims.Email,
		"role":    claims.Role,
	})
}

// Example admin handler
func adminUsersHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := handler.GetClaimsFromContext(r.Context())
	if err != nil {
		handler.RespondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	handler.RespondWithJSON(w, http.StatusOK, map[string]any{
		"message":  "Admin route",
		"admin_id": claims.UserID,
		"email":    claims.Email,
	})
}
