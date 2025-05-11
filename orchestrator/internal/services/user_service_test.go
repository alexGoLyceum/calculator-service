package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"unicode"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/repository"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/services"
	"github.com/alexGoLyceum/calculator-service/orchestrator/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockJWT := mocks.NewMockJWTManager(ctrl)

	userService := services.NewUserService(mockRepo, mockJWT)

	tests := []struct {
		name        string
		login       string
		password    string
		mockSetup   func()
		expected    string
		expectedErr error
	}{
		{
			name:     "successful registration",
			login:    "valid@login.com",
			password: "ValidPass123!",
			mockSetup: func() {
				userID := uuid.New()
				mockRepo.EXPECT().CreateUser(gomock.Any(), "valid@login.com", "ValidPass123!").Return(userID, nil)
				mockJWT.EXPECT().Generate(userID).Return("generated-token", nil)
			},
			expected:    "generated-token",
			expectedErr: nil,
		},
		{
			name:        "invalid login - empty",
			login:       "",
			password:    "ValidPass123!",
			mockSetup:   func() {},
			expected:    "",
			expectedErr: services.ErrInvalidLogin,
		},
		{
			name:        "invalid login - too long",
			login:       "thisloginistoolongthisloginistoolongthisloginistoolong",
			password:    "ValidPass123!",
			mockSetup:   func() {},
			expected:    "",
			expectedErr: services.ErrInvalidLogin,
		},
		{
			name:        "weak password - too short",
			login:       "valid@login.com",
			password:    "Short1!",
			mockSetup:   func() {},
			expected:    "",
			expectedErr: services.ErrWeakPassword,
		},
		{
			name:        "weak password - no uppercase",
			login:       "valid@login.com",
			password:    "lowercase123!",
			mockSetup:   func() {},
			expected:    "",
			expectedErr: services.ErrWeakPassword,
		},
		{
			name:        "weak password - no lowercase",
			login:       "valid@login.com",
			password:    "UPPERCASE123!",
			mockSetup:   func() {},
			expected:    "",
			expectedErr: services.ErrWeakPassword,
		},
		{
			name:        "weak password - no digit",
			login:       "valid@login.com",
			password:    "NoDigitsHere!",
			mockSetup:   func() {},
			expected:    "",
			expectedErr: services.ErrWeakPassword,
		},
		{
			name:        "weak password - no special char",
			login:       "valid@login.com",
			password:    "NoSpecial123",
			mockSetup:   func() {},
			expected:    "",
			expectedErr: services.ErrWeakPassword,
		},
		{
			name:     "user already exists",
			login:    "existing@user.com",
			password: "ValidPass123!",
			mockSetup: func() {
				mockRepo.EXPECT().CreateUser(gomock.Any(), "existing@user.com", "ValidPass123!").Return(uuid.Nil, repository.ErrUserWithLoginAlreadyExists)
			},
			expected:    "",
			expectedErr: services.ErrUserWithLoginAlreadyExists,
		},
		{
			name:     "database unavailable",
			login:    "valid@login.com",
			password: "ValidPass123!",
			mockSetup: func() {
				mockRepo.EXPECT().CreateUser(gomock.Any(), "valid@login.com", "ValidPass123!").Return(uuid.Nil, repository.ErrDatabaseNotAvailable)
			},
			expected:    "",
			expectedErr: services.ErrDatabaseUnavailable,
		},
		{
			name:     "jwt generation error",
			login:    "valid@login.com",
			password: "ValidPass123!",
			mockSetup: func() {
				userID := uuid.New()
				mockRepo.EXPECT().CreateUser(gomock.Any(), "valid@login.com", "ValidPass123!").Return(userID, nil)
				mockJWT.EXPECT().Generate(userID).Return("", errors.New("jwt error"))
			},
			expected:    "",
			expectedErr: errors.New("error generating token: jwt error"),
		},
		{
			name:     "unexpected repository error",
			login:    "valid@login.com",
			password: "ValidPass123!",
			mockSetup: func() {
				mockRepo.EXPECT().CreateUser(gomock.Any(), "valid@login.com", "ValidPass123!").Return(uuid.Nil, errors.New("unexpected repo error"))
			},
			expected:    "",
			expectedErr: errors.New("unexpected repo error"),
		},
		{
			name:     "unexpected jwt error with custom message",
			login:    "valid@login.com",
			password: "ValidPass123!",
			mockSetup: func() {
				userID := uuid.New()
				mockRepo.EXPECT().CreateUser(gomock.Any(), "valid@login.com", "ValidPass123!").Return(userID, nil)
				mockJWT.EXPECT().Generate(userID).Return("", errors.New("custom jwt error"))
			},
			expected:    "",
			expectedErr: errors.New("error generating token: custom jwt error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			token, err := userService.Register(context.Background(), tt.login, tt.password)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, token)
		})
	}
}

