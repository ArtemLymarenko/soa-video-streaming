package postgres

import (
	"context"
	"errors"
	"soa-video-streaming/services/content-service/internal/domain/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategories(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, c entity.Category) error {
	q := `INSERT INTO categories (id, name, description) VALUES ($1, $2, $3)`

	_, err := r.db.Exec(ctx, q, c.ID, c.Name, c.Description)
	return err
}

func (r *CategoryRepository) GetByID(ctx context.Context, id entity.CategoryID) (*entity.Category, error) {
	q := `SELECT id, name, description FROM categories WHERE id = $1`

	row := r.db.QueryRow(ctx, q, id)

	var c entity.Category
	if err := row.Scan(&c.ID, &c.Name, &c.Description); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &c, nil
}

func (r *CategoryRepository) Update(ctx context.Context, c entity.Category) error {
	q := `UPDATE categories SET name = $1, description = $2 WHERE id = $3`

	_, err := r.db.Exec(ctx, q, c.Name, c.Description, c.ID)
	return err
}

func (r *CategoryRepository) Delete(ctx context.Context, id entity.CategoryID) error {
	q := `DELETE FROM categories WHERE id = $1`

	_, err := r.db.Exec(ctx, q, id)
	return err
}

func (r *CategoryRepository) GetByTimestamp(ctx context.Context, from, to int64) ([]entity.Category, error) {
	q := `SELECT id, name, description FROM categories WHERE updated_at >= to_timestamp($1) AND updated_at <= to_timestamp($2)`

	rows, err := r.db.Query(ctx, q, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []entity.Category

	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description); err != nil {
			return nil, err
		}
		items = append(items, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *CategoryRepository) GetMaxTimestamp(ctx context.Context) (int64, error) {
	q := `SELECT COALESCE(
            EXTRACT(EPOCH FROM MAX(updated_at))::bigint, 
            0
        ) FROM categories`

	row := r.db.QueryRow(ctx, q)

	var maxTimestamp int64
	if err := row.Scan(&maxTimestamp); err != nil {
		return 0, err
	}

	return maxTimestamp, nil
}
