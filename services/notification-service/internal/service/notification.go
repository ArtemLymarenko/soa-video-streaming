package service

import (
	"context"
	"encoding/json"
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/notification-service/internal/config"
	"soa-video-streaming/services/notification-service/pkg/notifications"

	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewNotificationService,
		),
		fx.Invoke(func(lc fx.Lifecycle, n *NotificationService) {
			n.RegisterHooks(lc)
		}),
	)
}

type NotificationService struct {
	client *rabbitmq.Client
	cfg    *config.AppConfig
}

func NewNotificationService(client *rabbitmq.Client, cfg *config.AppConfig) (*NotificationService, error) {
	return &NotificationService{
		client: client,
		cfg:    cfg,
	}, nil
}

func (n *NotificationService) RegisterHooks(lc fx.Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			consumer, err := gorabbit.NewConsumer(
				n.client.Conn,
				notifications.QueueSignUpEvent,
				gorabbit.WithConsumerOptionsLogger(logrus.StandardLogger()),
				gorabbit.WithConsumerOptionsQueueDurable,
				gorabbit.WithConsumerOptionsQueueArgs(map[string]interface{}{
					"x-dead-letter-exchange": "global.dlx",
				}),
			)
			if err != nil {
				return err
			}

			go func() {
				err := consumer.Run(func(d gorabbit.Delivery) gorabbit.Action {
					var event notifications.EventSignUp
					if err := json.Unmarshal(d.Body, &event); err != nil {
						logrus.WithError(err).Error("Failed to unmarshal signup event")
						return gorabbit.NackDiscard
					}
					logrus.WithFields(logrus.Fields{
						"email": event.Email,
					}).Info("📧 User SignUp Event Received")
					return gorabbit.Ack
				})

				if err != nil {
					logrus.WithError(err).Error("SignUp Consumer stopped")
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return nil
		},
	})
}
