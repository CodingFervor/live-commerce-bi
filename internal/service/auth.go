package service

import (
	"context"
	"errors"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/model"
	"github.com/CodingFervor/live-commerce-bi/internal/repository"
	"github.com/CodingFervor/live-commerce-bi/pkg/hash"
	"github.com/CodingFervor/live-commerce-bi/pkg/jwt"
)

type AuthService struct {
	repo *repository.AuthRepo
}

func NewAuthService() *AuthService {
	return &AuthService{repo: repository.NewAuthRepo()}
}

func (s *AuthService) Register(ctx context.Context, req *model.RegisterRequest) (*model.User, error) {
	// Check duplicate
	existing, _ := s.repo.GetUserByUsername(ctx, req.Username)
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	// Enforce password policy
	if err := validatePassword(req.Password); err != nil {
		return nil, err
	}

	hashedPwd, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	// Always assign "analyst" role — privilege escalation prevention
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPwd,
		Role:     "analyst",
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// validatePassword enforces security password policies
func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	hasUpper := false
	hasDigit := false
	for _, c := range password {
		if c >= 'A' && c <= 'Z' {
			hasUpper = true
		}
		if c >= '0' && c <= '9' {
			hasDigit = true
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	return nil
}

func (s *AuthService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if !hash.CheckPassword(req.Password, user.Password) {
		return nil, errors.New("invalid credentials")
	}
	token, err := jwt.GenerateToken(user.ID, user.Username, user.Role, 24)
	if err != nil {
		return nil, err
	}
	_ = s.repo.UpdateLastLogin(ctx, user.ID)
	return &model.LoginResponse{
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour).Format(time.RFC3339),
		User:      *user,
	}, nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID int64) (*model.User, error) {
	return s.repo.GetUserByID(ctx, userID)
}
