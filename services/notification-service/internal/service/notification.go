package service

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"
	"go.uber.org/fx"
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/notification-service/pkg/notifications"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewNotificationService),
		fx.Invoke(func(n *NotificationService) error {
			return n.RunSignUpEventHandler()
		}),
	)
}

type NotificationService struct {
	client *rabbitmq.Client
}

func NewNotificationService(client *rabbitmq.Client) *NotificationService {
	return &NotificationService{
		client: client,
	}
}

func (n *NotificationService) RunSignUpEventHandler() error {
	consumer, err := n.client.NewConsumer(notifications.SignUpEventQueueName)
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
