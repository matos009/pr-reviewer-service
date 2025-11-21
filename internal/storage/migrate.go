package storage

import (
	"context"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ApplyMigrations(ctx context.Context, db *pgxpool.Pool) error {
	sqlBytes, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}

	queries := strings.Split(string(sqlBytes), ";")

	for _, q := range queries {
		q = strings.TrimSpace(q)
		if q == "" {
			continue
		}
		_, execErr := db.Exec(ctx, q)
		if execErr != nil {
			return execErr
		}
	}

	return nil
}
