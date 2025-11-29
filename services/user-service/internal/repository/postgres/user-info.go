package postgres

import (
	"context"
	"soa-video-streaming/pkg/postgres"
	"soa-video-streaming/services/user-service/internal/domain/entity"

	"github.com/jackc/pgx/v5"
)

type UserInfoRepository struct {
	db postgres.DB
}

func NewUserInfoRepository(db postgres.DB) *UserInfoRepository {
	return &UserInfoRepository{
		db: db,
	}
}

const saveQuery = `INSERT INTO user_info(user_id, first_name, last_name) VALUES ($1, $2, $3)`

func (r *UserInfoRepository) Save(ctx context.Context, userId string, userInfo entity.UserInfo) error {
	_, err := r.db.Exec(ctx, saveQuery, userId, userInfo.FirstName, userInfo.LastName)
	return err
}

func (r *UserInfoRepository) WithTx(tx pgx.Tx) *UserInfoRepository {
	return &UserInfoRepository{
		db: tx,
	}
}
