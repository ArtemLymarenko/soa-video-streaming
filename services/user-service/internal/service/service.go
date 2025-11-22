package service

import (
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/notification-service/pkg/notifications"

	"go.uber.org/fx"
)

func ProvideSignUpPublisher(lc fx.Lifecycle, client *rabbitmq.Client) (*rabbitmq.Publisher, error) {
	return client.NewPublisher(lc, "", notifications.SignUpEventQueueName)
}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewAuthService,
			NewUsersService,
			ProvideSignUpPublisher,
		),
	)
}
