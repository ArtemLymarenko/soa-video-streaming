package service

import (
	usersaga "soa-video-streaming/services/user-service/internal/saga"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewAuthService,
			NewUsersService,
			NewOutboxPublisher,
			usersaga.NewUserSagaHandler,
		),
		fx.Invoke(RunOutboxReader),
		fx.Invoke(usersaga.RegisterUserActor),
	)
}
