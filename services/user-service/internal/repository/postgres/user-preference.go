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

	q := `INSERT INTO user_service.user_preferred_categories (user_id, category_id)
        SELECT $1, unnest($2::uuid[])
        ON CONFLICT DO NOTHING
	`

	_, err := r.db.Exec(ctx, q, userID, categoryIDs)
	return err
}

func (r *UserPreference) GetUserPreferredCategories(ctx context.Context, userID string) ([]string, error) {
	q := `SELECT category_id FROM user_service.user_preferred_categories WHERE user_id = $1`

	rows, err := r.db.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categoryIDs []string
	for rows.Next() {
		var categoryID string
		if err := rows.Scan(&categoryID); err != nil {
			return nil, err
		}
		categoryIDs = append(categoryIDs, categoryID)
	}

	return categoryIDs, nil
}
