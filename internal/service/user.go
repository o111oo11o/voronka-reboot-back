package service

import (
	"context"

	"github.com/google/uuid"

	"voronka/internal/domain"
	"voronka/internal/repository"
)

type userService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) UserService {
	return &userService{users: users}
}

func (s *userService) Create(ctx context.Context, req domain.CreateUserRequest) (*domain.User, error) {
	return s.users.Create(ctx, req)
}

func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, req domain.UpdateUserRequest) (*domain.User, error) {
	if _, err := s.users.GetByID(ctx, id); err != nil {
		return nil, err
	}
	return s.users.Update(ctx, id, req)
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.users.Delete(ctx, id)
}

func (s *userService) List(ctx context.Context, filter domain.UserFilter) ([]*domain.User, error) {
	return s.users.List(ctx, filter)
}
