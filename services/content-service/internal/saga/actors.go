package saga

import (
	"go.uber.org/fx"

	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/orchestrator-service/domain"
)

func RegisterBucketsActor(lc fx.Lifecycle, service *BucketsService, client *rabbitmq.Client) *saga.Actor {
	actor := saga.NewActor(
		lc,
		client.Conn,
		nil, // No Outbox used in Content Service yet
		domain.QueueContentCommands,
	)

	actor.Register(
		domain.CmdCreateBucket,
		service.HandleCreateBucket,
		domain.EventBucketCreated,
		domain.QueueContentEvents,
	)

	actor.Register(
		domain.CmdCompensateBucket,
		service.HandleCompensateBucket,
		"", // No success event
		"", // No reply queue needed
	)

	return actor
}
