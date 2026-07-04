package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Nickname  string    `json:"nickname"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateInput struct {
	Username     string
	PasswordHash string
	Nickname     string
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, input CreateInput) (*User, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("user repository is not configured")
	}

	var created User
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO users (username, password_hash, nickname)
		VALUES ($1, $2, NULLIF($3, ''))
		RETURNING id, username, COALESCE(nickname, ''), created_at, updated_at
	`, input.Username, input.PasswordHash, input.Nickname).Scan(
		&created.ID, &created.Username, &created.Nickname,
		&created.CreatedAt, &created.UpdatedAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, ErrUsernameTaken
		}
		return nil, err
	}
	return &created, nil
}

func (r *Repository) FindByID(ctx context.Context, id int64) (*User, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("user repository is not configured")
	}

	var found User
	err := r.db.QueryRowContext(ctx, `
		SELECT id, username, COALESCE(nickname, ''), created_at, updated_at
		FROM users
		WHERE id = $1
	`, id).Scan(&found.ID, &found.Username, &found.Nickname, &found.CreatedAt, &found.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &found, nil
}

func (r *Repository) List(ctx context.Context, limit, offset int) ([]User, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("user repository is not configured")
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, COALESCE(nickname, ''), created_at, updated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Username, &u.Nickname, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *Repository) Count(ctx context.Context) (int, error) {
	if r == nil || r.db == nil {
		return 0, errors.New("user repository is not configured")
	}
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func isUniqueViolation(err error) bool {
	message := err.Error()
	return strings.Contains(message, "duplicate key value") ||
		strings.Contains(message, "unique constraint") ||
		strings.Contains(message, "idx_users_username")
}
