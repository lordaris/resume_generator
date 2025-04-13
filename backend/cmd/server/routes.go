package main

import (
	"net/http"
	"time"

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
	corsMiddleware := security.CORSMiddleware(security.DefaultCORSConfig())
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
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		handler.RespondWithJSON(w, http.StatusOK, map[string]any{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})
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
	mux.Handle("GET /api/v1/resumes/{id}/personal-info", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.GetPersonalInfoHandler))))
	mux.Handle("PUT /api/v1/resumes/{id}/personal-info", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.SavePersonalInfoHandler))))
	mux.Handle("GET /api/v1/resumes/{id}/education", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.GetEducationHandler))))
	mux.Handle("POST /api/v1/resumes/{id}/education", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.AddEducationHandler))))
	mux.Handle("DELETE /api/v1/resumes/{id}/education/{educationId}", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.DeleteEducationHandler))))
	mux.Handle("GET /api/v1/resumes/{id}/experience", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.GetExperienceHandler))))
	mux.Handle("POST /api/v1/resumes/{id}/experience", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.AddExperienceHandler))))
	mux.Handle("DELETE /api/v1/resumes/{id}/experience/{experienceId}", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.DeleteExperienceHandler))))
	mux.Handle("GET /api/v1/resumes/{id}/skills", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.GetSkillsHandler))))
	mux.Handle("POST /api/v1/resumes/{id}/skills", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.AddSkillHandler))))
	mux.Handle("DELETE /api/v1/resumes/{id}/skills/{skillId}", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.DeleteSkillHandler))))
	mux.Handle("GET /api/v1/resumes/{id}/projects", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.GetProjectsHandler))))
	mux.Handle("POST /api/v1/resumes/{id}/projects", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.AddProjectHandler))))
	mux.Handle("DELETE /api/v1/resumes/{id}/projects/{projectId}", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.DeleteProjectHandler))))
	mux.Handle("GET /api/v1/resumes/{id}/certifications", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.GetCertificationsHandler))))
	mux.Handle("POST /api/v1/resumes/{id}/certifications", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.AddCertificationHandler))))
	mux.Handle("DELETE /api/v1/resumes/{id}/certifications/{certificationId}", sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(resumeHandler.DeleteCertificationHandler))))

	// Wrap the entire router with CORS middleware
	handlerWithCORS := corsMiddleware(mux)

	return handlerWithCORS
}
