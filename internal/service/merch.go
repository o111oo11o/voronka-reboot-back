package service

import (
	"context"

	"github.com/google/uuid"

	"voronka/internal/domain"
	"voronka/internal/repository"
)

type merchService struct {
	items repository.MerchItemRepository
}

func NewMerchService(items repository.MerchItemRepository) MerchService {
	return &merchService{items: items}
}

func (s *merchService) CreateItem(ctx context.Context, req domain.CreateMerchItemRequest) (*domain.MerchItem, error) {
	return s.items.Create(ctx, req)
}

func (s *merchService) GetItem(ctx context.Context, id uuid.UUID) (*domain.MerchItem, error) {
	return s.items.GetByID(ctx, id)
}

func (s *merchService) UpdateItem(ctx context.Context, id uuid.UUID, req domain.UpdateMerchItemRequest) (*domain.MerchItem, error) {
	if _, err := s.items.GetByID(ctx, id); err != nil {
		return nil, err
	}
	return s.items.Update(ctx, id, req)
}

func (s *merchService) DeleteItem(ctx context.Context, id uuid.UUID) error {
	return s.items.Delete(ctx, id)
}

func (s *merchService) ListItems(ctx context.Context, filter domain.MerchItemFilter) ([]*domain.MerchItem, error) {
	return s.items.List(ctx, filter)
}
