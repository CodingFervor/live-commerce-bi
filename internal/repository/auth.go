package repository

import (
	"context"
	"time"

	"github.com/CodingFervor/live-commerce-bi/internal/database"
	"github.com/CodingFervor/live-commerce-bi/internal/model"
)

type AuthRepo struct{}

func NewAuthRepo() *AuthRepo { return &AuthRepo{} }

func (r *AuthRepo) CreateUser(ctx context.Context, user *model.User) error {
	query := `INSERT INTO users (username, email, password, role, status)
		VALUES ($1, $2, $3, $4, 'active') RETURNING id, created_at`
	return database.Get().QueryRow(ctx, query,
		user.Username, user.Email, user.Password, user.Role,
	).Scan(&user.ID, &user.CreatedAt)
}

func (r *AuthRepo) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	query := `SELECT id, username, email, password, role, avatar, status, last_login_at, created_at, updated_at
		FROM users WHERE username = $1 AND status = 'active'`
	err := database.Get().QueryRow(ctx, query, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.Avatar, &u.Status, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepo) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	var u model.User
	query := `SELECT id, username, email, password, role, avatar, status, last_login_at, created_at, updated_at
		FROM users WHERE id = $1`
	err := database.Get().QueryRow(ctx, query, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.Password, &u.Role, &u.Avatar, &u.Status, &u.LastLoginAt, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *AuthRepo) UpdateLastLogin(ctx context.Context, id int64) error {
	_, err := database.Get().Exec(ctx,
		"UPDATE users SET last_login_at = $1 WHERE id = $2", time.Now(), id)
	return err
}
