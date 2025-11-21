package repository

import (
	"context"
	"errors"
	"pr-reviewer-service/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row

	Begin(ctx context.Context) (pgx.Tx, error)
}

type TeamRepository interface {
	Create(ctx context.Context, name string) error
	Get(ctx context.Context, name string) (*domain.Team, error)

	AddMembers(ctx context.Context, teamName string, members []domain.TeamMember) error
}

type teamRepo struct {
	db DB
}

func NewTeamRepository(db DB) TeamRepository {
	return &teamRepo{db: db}
}

var ErrTeamExists = errors.New("team already exists")
var ErrTeamNotFound = errors.New("team not found")

func (r *teamRepo) Create(ctx context.Context, name string) error {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM teams WHERE name=$1)`,
		name,
	).Scan(&exists)

	if err != nil {
		return err
	}
	if exists {
		return ErrTeamExists
	}

	_, err = r.db.Exec(ctx,
		`INSERT INTO teams (name) VALUES ($1)`,
		name,
	)
	return err
}

func (r *teamRepo) Get(ctx context.Context, teamName string) (*domain.Team, error) {

	rows, err := r.db.Query(ctx,
		`SELECT user_id, username, is_active FROM users WHERE team_name=$1`,
		teamName,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var m domain.TeamMember
		if err := rows.Scan(&m.ID, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	if len(members) == 0 {
		// Проверяем, существует ли команда
		var exists bool
		err = r.db.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM teams WHERE name=$1)`,
			teamName,
		).Scan(&exists)

		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, ErrTeamNotFound
		}
	}

	return &domain.Team{
		Name:    teamName,
		Members: members,
	}, nil
}
func (r *teamRepo) AddMembers(ctx context.Context, teamName string, members []domain.TeamMember) error {
	for _, m := range members {
		_, err := r.db.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
		`, m.ID, m.Username, teamName, m.IsActive)

		if err != nil {
			return err
		}
	}
	return nil
}
