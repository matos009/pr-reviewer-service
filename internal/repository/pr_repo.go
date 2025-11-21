package repository

import (
	"context"
	"errors"
	"pr-reviewer-service/internal/domain"

	"github.com/jackc/pgx/v5"
)

type PRRepository interface {
	Create(ctx context.Context, pr domain.PullRequest) error
	AddReviewer(ctx context.Context, prID, reviewerID string) error
	Get(ctx context.Context, prID string) (domain.PullRequest, error)
	Merge(ctx context.Context, prID string) error
	ReplaceReviewer(ctx context.Context, prID, oldUser, newUser string) error
	GetForReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error)
	Stats(ctx context.Context) (map[string]int, map[string]int, error)
	ReassignForDeactivated(ctx context.Context, inactive []string) error
}
type prRepo struct {
	db DB
}

func NewPRRepository(db DB) PRRepository {
	return &prRepo{db: db}
}

func (r *prRepo) Create(ctx context.Context, pr domain.PullRequest) error {
	var exists bool

	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id=$1)`,
		pr.ID,
	).Scan(&exists)

	if err != nil {
		return err
	}
	if exists {
		return domain.ErrPRExists
	}

	_, err = r.db.Exec(ctx,
		`INSERT INTO pull_requests 
         (pull_request_id, pull_request_name, author_id, status, created_at)
         VALUES ($1, $2, $3, $4, $5)`,
		pr.ID, pr.Name, pr.AuthorID, pr.Status, pr.CreatedAt,
	)
	return err
}
func (r *prRepo) AddReviewer(ctx context.Context, prID, userID string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO pull_request_reviewers (pull_request_id, user_id)
         VALUES ($1, $2)
         ON CONFLICT DO NOTHING`,
		prID, userID,
	)
	return err
}
func (r *prRepo) Get(ctx context.Context, prID string) (domain.PullRequest, error) {
	var pr domain.PullRequest

	err := r.db.QueryRow(ctx,
		`SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
           FROM pull_requests
          WHERE pull_request_id=$1`,
		prID,
	).Scan(&pr.ID, &pr.Name, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return domain.PullRequest{}, domain.ErrPRNotFound
	}
	if err != nil {
		return domain.PullRequest{}, err
	}

	// reviewers
	rows, err := r.db.Query(ctx,
		`SELECT user_id FROM pull_request_reviewers WHERE pull_request_id=$1`,
		prID,
	)
	if err != nil {
		return domain.PullRequest{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var uid string
		rows.Scan(&uid)
		pr.Reviewers = append(pr.Reviewers, uid)
	}

	return pr, nil
}

func (r *prRepo) Merge(ctx context.Context, prID string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE pull_requests
            SET status='MERGED', merged_at=NOW()
          WHERE pull_request_id=$1`,
		prID,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrPRNotFound
	}
	return nil
}
func (r *prRepo) ReplaceReviewer(ctx context.Context, prID, oldUser, newUser string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE pull_request_reviewers
            SET user_id=$3
          WHERE pull_request_id=$1 AND user_id=$2`,
		prID, oldUser, newUser,
	)

	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("reviewer not found")
	}
	return nil
}

func (r *prRepo) GetForReviewer(ctx context.Context, reviewerID string) ([]domain.PullRequestShort, error) {
	rows, err := r.db.Query(ctx,
		`SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
           FROM pull_requests pr
           JOIN pull_request_reviewers prr
             ON pr.pull_request_id = prr.pull_request_id
          WHERE prr.user_id=$1`,
		reviewerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.PullRequestShort

	for rows.Next() {
		var p domain.PullRequestShort
		rows.Scan(&p.ID, &p.Name, &p.AuthorID, &p.Status)
		result = append(result, p)
	}

	return result, nil
}

func (r *prRepo) Stats(ctx context.Context) (map[string]int, map[string]int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT user_id, COUNT(*) 
		FROM pull_request_reviewers 
		GROUP BY user_id
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	reviewerCount := map[string]int{}
	for rows.Next() {
		var id string
		var cnt int
		if err := rows.Scan(&id, &cnt); err != nil {
			return nil, nil, err
		}
		reviewerCount[id] = cnt
	}

	rows2, err := r.db.Query(ctx, `
		SELECT status, COUNT(*)
		FROM pull_requests
		GROUP BY status
	`)
	if err != nil {
		return nil, nil, err
	}
	defer rows2.Close()

	statusCount := map[string]int{}
	for rows2.Next() {
		var status string
		var cnt int
		if err := rows2.Scan(&status, &cnt); err != nil {
			return nil, nil, err
		}
		statusCount[status] = cnt
	}

	return reviewerCount, statusCount, nil
}
func (r *prRepo) ReassignForDeactivated(ctx context.Context, inactive []string) error {
	rows, err := r.db.Query(ctx, `
		SELECT prr.pull_request_id
		FROM pull_request_reviewers prr
		JOIN pull_requests pr ON pr.pull_request_id = prr.pull_request_id
		WHERE pr.status = 'OPEN'
		  AND prr.user_id = ANY($1)
	`, inactive)
	if err != nil {
		return err
	}
	defer rows.Close()

	var prIDs []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		prIDs = append(prIDs, id)
	}

	if len(prIDs) == 0 {
		return nil
	}

	_, err = r.db.Exec(ctx, `
		DELETE FROM pull_request_reviewers
		WHERE user_id = ANY($1)
	`, inactive)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, `
		INSERT INTO pull_request_reviewers (pull_request_id, user_id)
		SELECT pr.pull_request_id, u.user_id
		FROM pull_requests pr
		JOIN users u ON u.team_name = (SELECT team_name FROM users WHERE user_id = pr.author_id)
		WHERE pr.pull_request_id = ANY($1)
		  AND u.is_active = TRUE
		  AND u.user_id != pr.author_id
	`, prIDs)

	return err
}
