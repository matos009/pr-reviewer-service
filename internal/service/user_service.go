package service

import (
	"context"
	"errors"

	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/repository"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(ctx context.Context, user domain.User) error {
	return s.repo.Create(ctx, user)
}

func (s *UserService) SetActive(ctx context.Context, id string, isActive bool) error {
	return s.repo.SetActive(ctx, id, isActive)
}

func (s *UserService) Get(ctx context.Context, id string) (*domain.User, error) {
	u, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}
