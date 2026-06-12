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
	hashedPwd, err := hash.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	role := req.Role
	if role == "" {
		role = "analyst"
	}
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPwd,
		Role:     role,
	}
	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
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
