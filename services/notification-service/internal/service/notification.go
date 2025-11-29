package service

import (
	"context"
	"encoding/json"
	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/notification-service/pkg/notifications"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewNotificationService),
		fx.Invoke(func(lc fx.Lifecycle, client *rabbitmq.Client, svc *NotificationService) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						client.Consume(ctx, notifications.SignUpEventQueueName, svc.handleSignUpEvent)
					}()
					return nil
				},
			})
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

func (s *NotificationService) handleSignUpEvent(ctx context.Context, d amqp091.Delivery) error {
	var event notifications.EventSignUp

	if err := json.Unmarshal(d.Body, &event); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal signup event")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"user_id":    event.UserID,
		"email":      event.Email,
		"message":    event.Message,
		"created_at": event.CreatedAt,
	}).Info("📧 User SignUp Event Received")

	return nil
}
