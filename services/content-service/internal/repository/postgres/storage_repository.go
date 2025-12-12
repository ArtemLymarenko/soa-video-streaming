package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"soa-video-streaming/services/content-service/internal/domain/entity"
)

type StorageRepository struct {
	pool *pgxpool.Pool
}

func NewStorageRepository(pool *pgxpool.Pool) *StorageRepository {
	return &StorageRepository{pool: pool}
}

func (r *StorageRepository) Create(ctx context.Context, account entity.StorageAccount) error {
	query := `
		INSERT INTO user_storage_accounts (user_id, bucket_name, status, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, account.UserID, account.BucketName, account.Status, account.CreatedAt)
	return err
}

func (r *StorageRepository) Delete(ctx context.Context, userID string) error {
	query := `DELETE FROM user_storage_accounts WHERE user_id = $1`
	_, err := r.pool.Exec(ctx, query, userID)
	return err
}

func (r *StorageRepository) FindByUserID(ctx context.Context, userID string) (*entity.StorageAccount, error) {
	query := `
		SELECT user_id, bucket_name, status, created_at
		FROM user_storage_accounts
		WHERE user_id = $1
	`

	var account entity.StorageAccount
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&account.UserID,
		&account.BucketName,
		&account.Status,
		&account.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &account, nil
}
