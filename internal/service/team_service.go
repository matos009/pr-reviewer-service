package service

import (
	"context"
	"errors"
	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/repository"
)

type TeamService struct {
	repo repository.TeamRepository
}

func NewTeamService(repo repository.TeamRepository) *TeamService {
	return &TeamService{repo: repo}
}

func (s *TeamService) Create(ctx context.Context, name string) error {
	return s.repo.Create(ctx, name)
}

func (s *TeamService) Get(ctx context.Context, name string) (*domain.Team, error) {
	team, err := s.repo.Get(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrTeamNotFound) {
			return nil, domain.ErrTeamNotFound
		}
		return nil, err
	}
	return team, nil
}
func (s *TeamService) CreateWithMembers(ctx context.Context, team *domain.Team) error {
	if err := s.repo.Create(ctx, team.Name); err != nil {
		if errors.Is(err, repository.ErrTeamExists) {
			return domain.ErrTeamExists
		}
		return err
	}

	return s.repo.AddMembers(ctx, team.Name, team.Members)
}
