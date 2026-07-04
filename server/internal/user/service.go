package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"ans-b/server/internal/auth"
)

var ErrUsernameTaken = errors.New("username already exists")

type RegisterInput struct {
	Username string
	Password string
	Nickname string
}

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (*User, error) {
	username := strings.TrimSpace(input.Username)
	nickname := strings.TrimSpace(input.Nickname)
	if username == "" {
		return nil, errors.New("username is required")
	}
	if len(username) > 64 {
		return nil, errors.New("username is too long")
	}
	if nickname != "" && len(nickname) > 100 {
		return nil, errors.New("nickname is too long")
	}
	passwordHash, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}
	if s == nil || s.repository == nil {
		return nil, errors.New("user service is not configured")
	}
	return s.repository.Create(ctx, CreateInput{
		Username:     username,
		PasswordHash: passwordHash,
		Nickname:     nickname,
	})
}

func (s *Service) Profile(ctx context.Context, id int64) (*User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user id")
	}
	if s == nil || s.repository == nil {
		return nil, errors.New("user service is not configured")
	}
	user, err := s.repository.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("user not found")
	}
	return user, err
}

type ListInput struct {
	Page     int
	PageSize int
}

type ListResult struct {
	Items    []User `json:"items"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("user service is not configured")
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 10
	}

	total, err := s.repository.Count(ctx)
	if err != nil {
		return nil, err
	}

	offset := (input.Page - 1) * input.PageSize
	users, err := s.repository.List(ctx, input.PageSize, offset)
	if err != nil {
		return nil, err
	}
	if users == nil {
		users = []User{}
	}

	return &ListResult{
		Items:    users,
		Total:    total,
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}
