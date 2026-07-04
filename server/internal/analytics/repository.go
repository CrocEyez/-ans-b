package analytics

import (
	"context"
	"database/sql"
	"errors"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) IncrementKnowledgeAccess(ctx context.Context, itemID int64) error {
	if r == nil || r.db == nil {
		return errors.New("analytics repository is not configured")
	}
	if itemID <= 0 {
		return errors.New("knowledge item id is required")
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE knowledge_items
		SET access_count = access_count + 1,
		    last_accessed_at = now()
		WHERE id = $1
	`, itemID)
	return err
}

type HotQuestion struct {
	ID          int64  `json:"id"`
	Question    string `json:"question"`
	AccessCount int64  `json:"access_count"`
}

func (r *Repository) TopQuestions(ctx context.Context, limit int) ([]HotQuestion, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("analytics repository is not configured")
	}
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, question, access_count
		FROM knowledge_items
		WHERE access_count > 0
		ORDER BY access_count DESC, last_accessed_at DESC NULLS LAST
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var questions []HotQuestion
	for rows.Next() {
		var q HotQuestion
		if err := rows.Scan(&q.ID, &q.Question, &q.AccessCount); err != nil {
			return nil, err
		}
		questions = append(questions, q)
	}
	return questions, rows.Err()
}
