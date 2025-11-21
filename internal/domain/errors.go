package domain

import "errors"

var (
	ErrTeamNotFound  = errors.New("team not found")
	ErrUserNotFound  = errors.New("user not found")
	ErrPRNotFound    = errors.New("pull request not found")
	ErrUserNotActive = errors.New("user is not active")
	ErrAlreadyMerged = errors.New("pull request already merged")
	ErrPRExists      = errors.New("pr already exists")
	ErrPRMerged      = errors.New("cannot update merged PR")
	ErrNoCandidate   = errors.New("no candidate for reviewer")
	ErrNotAssigned   = errors.New("reviewer not assigned to this PR")
	ErrTeamExists    = errors.New("team already exists")
)
