package service

import (
	"context"
	"encoding/json"

	"github.com/wagslane/go-rabbitmq"

	pkgrabbitmq "soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/pkg/saga"
)

type RabbitMQPublisher struct {
	client *pkgrabbitmq.Client
}

func NewRabbitMQPublisher(client *pkgrabbitmq.Client) *RabbitMQPublisher {
	return &RabbitMQPublisher{
		client: client,
	}
}

func (p *RabbitMQPublisher) PublishCommand(ctx context.Context, queue string, msg *saga.Message) error {
	publisher, err := rabbitmq.NewPublisher(
		p.client.Conn,
		rabbitmq.WithPublisherOptionsExchangeName(""),
	)
	if err != nil {
		return err
	}
	defer publisher.Close()

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return publisher.Publish(
		msgBytes,
		[]string{queue},
		rabbitmq.WithPublishOptionsContentType("application/json"),
		rabbitmq.WithPublishOptionsPersistentDelivery,
	)
}
