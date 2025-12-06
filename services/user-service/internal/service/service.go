package service

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewAuthService,
			NewUsersService,
			NewOutboxPublisher,
			NewUserSagaHandler,
		),
		fx.Invoke(RunOutboxReader),
		fx.Invoke(RunSagaConsumer),
	)
}
