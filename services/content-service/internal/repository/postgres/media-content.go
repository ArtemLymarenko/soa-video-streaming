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

	if len(m.Categories) > 0 {
		catIDs := make([]string, len(m.Categories))
		for i, c := range m.Categories {
			catIDs[i] = string(c.ID)
		}

		const insertCategories = `
			INSERT INTO media_content_categories (media_content_id, category_id)
			SELECT $1, unnest($2::text[])
		`

		_, err = tx.Exec(ctx, insertCategories, m.ID, catIDs)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *MediaContentRepository) GetByID(ctx context.Context, id string) (*entity.MediaContent, error) {
	qMedia := `
		SELECT id, name, description, type, duration, created_at, updated_at
		FROM media_content
		WHERE id = $1
	`

	row := r.db.QueryRow(ctx, qMedia, id)

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

	qCategories := `
		SELECT c.id, c.name, c.description
		FROM categories c
		JOIN media_content_categories mcc ON c.id = mcc.category_id
		WHERE mcc.media_content_id = $1
	`

	rows, err := r.db.Query(ctx, qCategories, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description); err != nil {
			return nil, err
		}

		m.Categories = append(m.Categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &m, nil
}

func (r *MediaContentRepository) Delete(ctx context.Context, id string) error {
	q := `DELETE FROM media_content WHERE id = $1`
	_, err := r.db.Exec(ctx, q, id)
	return err
}

func (r *MediaContentRepository) GetRandomByCategories(
	ctx context.Context,
	categoryIDs []string,
	limit int64,
) ([]entity.MediaContent, error) {
	if len(categoryIDs) == 0 {
		return []entity.MediaContent{}, nil
	}

	q := `
		SELECT DISTINCT mc.id, mc.name, mc.description, mc.type, mc.duration, mc.created_at, mc.updated_at
		FROM media_content mc
		JOIN media_content_categories mcc ON mc.id = mcc.media_content_id
		WHERE mcc.category_id = ANY($1)
		ORDER BY RANDOM()
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, q, categoryIDs, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mediaList []entity.MediaContent
	for rows.Next() {
		var m entity.MediaContent
		if err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.Description,
			&m.Type,
			&m.Duration,
			&m.CreatedAt,
			&m.UpdatedAt,
		); err != nil {
			return nil, err
		}
		mediaList = append(mediaList, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mediaList, nil
}
