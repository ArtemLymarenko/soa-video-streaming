package postgres

import (
	"context"
	"errors"
	"soa-video-streaming/pkg/postgres"
)

type UserPreference struct {
	db postgres.DB
}

func NewUserPreference(db postgres.DB) *UserPreference {
	return &UserPreference{
		db: db,
	}
}

func (r *UserPreference) AddPreferredCategories(ctx context.Context, userID string, categoryIDs []string) error {
	if len(categoryIDs) == 0 {
		return errors.New("category list is empty")
	}

	q := `INSERT INTO user_preferred_categories (user_id, category_id)
        SELECT $1, unnest($2::uuid[])
        ON CONFLICT DO NOTHING
	`

	_, err := r.db.Exec(ctx, q, userID, categoryIDs)
	return err
}
