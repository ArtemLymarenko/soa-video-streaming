package saga

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"
	"go.uber.org/fx"
)

type CommandHandler func(ctx context.Context, msg *Message) (any, error)

type actorHandler struct {
	handler      CommandHandler
	successEvent string
	replyQueue   string
}

type Actor struct {
	client     *gorabbit.Consumer
	handlers   map[string]actorHandler
	outboxRepo OutboxRepository
	publisher  *gorabbit.Publisher
}

func NewActor(lc fx.Lifecycle, conn *gorabbit.Conn, outboxRepo OutboxRepository, queue string) *Actor {
	publisher, err := gorabbit.NewPublisher(
		conn,
		gorabbit.WithPublisherOptionsLogger(logrus.StandardLogger()),
	)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create actor publisher")
	}

	actor := &Actor{
		handlers:   make(map[string]actorHandler),
		outboxRepo: outboxRepo,
		publisher:  publisher,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			consumer, err := gorabbit.NewConsumer(
				conn,
				queue,
				gorabbit.WithConsumerOptionsLogger(logrus.StandardLogger()),
				gorabbit.WithConsumerOptionsQueueDurable,
				gorabbit.WithConsumerOptionsQueueArgs(map[string]any{
					"x-dead-letter-exchange":    "",
					"x-dead-letter-routing-key": "queue.saga.errors",
				}),
			)
			if err != nil {
				return fmt.Errorf("create consumer: %w", err)
			}

			actor.client = consumer

			go func() {
				if err := consumer.Run(actor.handleMessage); err != nil {
					logrus.WithError(err).Error("Actor consumer stopped")
				}
			}()

			logrus.WithField("queue", queue).Info("Saga Actor started")
			return nil
		},
		OnStop: func(ctx context.Context) error {
			if actor.client != nil {
				actor.client.Close()
			}
			if actor.publisher != nil {
				actor.publisher.Close()
			}
			return nil
		},
	})

	return actor
}

func (a *Actor) Register(cmdType string, handler CommandHandler, successEvent, replyQueue string) {
	a.handlers[cmdType] = actorHandler{
		handler:      handler,
		successEvent: successEvent,
		replyQueue:   replyQueue,
	}
}

func (a *Actor) handleMessage(d gorabbit.Delivery) gorabbit.Action {
	var msg Message
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		return gorabbit.NackDiscard
	}

	handler, ok := a.handlers[msg.Type]
	if !ok {
		return gorabbit.NackDiscard
	}

	result, err := handler.handler(context.Background(), &msg)
	if err != nil {
		logrus.WithError(err).WithField("cmd", msg.Type).Error("Command failed, sending to DLQ")
		return gorabbit.NackDiscard
	}

	if handler.successEvent != "" {
		if err := a.sendReply(context.Background(), msg.CorrelationID, handler.successEvent, handler.replyQueue, result); err != nil {
			logrus.WithError(err).Error("Failed to send reply event")
			return gorabbit.NackRequeue
		}
	}

	return gorabbit.Ack
}

func (a *Actor) sendReply(ctx context.Context, correlationID, eventType, queue string, payload any) error {
	if eventType == "" {
		return nil
	}

	replyMsg, err := NewSagaMessage(correlationID, eventType, payload)
	if err != nil {
		return fmt.Errorf("create reply message: %w", err)
	}

	raw, err := json.Marshal(replyMsg)
	if err != nil {
		return fmt.Errorf("marshal reply: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"event":          eventType,
		"queue":          queue,
	}).Info("Publishing reply event")

	return a.publisher.PublishWithContext(
		ctx,
		raw,
		[]string{queue},
		gorabbit.WithPublishOptionsContentType("application/json"),
	)
}
