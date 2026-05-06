package waypoint

import (
	"context"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, tripID string) ([]Waypoint, error) {
	wps, err := s.repo.ListByTrip(ctx, tripID)
	if err != nil {
		return nil, err
	}
	if wps == nil {
		wps = []Waypoint{}
	}
	return wps, nil
}

func (s *Service) Create(ctx context.Context, tripID string, req CreateRequest) (*Waypoint, error) {
	if req.Lat == 0 && req.Lng == 0 {
		return nil, errors.New("lat and lng required")
	}
	return s.repo.Create(ctx, tripID, req)
}

func (s *Service) Update(ctx context.Context, id, tripID string, req UpdateRequest) (*Waypoint, error) {
	return s.repo.Update(ctx, id, tripID, req)
}

func (s *Service) Delete(ctx context.Context, id, tripID string) error {
	return s.repo.Delete(ctx, id, tripID)
}

func (s *Service) Reorder(ctx context.Context, tripID string, items []ReorderItem) error {
	if len(items) == 0 {
		return errors.New("items required")
	}
	return s.repo.Reorder(ctx, tripID, items)
}
