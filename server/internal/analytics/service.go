package analytics

import (
	"context"
	"errors"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) IncrementKnowledgeAccess(ctx context.Context, itemID int64) error {
	if s == nil || s.repository == nil {
		return errors.New("analytics service is not configured")
	}
	return s.repository.IncrementKnowledgeAccess(ctx, itemID)
}

func (s *Service) GetHotQuestions(ctx context.Context, limit int) ([]HotQuestion, error) {
	if s == nil || s.repository == nil {
		return nil, errors.New("analytics service is not configured")
	}
	if limit <= 0 {
		limit = 10
	}
	return s.repository.TopQuestions(ctx, limit)
}
