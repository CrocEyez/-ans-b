package knowledge

import (
	"context"
	"errors"
	"strings"

	"ans-b/server/internal/qaimport"
)

type Service struct {
	repository *Repository
	embedder   qaimport.Embedder
}

type CreateInput struct {
	Question string
	Answer   string
	Category string
	Tags     []string
	Source   string
	Remark   string
}

type ListInput struct {
	Page     int
	PageSize int
	Category string
	Status   string
	Query    string
}

type ListResult struct {
	Items    []ListItem `json:"items"`
	Total    int        `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}

func NewService(repository *Repository, embedder qaimport.Embedder) *Service {
	return &Service{repository: repository, embedder: embedder}
}

func (s *Service) Create(ctx context.Context, input CreateInput) error {
	if s.repository == nil {
		return errors.New("knowledge repository is not configured")
	}
	if s.embedder == nil {
		return errors.New("embedder is not configured")
	}

	item := qaimport.Item{
		Question: strings.TrimSpace(input.Question),
		Answer:   strings.TrimSpace(input.Answer),
		Category: strings.TrimSpace(input.Category),
		Tags:     input.Tags,
		Source:   strings.TrimSpace(input.Source),
		Remark:   strings.TrimSpace(input.Remark),
	}
	_, err := qaimport.ImportItems(ctx, s.repository.inner, s.embedder, []qaimport.Item{item})
	return err
}

func (s *Service) List(ctx context.Context, input ListInput) (*ListResult, error) {
	if s.repository == nil {
		return nil, errors.New("knowledge repository is not configured")
	}
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 || input.PageSize > 100 {
		input.PageSize = 10
	}

	filters := ListFilters{
		Category: input.Category,
		Status:   input.Status,
		Query:    input.Query,
	}

	total, err := s.repository.Count(ctx, filters)
	if err != nil {
		return nil, err
	}

	offset := (input.Page - 1) * input.PageSize
	items, err := s.repository.List(ctx, input.PageSize, offset, filters)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []ListItem{}
	}

	return &ListResult{
		Items:    items,
		Total:    total,
		Page:     input.Page,
		PageSize: input.PageSize,
	}, nil
}
