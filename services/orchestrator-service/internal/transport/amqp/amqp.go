package amqptransport

import (
	"context"

	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/orchestrator-service/domain"

	"soa-video-streaming/pkg/saga"

	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Invoke(RunConsumers),
	)
}

type Consumer struct {
	Consumer *gorabbit.Consumer
	Handler  func(d gorabbit.Delivery) gorabbit.Action
}

func GetAllConsumers(client *rabbitmq.Client, eventsController *saga.RabbitMQEventsController) ([]Consumer, error) {
	var consumers []Consumer
	userSignUpConsumer, err := gorabbit.NewConsumer(
		client.Conn,
		domain.QueueUserSignUp,
		gorabbit.WithConsumerOptionsLogger(logrus.StandardLogger()),
		gorabbit.WithConsumerOptionsQueueDurable,
	)
	if err != nil {
		return nil, err
	}

	consumers = append(consumers, Consumer{
		Consumer: userSignUpConsumer,
		Handler:  eventsController.HandleEvent,
	})

	bucketConsumer, err := gorabbit.NewConsumer(
		client.Conn,
		domain.QueueBucketEvents,
		gorabbit.WithConsumerOptionsLogger(logrus.StandardLogger()),
		gorabbit.WithConsumerOptionsQueueDurable,
	)
	if err != nil {
		return nil, err
	}

	consumers = append(consumers, Consumer{
		Consumer: bucketConsumer,
		Handler:  eventsController.HandleEvent,
	})

	emailConsumer, err := gorabbit.NewConsumer(
		client.Conn,
		domain.QueueEmailEvents,
		gorabbit.WithConsumerOptionsLogger(logrus.StandardLogger()),
		gorabbit.WithConsumerOptionsQueueDurable,
	)
	if err != nil {
		return nil, err
	}

	consumers = append(consumers, Consumer{
		Consumer: emailConsumer,
		Handler:  eventsController.HandleEvent,
	})

	return consumers, nil
}

func RunConsumers(lc fx.Lifecycle, eventsController *saga.RabbitMQEventsController, client *rabbitmq.Client) error {
	var consumers []*gorabbit.Consumer

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			allConsumers, err := GetAllConsumers(client, eventsController)
			if err != nil {
				return err
			}

			consumers = make([]*gorabbit.Consumer, len(allConsumers))
			for i, consumer := range allConsumers {
				consumers[i] = consumer.Consumer
				go func(c Consumer) {
					if err := c.Consumer.Run(c.Handler); err != nil {
						logrus.WithError(err).Error("Consumer stopped")
					}
				}(consumer)
			}

			logrus.Info("Orchestrator consumers started")
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
