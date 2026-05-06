package trip

import (
	"context"
	"errors"
)

var ErrTripLimitReached = errors.New("free tier trip limit reached")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, userID string, page, limit int) ([]Trip, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.repo.ListByUser(ctx, userID, page, limit)
}

func (s *Service) Get(ctx context.Context, id, userID string) (*Trip, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *Service) Create(ctx context.Context, userID, tier string, req CreateRequest) (*Trip, error) {
	if req.Title == "" {
		return nil, errors.New("title required")
	}

	// Free tier limit
	if tier == "free" {
		count, err := s.repo.CountByUser(ctx, userID)
		if err != nil {
			return nil, err
		}
		if count >= FreeTierTripLimit {
			return nil, ErrTripLimitReached
		}
	}

	return s.repo.Create(ctx, userID, req)
}

func (s *Service) Update(ctx context.Context, id, userID string, req UpdateRequest) (*Trip, error) {
	return s.repo.Update(ctx, id, userID, req)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}
