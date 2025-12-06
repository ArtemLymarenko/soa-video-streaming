package amqpcontrollers

import (
	"context"
	"encoding/json"

	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"
	"go.uber.org/fx"

	"soa-video-streaming/pkg/saga"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewEventsController),
	)
}

type EventsController struct {
	coordinator *saga.Coordinator
}

func NewEventsController(coordinator *saga.Coordinator) *EventsController {
	return &EventsController{
		coordinator: coordinator,
	}
}

func (c *EventsController) HandleUserSignUpEvent(d gorabbit.Delivery) gorabbit.Action {
	var msg saga.Message
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal message")
		return gorabbit.NackDiscard
	}

	if err := c.coordinator.HandleEvent(context.Background(), &msg); err != nil {
		logrus.WithError(err).Error("Failed to handle event")
		return gorabbit.NackRequeue
	}

	return gorabbit.Ack
}

func (c *EventsController) HandleBucketEvents(d gorabbit.Delivery) gorabbit.Action {
	var msg saga.Message
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal message")
		return gorabbit.NackDiscard
	}

	if err := c.coordinator.HandleEvent(context.Background(), &msg); err != nil {
		logrus.WithError(err).Error("Failed to handle event")
		return gorabbit.NackRequeue
	}

	return gorabbit.Ack
}

func (c *EventsController) HandleEmailEvents(d gorabbit.Delivery) gorabbit.Action {
	var msg saga.Message
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal message")
		return gorabbit.NackDiscard
	}

	if err := c.coordinator.HandleEvent(context.Background(), &msg); err != nil {
		logrus.WithError(err).Error("Failed to handle event")
		return gorabbit.NackRequeue
	}

	return gorabbit.Ack
}
