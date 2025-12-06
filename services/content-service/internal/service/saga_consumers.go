package service

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"
	"go.uber.org/fx"

	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/orchestrator-service/domain"
)

func RunSagaConsumers(lc fx.Lifecycle, handler *BucketHandler, client *rabbitmq.Client) error {
	var consumer *gorabbit.Consumer

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			var err error
			consumer, err = gorabbit.NewConsumer(
				client.Conn,
				domain.QueueContentCommands,
				gorabbit.WithConsumerOptionsLogger(logrus.StandardLogger()),
				gorabbit.WithConsumerOptionsQueueDurable,
			)
			if err != nil {
				return err
			}

			go func() {
				if err := consumer.Run(handleContentCommands(handler)); err != nil {
					logrus.WithError(err).Error("Content commands consumer stopped")
				}
			}()

			logrus.Info("Content Service saga consumers started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if consumer != nil {
				consumer.Close()
			}
			return nil
		},
	})

	return nil
}

func handleContentCommands(handler *BucketHandler) func(d gorabbit.Delivery) gorabbit.Action {
	return func(d gorabbit.Delivery) gorabbit.Action {
		var msg saga.Message
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			logrus.WithError(err).Error("Failed to unmarshal saga message")
			return gorabbit.NackDiscard
		}

		var err error
		switch msg.Type {
		case domain.CmdCreateBucket:
			err = handler.HandleCreateBucket(context.Background(), &msg)
		case domain.CmdCompensateBucket:
			err = handler.HandleCompensateBucket(context.Background(), &msg)
		default:
			logrus.WithField("type", msg.Type).Warn("Unexpected message type in Content commands queue")
			return gorabbit.NackDiscard
		}

		if err != nil {
			logrus.WithError(err).WithField("type", msg.Type).Error("Failed to handle content command")
			return gorabbit.NackRequeue
		}

		return gorabbit.Ack
	}
}
