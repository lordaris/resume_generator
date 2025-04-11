package main

import (
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/lordaris/resume_generator/internal/handler"
	"github.com/lordaris/resume_generator/internal/repository"
	"github.com/lordaris/resume_generator/internal/service"
	"github.com/lordaris/resume_generator/pkg/auth"
	"github.com/redis/go-redis/v9"
)

// setupRoutes configures and returns the router with all routes
func setupRoutes(db *sqlx.DB, redisClient *redis.Client, jwtConfig auth.JWTConfig) http.Handler {
	// Create router
	mux := http.NewServeMux()

	// Create repositories
	userRepo := repository.NewPostgresUserRepository(db)

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

	// Public routes
	mux.HandleFunc("POST /api/v1/register", authHandler.RegisterHandler)
	mux.HandleFunc("POST /api/v1/login", authHandler.LoginHandler)
	mux.HandleFunc("POST /api/v1/refresh-token", authHandler.RefreshTokenHandler)
	mux.HandleFunc("POST /api/v1/logout", authHandler.LogoutHandler)
	mux.HandleFunc("POST /api/v1/request-password-reset", authHandler.RequestPasswordResetHandler)
	mux.HandleFunc("POST /api/v1/reset-password", authHandler.ResetPasswordHandler)

	// Protected routes - need authentication
	apiHandler := sessionLogger.LogActivity(authMiddleware.AuthRequired(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This handler delegates to specific routes based on the request path
		// We'll use Go 1.22's new handler pattern matching
		switch r.URL.Path {
		case "/api/v1/user/profile":
			// Example protected route
			userProfileHandler(w, r)
		default:
			http.NotFound(w, r)
		}
	})))

	mux.Handle("GET /api/v1/user/profile", apiHandler)

	// Admin routes - need authentication and admin role
	adminHandler := sessionLogger.LogActivity(authMiddleware.AuthRequired(
		authMiddleware.RequireRole("admin")(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Admin handlers
				switch r.URL.Path {
				case "/api/v1/admin/users":
					// Example admin route
					adminUsersHandler(w, r)
				default:
					http.NotFound(w, r)
				}
			}),
		),
	))

	mux.Handle("GET /api/v1/admin/users", adminHandler)

	return mux
}

// Example protected handler
func userProfileHandler(w http.ResponseWriter, r *http.Request) {
	claims, err := handler.GetClaimsFromContext(r.Context())
	if err != nil {
		handler.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	handler.respondWithJSON(w, http.StatusOK, map[string]interface{}{
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
		handler.respondWithError(w, http.StatusUnauthorized, "Unauthorized", "UNAUTHORIZED")
		return
	}

	handler.respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"message":  "Admin route",
		"admin_id": claims.UserID,
		"email":    claims.Email,
	})
}
