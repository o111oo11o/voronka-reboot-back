package service

import (
	"context"

	"voronka/internal/domain"
	"voronka/internal/repository"
)

type chatService struct {
	messages repository.MessageRepository
}

func NewChatService(messages repository.MessageRepository) ChatService {
	return &chatService{messages: messages}
}

func (s *chatService) SendMessage(ctx context.Context, req domain.SendMessageRequest) (*domain.Message, error) {
	return s.messages.Create(ctx, req)
}

func (s *chatService) GetHistory(ctx context.Context, filter domain.MessageFilter) ([]*domain.Message, error) {
	return s.messages.List(ctx, filter)
}
