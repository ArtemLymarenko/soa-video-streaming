package postgres

import (
	"context"
	"errors"
	"soa-video-streaming/services/content-service/internal/domain/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MediaContentRepository struct {
	db *pgxpool.Pool
}

func NewMediaContent(db *pgxpool.Pool) *MediaContentRepository {
	return &MediaContentRepository{db: db}
}

func (r *MediaContentRepository) Create(ctx context.Context, m entity.MediaContent) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const insertMedia = `
		INSERT INTO media_content (
			id, name, description, type, duration, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = tx.Exec(ctx, insertMedia,
		m.ID,
		m.Name,
		m.Description,
		m.Type,
		m.Duration,
		m.CreatedAt,
		m.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if len(m.Categories) == 0 {
		return tx.Commit(ctx)
	}

	catIDs := make([]string, len(m.Categories))
	for i, c := range m.Categories {
		catIDs[i] = string(c)
	}

	const insertCategories = `
		INSERT INTO media_content_categories (media_content_id, category_id)
		SELECT $1, unnest($2::text[])
	`

	_, err = tx.Exec(ctx, insertCategories, m.ID, catIDs)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *MediaContentRepository) GetByID(ctx context.Context, id string) (*entity.MediaContent, error) {
	q := `
		SELECT id, name, description, type, duration, created_at, updated_at
		FROM media_content
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, q, id)

	var m entity.MediaContent
	if err := row.Scan(
		&m.ID,
		&m.Name,
		&m.Description,
		&m.Type,
		&m.Duration,
		&m.CreatedAt,
		&m.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	catRows, err := r.db.Query(ctx,
		`SELECT category_id FROM media_content_categories WHERE media_content_id = $1`,
		id,
	)
	if err != nil {
		return nil, err
	}
	defer catRows.Close()

	for catRows.Next() {
		var cid entity.CategoryID
		if err := catRows.Scan(&cid); err != nil {
			return nil, err
		}
		m.Categories = append(m.Categories, cid)
	}

	if err := catRows.Err(); err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *MediaContentRepository) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM media_content WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}
