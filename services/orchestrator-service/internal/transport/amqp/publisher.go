package amqptransport

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"

	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/pkg/saga"

	gorabbit "github.com/wagslane/go-rabbitmq"
)

type Publisher struct {
	client *rabbitmq.Client
	pub    *gorabbit.Publisher
}

func NewPublisher(client *rabbitmq.Client) (saga.MessagePublisher, error) {
	publisher, err := gorabbit.NewPublisher(client.Conn, gorabbit.WithPublisherOptionsLogger(logrus.StandardLogger()))
	if err != nil {
		return nil, err
	}

	return &Publisher{
		client: client,
		pub:    publisher,
	}, nil
}

func (p *Publisher) PublishCommand(ctx context.Context, queue string, msg *saga.Message) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	return p.pub.Publish(
		body,
		[]string{queue},
		gorabbit.WithPublishOptionsContentType("application/json"),
		gorabbit.WithPublishOptionsPersistentDelivery,
	)
}
