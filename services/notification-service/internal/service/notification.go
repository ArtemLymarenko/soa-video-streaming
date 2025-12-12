package service

import (
	"encoding/json"
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/notification-service/internal/config"
	notificationsaga "soa-video-streaming/services/notification-service/internal/saga"
	"soa-video-streaming/services/notification-service/pkg/notifications"

	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewNotificationService,
			notificationsaga.NewNotificationSagaHandler,
		),
		fx.Invoke(notificationsaga.RegisterNotificationActor),
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

func (n *NotificationService) RunSignUpEventHandler() error {
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

	handler := func(d gorabbit.Delivery) gorabbit.Action {
		var event notifications.EventSignUp

		if err := json.Unmarshal(d.Body, &event); err != nil {
			logrus.WithError(err).Error("Failed to unmarshal signup event")
			return gorabbit.NackDiscard
		}

		logrus.WithFields(logrus.Fields{
			"user_id":    event.UserID,
			"email":      event.Email,
			"message":    event.Message,
			"created_at": event.CreatedAt,
		}).Info("📧 User SignUp Event Received")

		return gorabbit.Ack
	}

	return consumer.Run(handler)
}
