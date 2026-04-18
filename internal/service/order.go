package service

import (
	"context"

	"github.com/google/uuid"

	"voronka/internal/domain"
	"voronka/internal/repository"
)

type orderService struct {
	orders repository.OrderRepository
}

func NewOrderService(orders repository.OrderRepository) OrderService {
	return &orderService{orders: orders}
}

func (s *orderService) PlaceOrder(ctx context.Context, req domain.PlaceOrderRequest) (*domain.Order, error) {
	if len(req.Items) == 0 {
		return nil, domain.ErrBadRequest
	}
	return s.orders.Create(ctx, req)
}

func (s *orderService) GetOrder(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.orders.GetByID(ctx, id)
}

func (s *orderService) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) (*domain.Order, error) {
	return s.orders.UpdateStatus(ctx, id, status)
}

func (s *orderService) ListOrders(ctx context.Context, filter domain.OrderFilter) ([]*domain.Order, error) {
	return s.orders.List(ctx, filter)
}
