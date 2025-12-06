package service

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"
	"go.uber.org/fx"

	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/pkg/saga"
)

// RunConsumers starts all saga event consumers
func RunConsumers(lc fx.Lifecycle, coordinator *Coordinator, client *rabbitmq.Client) error {
	var consumers []*gorabbit.Consumer

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Consumer for EventUserSignUp
			userSignUpConsumer, err := gorabbit.NewConsumer(
				client.Conn,
				saga.QueueUserSignUp,
				gorabbit.WithConsumerOptionsLogger(logrus.StandardLogger()),
				gorabbit.WithConsumerOptionsQueueDurable,
			)
			if err != nil {
				return err
			}
			consumers = append(consumers, userSignUpConsumer)

			// Consumer for EventBucketCreated and EventBucketFailed
			bucketConsumer, err := gorabbit.NewConsumer(
				client.Conn,
				saga.QueueBucketEvents,
				gorabbit.WithConsumerOptionsLogger(logrus.StandardLogger()),
				gorabbit.WithConsumerOptionsQueueDurable,
			)
			if err != nil {
				return err
			}
			consumers = append(consumers, bucketConsumer)

			// Consumer for EventEmailSent and EventEmailFailed
			emailConsumer, err := gorabbit.NewConsumer(
				client.Conn,
				saga.QueueEmailEvents,
				gorabbit.WithConsumerOptionsLogger(logrus.StandardLogger()),
				gorabbit.WithConsumerOptionsQueueDurable,
			)
			if err != nil {
				return err
			}
			consumers = append(consumers, emailConsumer)

			// Start consuming
			go func() {
				if err := userSignUpConsumer.Run(handleUserSignUpEvent(coordinator)); err != nil {
					logrus.WithError(err).Error("UserSignUp consumer stopped")
				}
			}()

			go func() {
				if err := bucketConsumer.Run(handleBucketEvents(coordinator)); err != nil {
					logrus.WithError(err).Error("Bucket consumer stopped")
				}
			}()

			go func() {
				if err := emailConsumer.Run(handleEmailEvents(coordinator)); err != nil {
					logrus.WithError(err).Error("Email consumer stopped")
				}
			}()

			logrus.Info("✅ Orchestrator consumers started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			for _, consumer := range consumers {
				consumer.Close()
			}
			return nil
		},
	})

	return nil
}

// handleUserSignUpEvent creates a handler for EventUserSignUp
func handleUserSignUpEvent(coordinator *Coordinator) func(d gorabbit.Delivery) gorabbit.Action {
	return func(d gorabbit.Delivery) gorabbit.Action {
		var msg saga.SagaMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			logrus.WithError(err).Error("Failed to unmarshal saga message")
			return gorabbit.NackDiscard
		}

		if msg.Type != saga.EventUserSignUp {
			logrus.WithField("type", msg.Type).Warn("Unexpected message type in UserSignUp queue")
			return gorabbit.NackDiscard
		}

		if err := coordinator.HandleUserSignUp(context.Background(), &msg); err != nil {
			logrus.WithError(err).Error("Failed to handle EventUserSignUp")
			return gorabbit.NackRequeue
		}

		return gorabbit.Ack
	}
}

// handleBucketEvents creates a handler for bucket events
func handleBucketEvents(coordinator *Coordinator) func(d gorabbit.Delivery) gorabbit.Action {
	return func(d gorabbit.Delivery) gorabbit.Action {
		var msg saga.SagaMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			logrus.WithError(err).Error("Failed to unmarshal saga message")
			return gorabbit.NackDiscard
		}

		var err error
		switch msg.Type {
		case saga.EventBucketCreated:
			err = coordinator.HandleBucketCreated(context.Background(), &msg)
		case saga.EventBucketFailed:
			err = coordinator.HandleBucketFailed(context.Background(), &msg)
		default:
			logrus.WithField("type", msg.Type).Warn("Unexpected message type in Bucket queue")
			return gorabbit.NackDiscard
		}

		if err != nil {
			logrus.WithError(err).WithField("type", msg.Type).Error("Failed to handle bucket event")
			return gorabbit.NackRequeue
		}

		return gorabbit.Ack
	}
}

// handleEmailEvents creates a handler for email events
func handleEmailEvents(coordinator *Coordinator) func(d gorabbit.Delivery) gorabbit.Action {
	return func(d gorabbit.Delivery) gorabbit.Action {
		var msg saga.SagaMessage
		if err := json.Unmarshal(d.Body, &msg); err != nil {
			logrus.WithError(err).Error("Failed to unmarshal saga message")
			return gorabbit.NackDiscard
		}

		var err error
		switch msg.Type {
		case saga.EventEmailSent:
			err = coordinator.HandleEmailSent(context.Background(), &msg)
		case saga.EventEmailFailed:
			err = coordinator.HandleEmailFailed(context.Background(), &msg)
		default:
			logrus.WithField("type", msg.Type).Warn("Unexpected message type in Email queue")
			return gorabbit.NackDiscard
		}

		if err != nil {
			logrus.WithError(err).WithField("type", msg.Type).Error("Failed to handle email event")
			return gorabbit.NackRequeue
		}

		return gorabbit.Ack
	}
}
