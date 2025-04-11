package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/lordaris/resume_generator/internal/domain"
	"github.com/lordaris/resume_generator/internal/service"
	"github.com/lordaris/resume_generator/pkg/auth"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock user repository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserByID(id uuid.UUID) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) CreateSession(session *domain.Session) error {
	args := m.Called(session)
	return args.Error(0)
}

func (m *MockUserRepository) GetSessionByID(id uuid.UUID) (*domain.Session, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockUserRepository) GetSessionByToken(token string) (*domain.Session, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *MockUserRepository) DeleteSession(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUserSessions(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepository) CreatePasswordReset(reset *domain.PasswordReset) error {
	args := m.Called(reset)
	return args.Error(0)
}

func (m *MockUserRepository) GetPasswordResetByToken(token string) (*domain.PasswordReset, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PasswordReset), args.Error(1)
}

func (m *MockUserRepository) MarkPasswordResetUsed(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteExpiredPasswordResets() error {
	args := m.Called()
	return args.Error(0)
}

// Test setup helper
func setupTest(t *testing.T) (*AuthHandler, *MockUserRepository, *miniredis.Miniredis) {
	// Create mock repository
	mockRepo := new(MockUserRepository)

	// Create JWT handler
	jwtConfig := auth.JWTConfig{
		Secret:             "test-secret",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 24 * time.Hour,
		ResetTokenExpiry:   1 * time.Hour,
	}
	jwtHandler := auth.NewJWT(jwtConfig)

	// Create auth service
	authServiceConfig := service.AuthServiceConfig{
		AccessTokenExpiry:  jwtConfig.AccessTokenExpiry,
		RefreshTokenExpiry: jwtConfig.RefreshTokenExpiry,
		ResetTokenExpiry:   jwtConfig.ResetTokenExpiry,
	}
	authService := service.NewAuthService(mockRepo, jwtHandler, authServiceConfig)

	// Create miniredis for testing
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}

	// Create Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create auth handler
	authHandler := NewAuthHandler(authService, redisClient)

	return authHandler, mockRepo, mr
}

func TestRegisterHandler(t *testing.T) {
	// Setup test
	handler, mockRepo, mr := setupTest(t)
	defer mr.Close()

	// Test data
	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful registration",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			setupMock: func() {
				mockRepo.On("GetUserByEmail", "test@example.com").Return(nil, errors.New("not found"))
				mockRepo.On("CreateUser", mock.AnythingOfType("*domain.User")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"message":"User registered successfully"`,
		},
		{
			name: "User already exists",
			requestBody: map[string]interface{}{
				"email":    "existing@example.com",
				"password": "password123",
			},
			setupMock: func() {
				user := &domain.User{
					ID:    uuid.New(),
					Email: "existing@example.com",
				}
				mockRepo.On("GetUserByEmail", "existing@example.com").Return(user, nil)
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   `{"error":"User with this email already exists","code":"USER_EXISTS"}`,
		},
		{
			name: "Invalid email",
			requestBody: map[string]interface{}{
				"email":    "invalid-email",
				"password": "password123",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Validation failed","code":"VALIDATION_FAILED","details":{"fields":{"email":"Must be a valid email address"}}}`,
		},
		{
			name: "Password too short",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "short",
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"Validation failed","code":"VALIDATION_FAILED","details":{"fields":{"password":"Must be at least 8 characters long"}}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			tc.setupMock()

			// Create request
			jsonBody, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.RegisterHandler(rr, req)

			// Check response
			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)

			// Reset mock
			mockRepo.ExpectedCalls = nil
		})
	}
}

func TestLoginHandler(t *testing.T) {
	// Setup test
	handler, mockRepo, mr := setupTest(t)
	defer mr.Close()

	// Create test user
	testUserID := uuid.New()
	testUser := &domain.User{
		ID:           testUserID,
		Email:        "test@example.com",
		PasswordHash: "$argon2id$v=19$m=65536,t=3,p=4$c29tZXNhbHQ$NkOaJEF81zhyjzEgPpiFUzdmVL8/TxHIxc1qQB/qKj4", // "password123"
		Role:         "user",
	}

	// Test data
	testCases := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMock      func()
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "Successful login",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "password123",
			},
			setupMock: func() {
				mockRepo.On("GetUserByEmail", "test@example.com").Return(testUser, nil)
				mockRepo.On("CreateSession", mock.AnythingOfType("*domain.Session")).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"access_token"`,
		},
		{
			name: "Invalid email",
			requestBody: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "password123",
			},
			setupMock: func() {
				mockRepo.On("GetUserByEmail", "nonexistent@example.com").Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid email or password","code":"INVALID_CREDENTIALS"}`,
		},
		{
			name: "Invalid password",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			setupMock: func() {
				mockRepo.On("GetUserByEmail", "test@example.com").Return(testUser, nil)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   `{"error":"Invalid email or password","code":"INVALID_CREDENTIALS"}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mock
			tc.setupMock()

			// Create request
			jsonBody, _ := json.Marshal(tc.requestBody)
			req, _ := http.NewRequest("POST", "/api/v1/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.LoginHandler(rr, req)

			// Check response
			assert.Equal(t, tc.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tc.expectedBody)

			// Reset mock
			mockRepo.ExpectedCalls = nil
		})
	}
}
