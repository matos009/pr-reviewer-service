package domain

import "time"

// ---------------- PR STATUS -----------------

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

// ---------------- FULL PR -----------------

type PullRequest struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	AuthorID  string     `json:"author_id"`
	Status    PRStatus   `json:"status"`
	Reviewers []string   `json:"reviewers"`
	CreatedAt time.Time  `json:"created_at"`
	MergedAt  *time.Time `json:"merged_at"`
}

// ---------------- SHORT PR (для списка ревьюверов) -----------------

type PullRequestShort struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	AuthorID string   `json:"author_id"`
	Status   PRStatus `json:"status"`
}
