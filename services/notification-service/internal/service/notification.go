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
	client    *rabbitmq.Client
	publisher *rabbitmq.Publisher
}

func NewNotificationService(client *rabbitmq.Client) (*NotificationService, error) {
	return &NotificationService{
		client:    client,
		publisher: client.NewPublisher(),
	}, nil
}

func (n *NotificationService) RunSignUpEventHandler() error {
	consumer := n.client.NewConsumer(notifications.SignUpEventQueueName)

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

const maxSignUpRetries = 3

func (n *NotificationService) RetrySignUp(d gorabbit.Delivery) gorabbit.Action {
	var retryCount int
	if v, ok := d.Headers["x-retry-count"].(int32); ok {
		retryCount = int(v)
	}

	if retryCount >= maxSignUpRetries {
		logrus.Warnf("Retry limit reached (%d). Discarding message", retryCount)
		return gorabbit.NackDiscard
	}

	retryCount++

	err := n.publisher.Publish(d.Body, []string{notifications.SignUpEventQueueName}, gorabbit.WithPublishOptionsHeaders(gorabbit.Table{
		"x-retry-count": retryCount,
	}))
	if err != nil {
		logrus.WithError(err).Error("Failed to republish message during retry")
		return gorabbit.NackRequeue
	}

	logrus.Info("Successfully republished message during retry")

	return gorabbit.Ack
}
