package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"pr-reviewer-service/internal/domain"
	"pr-reviewer-service/internal/repository"
)

type PRService struct {
	prRepo   repository.PRRepository
	userRepo repository.UserRepository
	rnd      *rand.Rand
}

func NewPRService(prRepo repository.PRRepository, userRepo repository.UserRepository) *PRService {
	return &PRService{
		prRepo:   prRepo,
		userRepo: userRepo,
		rnd:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// ----------------- CREATE PR (+ автоназначение ревьюверов) -----------------
func (s *PRService) Create(ctx context.Context, id, name, authorID string) (domain.PullRequest, error) {
	author, err := s.userRepo.Get(ctx, authorID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.PullRequest{}, domain.ErrUserNotFound
		}
		return domain.PullRequest{}, err
	}
	if !author.IsActive {
		return domain.PullRequest{}, domain.ErrUserNotActive
	}

	pr := domain.PullRequest{
		ID:        id,
		Name:      name,
		AuthorID:  authorID,
		Status:    domain.PRStatusOpen,
		CreatedAt: time.Now().UTC(),
	}

	if err := s.prRepo.Create(ctx, pr); err != nil {
		if errors.Is(err, domain.ErrPRExists) {
			return domain.PullRequest{}, domain.ErrPRExists
		}
		return domain.PullRequest{}, err
	}

	reviewers, err := s.pickReviewers(ctx, author.TeamName, authorID, 2)
	if err != nil && !errors.Is(err, domain.ErrNoCandidate) {

		return domain.PullRequest{}, err
	}

	for _, rID := range reviewers {
		if err := s.prRepo.AddReviewer(ctx, pr.ID, rID); err != nil {
			return domain.PullRequest{}, err
		}
	}

	pr.Reviewers = reviewers
	return pr, nil
}

func (s *PRService) pickReviewers(ctx context.Context, teamName, excludeUserID string, limit int) ([]string, error) {
	users, err := s.userRepo.GetActiveUsersByTeam(ctx, teamName)
	if err != nil {
		return nil, err
	}

	var candidates []string
	for _, u := range users {
		if u.ID == excludeUserID {
			continue
		}
		candidates = append(candidates, u.ID)
	}

	if len(candidates) == 0 {
		return nil, domain.ErrNoCandidate
	}
	if len(candidates) <= limit {
		return candidates, nil
	}

	res := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		j := i + s.rnd.Intn(len(candidates)-i)
		candidates[i], candidates[j] = candidates[j], candidates[i]
		res = append(res, candidates[i])
	}
	return res, nil
}

// ----------------- MERGE (идемпотентный) -----------------

func (s *PRService) Merge(ctx context.Context, prID string) (domain.PullRequest, error) {
	pr, err := s.prRepo.Get(ctx, prID)
	if err != nil {
		if errors.Is(err, domain.ErrPRNotFound) {
			return domain.PullRequest{}, domain.ErrPRNotFound
		}
		return domain.PullRequest{}, err
	}

	if pr.Status == domain.PRStatusMerged {
		return pr, nil
	}

	if err := s.prRepo.Merge(ctx, prID); err != nil {
		if errors.Is(err, domain.ErrPRNotFound) {
			return domain.PullRequest{}, domain.ErrPRNotFound
		}
		return domain.PullRequest{}, err
	}

	return s.prRepo.Get(ctx, prID)
}

// ----------------- REASSIGN REVIEWER -----------------
func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (domain.PullRequest, string, error) {
	pr, err := s.prRepo.Get(ctx, prID)
	if err != nil {
		if errors.Is(err, domain.ErrPRNotFound) {
			return domain.PullRequest{}, "", domain.ErrPRNotFound
		}
		return domain.PullRequest{}, "", err
	}

	if pr.Status == domain.PRStatusMerged {
		return domain.PullRequest{}, "", domain.ErrPRMerged
	}

	assigned := false
	var otherReviewers []string
	for _, rID := range pr.Reviewers {
		if rID == oldReviewerID {
			assigned = true
		} else {
			otherReviewers = append(otherReviewers, rID)
		}
	}
	if !assigned {
		return domain.PullRequest{}, "", domain.ErrNotAssigned
	}

	oldUser, err := s.userRepo.Get(ctx, oldReviewerID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.PullRequest{}, "", domain.ErrUserNotFound
		}
		return domain.PullRequest{}, "", err
	}

	users, err := s.userRepo.GetActiveUsersByTeam(ctx, oldUser.TeamName)
	if err != nil {
		return domain.PullRequest{}, "", err
	}

	var candidates []string
	for _, u := range users {
		if u.ID == oldReviewerID {
			continue
		}
		if u.ID == pr.AuthorID {
			continue
		}
		dup := false
		for _, existing := range otherReviewers {
			if existing == u.ID {
				dup = true
				break
			}
		}
		if dup {
			continue
		}
		candidates = append(candidates, u.ID)
	}

	if len(candidates) == 0 {
		return domain.PullRequest{}, "", domain.ErrNoCandidate
	}

	newID := candidates[s.rnd.Intn(len(candidates))]

	if err := s.prRepo.ReplaceReviewer(ctx, prID, oldReviewerID, newID); err != nil {
		if err.Error() == "reviewer not found" {
			return domain.PullRequest{}, "", domain.ErrNotAssigned
		}
		return domain.PullRequest{}, "", err
	}

	updated, err := s.prRepo.Get(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, "", err
	}

	return updated, newID, nil
}

// ----------------- GET PRs WHERE USER IS REVIEWER -----------------

func (s *PRService) GetUserReviews(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	return s.prRepo.GetForReviewer(ctx, userID)
}

func (s *PRService) Stats(ctx context.Context) (map[string]int, map[string]int, error) {
	return s.prRepo.Stats(ctx)
}
