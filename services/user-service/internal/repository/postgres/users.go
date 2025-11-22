package postgres

import (
	"context"
	"errors"
	"fmt"
	"soa-video-streaming/pkg/postgres"
	"soa-video-streaming/services/user-service/internal/domain/entity"

	"github.com/jackc/pgx/v5"
)

type UsersRepository struct {
	db           postgres.DB
	client       *postgres.Client
	userInfoRepo *UserInfoRepository
}

func NewUsersRepository(db postgres.DB, client *postgres.Client, userInfoRepo *UserInfoRepository) *UsersRepository {
	return &UsersRepository{
		db:           db,
		userInfoRepo: userInfoRepo,
		client:       client,
	}
}

func (r *UsersRepository) findOne(ctx context.Context, query string, args ...interface{}) (entity.User, error) {
	row := r.db.QueryRow(ctx, query, args...)

	var user entity.User
	err := row.Scan(
		&user.Id,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.UserInfo.FirstName,
		&user.UserInfo.LastName,
		&user.UserInfo.CreatedAt,
		&user.UserInfo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, nil
		}

		return entity.User{}, err
	}

	return user, nil
}

const findQuery = `SELECT
		u.id, u.email, u.password, u.created_at, u.updated_at,
		ui.first_name, ui.last_name, ui.created_at, ui.updated_at
		FROM user_service.users AS u
		INNER JOIN user_service.user_info AS ui ON u.id = ui.user_id
		WHERE %s = $1`

func (r *UsersRepository) FindById(ctx context.Context, id string) (entity.User, error) {
	q := fmt.Sprintf(findQuery, "u.id")
	return r.findOne(ctx, q, id)
}

func (r *UsersRepository) FindByEmail(ctx context.Context, email string) (entity.User, error) {
	q := fmt.Sprintf(findQuery, "u.email")
	return r.findOne(ctx, q, email)
}

func (r *UsersRepository) Save(ctx context.Context, user entity.User) error {
	err := r.client.Tx(ctx, func(tx pgx.Tx) error {
		q := `INSERT INTO user_service.users(id, email, password) VALUES ($1, $2, $3)`

		_, err := tx.Exec(ctx, q, user.Id, user.Email, user.Password)
		if err != nil {
			return err
		}

		err = r.userInfoRepo.WithTX(tx).Save(ctx, user.Id, user.UserInfo)
		return err
	})

	return err
}
