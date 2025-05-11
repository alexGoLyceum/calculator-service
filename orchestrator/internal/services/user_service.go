package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/auth"
	"github.com/alexGoLyceum/calculator-service/orchestrator/internal/repository"
)

type UserService interface {
	Register(ctx context.Context, login, password string) (string, error)
	Authenticate(ctx context.Context, login, password string) (string, error)
}

type userService struct {
	repo repository.Repository
	jwt  auth.JWTManager
}

func NewUserService(repo repository.Repository, jwt auth.JWTManager) UserService {
	return &userService{
		repo: repo,
		jwt:  jwt,
	}
}

func (s *userService) Register(ctx context.Context, login, password string) (string, error) {
	if !IsValidLogin(login) {
		return "", ErrInvalidLogin
	}

	if !IsValidPassword(password) {
		return "", ErrWeakPassword
	}

	userID, err := s.repo.CreateUser(ctx, login, password)
	if err != nil {
		if errors.Is(err, repository.ErrUserWithLoginAlreadyExists) {
			return "", ErrUserWithLoginAlreadyExists
		}
		if errors.Is(err, repository.ErrDatabaseNotAvailable) {
			return "", ErrDatabaseUnavailable
		}
		return "", err
	}

	token, err := s.jwt.Generate(userID)
	if err != nil {
		return "", fmt.Errorf("error generating token: %v", err)
	}
	return token, nil
}

func (s *userService) Authenticate(ctx context.Context, login, password string) (string, error) {
	if !IsValidLogin(login) {
		return "", ErrInvalidLogin
	}

	userID, err := s.repo.AuthUser(ctx, login, password)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFoundByLogin) {
			return "", ErrUserNotFoundByLogin
		}
		if errors.Is(err, repository.ErrInvalidPassword) {
			return "", ErrInvalidPassword
		}
		if errors.Is(err, repository.ErrDatabaseNotAvailable) {
			return "", ErrDatabaseUnavailable
		}
		return "", err
	}

	token, err := s.jwt.Generate(userID)
	if err != nil {
		return "", fmt.Errorf("error generating token: %v", err)
	}
	return token, nil
}

func IsValidLogin(email string) bool {
	return len(email) > 0 && len(email) < 32
}

func IsValidPassword(password string) bool {
	if len(password) < 8 || len(password) > 25 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, c := range password {
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
	return hasUpper && hasLower && hasDigit && hasSpecial
}
