package saga

import (
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/orchestrator-service/domain"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewNotificationSagaHandler),
		fx.Invoke(RegisterNotificationActor),
	)
}

func RegisterNotificationActor(
	lc fx.Lifecycle,
	client *rabbitmq.Client,
	handler *NotificationSagaHandler,
) *saga.Actor {
	actor := saga.NewActor(
		lc,
		client.Conn,
		nil, // No Outbox as requested
		domain.QueueNotificationCommands,
	)

	actor.Register(
		domain.CmdSendEmail,
		handler.HandleSendEmail,
		domain.EventEmailSent,
		domain.QueueNotificationEvents,
	)

	return actor
}
