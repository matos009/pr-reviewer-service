package repository

import (
	"context"
	"errors"
	"pr-reviewer-service/internal/domain"

	"github.com/jackc/pgx/v5"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) error
	SetActive(ctx context.Context, userID string, active bool) error
	Get(ctx context.Context, userID string) (*domain.User, error)
	GetActiveUsersByTeam(ctx context.Context, team string) ([]domain.User, error)
	DeactivateMany(ctx context.Context, ids []string) error
}
type userRepo struct {
	db DB
}

func NewUserRepository(db DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user domain.User) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (user_id, username, team_name, is_active)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (user_id)
		 DO UPDATE SET username=EXCLUDED.username,
		               team_name=EXCLUDED.team_name,
	                   is_active=EXCLUDED.is_active`,
		user.ID, user.Username, user.TeamName, user.IsActive,
	)
	return err
}

func (r *userRepo) SetActive(ctx context.Context, userID string, active bool) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE users SET is_active=$2 WHERE user_id=$1`,
		userID, active,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *userRepo) Get(ctx context.Context, userID string) (*domain.User, error) {
	var u domain.User
	err := r.db.QueryRow(ctx,
		`SELECT user_id, username, team_name, is_active
		   FROM users WHERE user_id=$1`,
		userID,
	).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	return &u, err
}

func (r *userRepo) GetActiveUsersByTeam(ctx context.Context, team string) ([]domain.User, error) {
	rows, err := r.db.Query(ctx,
		`SELECT user_id, username, team_name, is_active
		   FROM users
		  WHERE team_name=$1 AND is_active=true`,
		team,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.User
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, nil
}
func (r *userRepo) DeactivateMany(ctx context.Context, ids []string) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users SET is_active = false
		WHERE user_id = ANY($1)
	`, ids)
	return err
}