func TestUserService_Authenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockRepository(ctrl)
	mockJWT := mocks.NewMockJWTManager(ctrl)

	userService := services.NewUserService(mockRepo, mockJWT)

	tests := []struct {
		name        string
		login       string
		password    string
		mockSetup   func()
		expected    string
		expectedErr error
	}{
		{
			name:     "successful authentication",
			login:    "valid@login.com",
			password: "ValidPass123!",
			mockSetup: func() {
				userID := uuid.New()
				mockRepo.EXPECT().AuthUser(gomock.Any(), "valid@login.com", "ValidPass123!").Return(userID, nil)
				mockJWT.EXPECT().Generate(userID).Return("generated-token", nil)
			},
			expected:    "generated-token",
			expectedErr: nil,
		},
		{
			name:        "invalid login",
			login:       "",
			password:    "ValidPass123!",
			mockSetup:   func() {},
			expected:    "",
			expectedErr: services.ErrInvalidLogin,
		},
		{
			name:     "user not found",
			login:    "nonexistent@user.com",
			password: "ValidPass123!",
			mockSetup: func() {
				mockRepo.EXPECT().AuthUser(gomock.Any(), "nonexistent@user.com", "ValidPass123!").Return(uuid.Nil, repository.ErrUserNotFoundByLogin)
			},
			expected:    "",
			expectedErr: services.ErrUserNotFoundByLogin,
		},
		{
			name:     "invalid password",
			login:    "valid@login.com",
			password: "WrongPass123!",
			mockSetup: func() {
				mockRepo.EXPECT().AuthUser(gomock.Any(), "valid@login.com", "WrongPass123!").Return(uuid.Nil, repository.ErrInvalidPassword)
			},
			expected:    "",
			expectedErr: services.ErrInvalidPassword,
		},
		{
			name:     "database unavailable",
			login:    "valid@login.com",
			password: "ValidPass123!",
			mockSetup: func() {
				mockRepo.EXPECT().AuthUser(gomock.Any(), "valid@login.com", "ValidPass123!").Return(uuid.Nil, repository.ErrDatabaseNotAvailable)
			},
			expected:    "",
			expectedErr: services.ErrDatabaseUnavailable,
		},
		{
			name:     "jwt generation error",
			login:    "valid@login.com",
			password: "ValidPass123!",
			mockSetup: func() {
				userID := uuid.New()
				mockRepo.EXPECT().AuthUser(gomock.Any(), "valid@login.com", "ValidPass123!").Return(userID, nil)
				mockJWT.EXPECT().Generate(userID).Return("", errors.New("jwt error"))
			},
			expected:    "",
			expectedErr: errors.New("error generating token: jwt error"),
		},
		{
			name:     "unexpected repository error",
			login:    "valid@login.com",
			password: "ValidPass123!",
			mockSetup: func() {
				mockRepo.EXPECT().AuthUser(gomock.Any(), "valid@login.com", "ValidPass123!").Return(uuid.Nil, errors.New("unexpected repo error"))
			},
			expected:    "",
			expectedErr: errors.New("unexpected repo error"),
		},
		{
			name:     "unexpected jwt error with custom message",
			login:    "valid@login.com",
			password: "ValidPass123!",
			mockSetup: func() {
				userID := uuid.New()
				mockRepo.EXPECT().AuthUser(gomock.Any(), "valid@login.com", "ValidPass123!").Return(userID, nil)
				mockJWT.EXPECT().Generate(userID).Return("", errors.New("custom jwt error"))
			},
			expected:    "",
			expectedErr: errors.New("error generating token: custom jwt error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			token, err := userService.Authenticate(context.Background(), tt.login, tt.password)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, token)
		})
	}
}

func TestIsValidLogin(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		expected bool
	}{
		{"valid login", "user@example.com", true},
		{"empty login", "", false},
		{"too long login", "thisloginistoolongthisloginistoolongthisloginistoolong", false},
		{"boundary length", strings.Repeat("a", 31), true},
		{"exceeds boundary", strings.Repeat("a", 32), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, services.IsValidLogin(tt.login))
		})
	}
}

func TestIsValidPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected bool
	}{
		{"valid password", "ValidPass123!", true},
		{"too short", "Short1!", false},
		{"too long", strings.Repeat("a", 26) + "A1!", false},
		{"no uppercase", "lowercase123!", false},
		{"no lowercase", "UPPERCASE123!", false},
		{"no digit", "NoDigitsHere!", false},
		{"no special char", "NoSpecial123", false},
		{"boundary length min", "A1!abcde", true},
		{"boundary length max", strings.Repeat("a", 22) + "A1!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, services.IsValidPassword(tt.password))
		})
	}
}

func TestPasswordComplexity(t *testing.T) {
	tests := []struct {
		name       string
		password   string
		hasUpper   bool
		hasLower   bool
		hasDigit   bool
		hasSpecial bool
	}{
		{"all requirements", "A1!bcdef", true, true, true, true},
		{"no upper", "a1!bcdef", false, true, true, true},
		{"no lower", "A1!BCDEF", true, false, true, true},
		{"no digit", "A!bcdefg", true, true, false, true},
		{"no special", "A1bcdefg", true, true, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var hasUpper, hasLower, hasDigit, hasSpecial bool
			for _, c := range tt.password {
				switch {
				case unicode.IsUpper(c):
					hasUpper = true
				case unicode.IsLower(c):
					hasLower = true
				case unicode.IsDigit(c):
					hasDigit = true
				case strings.ContainsRune("!@#$%?&*", c):
					hasSpecial = true
				}
			}

			assert.Equal(t, tt.hasUpper, hasUpper, "uppercase check failed")
			assert.Equal(t, tt.hasLower, hasLower, "lowercase check failed")
			assert.Equal(t, tt.hasDigit, hasDigit, "digit check failed")
			assert.Equal(t, tt.hasSpecial, hasSpecial, "special char check failed")
		})
	}
}
