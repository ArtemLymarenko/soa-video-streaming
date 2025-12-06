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
	failureEvent string
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

func (a *Actor) Register(cmdType string, handler CommandHandler, successEvent, failureEvent, replyQueue string) {
	a.handlers[cmdType] = actorHandler{
		handler:      handler,
		successEvent: successEvent,
		failureEvent: failureEvent,
		replyQueue:   replyQueue,
	}
}

func (a *Actor) handleMessage(d gorabbit.Delivery) gorabbit.Action {
	var msg Message
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal saga message")
		return gorabbit.NackDiscard
	}

	handler, ok := a.handlers[msg.Type]
	if !ok {
		logrus.WithField("type", msg.Type).Warn("No handler registered for command")
		return gorabbit.NackDiscard
	}

	logrus.WithFields(logrus.Fields{
		"type":           msg.Type,
		"correlation_id": msg.CorrelationID,
	}).Info("Handling saga command")

	// Execute handler
	result, err := handler.handler(context.Background(), &msg)

	// Determine response type
	respType := handler.successEvent
	var payload any = result

	if err != nil {
		respType = handler.failureEvent
		payload = map[string]string{"error": err.Error()}
		logrus.WithError(err).WithField("type", msg.Type).Error("Handler failed")
	}

	// Send reply
	if err := a.sendReply(context.Background(), msg.CorrelationID, respType, handler.replyQueue, payload); err != nil {
		logrus.WithError(err).Error("Failed to send reply")
		return gorabbit.NackRequeue
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
