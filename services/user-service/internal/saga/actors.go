package saga

import (
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/orchestrator-service/domain"

	"go.uber.org/fx"
)

func RegisterUserActor(
	lc fx.Lifecycle,
	client *rabbitmq.Client,
	handler *UserSagaHandler,
) *saga.Actor {
	actor := saga.NewActor(
		lc,
		client.Conn,
		nil, // No Outbox used in specific actor registration
		domain.QueueUserCommands,
	)

	actor.Register(
		domain.CmdCompensateUser,
		handler.HandleCompensateUser,
		domain.EventUserCompensated,
		domain.QueueUserEvents,
	)

	return actor
}
