package service

import (
	"context"
	"soa-video-streaming/services/user-service/internal/domain/entity"
	"soa-video-streaming/services/user-service/internal/repository/postgres"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewAuthService,
			NewUsersService,
		),
	)
}

type AuthService struct {
	usersRepo *postgres.UsersRepository
}

func NewAuthService(usersRepo *postgres.UsersRepository) *AuthService {
	return &AuthService{
		usersRepo: usersRepo,
	}
}

func (u *AuthService) SignUp(ctx context.Context, user *entity.User) error {
	return nil
}

func (u *AuthService) SignIn(ctx context.Context, user *entity.User) error {
	return nil
}
