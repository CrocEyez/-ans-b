package knowledge

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"ans-b/server/internal/qaimport"
)

type Repository struct {
	db    *sql.DB
	inner *qaimport.PostgresRepository
}

type ListFilters struct {
	Category  string
	Status    string
	Query     string
}

type ListItem struct {
	ID        int64     `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db, inner: qaimport.NewPostgresRepository(db)}
}

func (r *Repository) InsertKnowledge(ctx context.Context, record qaimport.KnowledgeRecord) error {
	if r == nil || r.inner == nil {
		return sql.ErrConnDone
	}
	return r.inner.InsertKnowledge(ctx, record)
}

func (r *Repository) List(ctx context.Context, limit, offset int, filters ListFilters) ([]ListItem, error) {
	if r == nil || r.db == nil {
		return nil, sql.ErrConnDone
	}
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, filters.Category)
		argIdx++
	}
	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(question ILIKE $%d OR answer ILIKE $%d)", argIdx, argIdx+1))
		pattern := "%" + filters.Query + "%"
		args = append(args, pattern, pattern)
		argIdx += 2
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT id, question, answer, category, tags, status, created_at, updated_at
		FROM knowledge_items
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ListItem
	for rows.Next() {
		var item ListItem
		var tagsStr string
		if err := rows.Scan(&item.ID, &item.Question, &item.Answer, &item.Category,
			&tagsStr, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.Tags = parseTagArray(tagsStr)
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) Count(ctx context.Context, filters ListFilters) (int, error) {
	if r == nil || r.db == nil {
		return 0, sql.ErrConnDone
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filters.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIdx))
		args = append(args, filters.Category)
		argIdx++
	}
	if filters.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filters.Status)
		argIdx++
	}
	if filters.Query != "" {
		conditions = append(conditions, fmt.Sprintf("(question ILIKE $%d OR answer ILIKE $%d)", argIdx, argIdx+1))
		pattern := "%" + filters.Query + "%"
		args = append(args, pattern, pattern)
		argIdx += 2
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	query := fmt.Sprintf(`SELECT COUNT(*) FROM knowledge_items %s`, where)
	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

func parseTagArray(s string) []string {
	if s == "" || s == "{}" {
		return []string{}
	}
	s = strings.Trim(s, "{}")
	if s == "" {
		return []string{}
	}
	tags := strings.Split(s, ",")
	result := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}
