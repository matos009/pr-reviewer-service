package service

import (
	"context"
	"pr-reviewer-service/internal/repository"
)

type TeamAdminService struct {
	users repository.UserRepository
	prs   repository.PRRepository
}

func NewTeamAdminService(
	users repository.UserRepository,
	prs repository.PRRepository,
) *TeamAdminService {
	return &TeamAdminService{users: users, prs: prs}
}

func (s *TeamAdminService) DeactivateTeam(ctx context.Context, team string) error {
	users, err := s.users.GetActiveUsersByTeam(ctx, team)
	if err != nil {
		return err
	}
	if len(users) == 0 {
		return nil
	}

	var ids []string
	for _, u := range users {
		ids = append(ids, u.ID)
	}

	if err := s.users.DeactivateMany(ctx, ids); err != nil {
		return err
	}

	return s.prs.ReassignForDeactivated(ctx, ids)
}
