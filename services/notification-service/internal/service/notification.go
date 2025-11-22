package service

import (
	"context"
	"encoding/json"
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/notification-service/pkg/notifications"

	"github.com/rabbitmq/amqp091-go"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewNotificationService),
		fx.Invoke(func(lc fx.Lifecycle, svc *NotificationService) {
			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					go func() {
						if err := svc.StartConsumer(context.Background()); err != nil {
							logrus.WithError(err).Error("Consumer stopped")
						}
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

func (s *NotificationService) StartConsumer(ctx context.Context) error {
	_, ch, err := s.client.DeclareQueue(notifications.SignUpEventQueueName, rabbitmq.QueueOptions{
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	logrus.Info("Starting to consume messages from user.signup queue")

	return s.client.ConsumeQueue(
		ctx,
		"user.signup",
		rabbitmq.ConsumeOptions{
			Consumer:  "",
			AutoAck:   false,
			Exclusive: false,
			NoLocal:   false,
			NoWait:    false,
		},
		s.handleSignUpEvent,
	)
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
