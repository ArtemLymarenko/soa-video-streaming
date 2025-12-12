package saga

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"
	"go.uber.org/fx"
)

func ModuleRabbitMQEventsController() fx.Option {
	return fx.Options(
		fx.Provide(NewRabbitMQEventsController),
	)
}

type RabbitMQEventsController struct {
	coordinator *Coordinator
}

func NewRabbitMQEventsController(coordinator *Coordinator) *RabbitMQEventsController {
	return &RabbitMQEventsController{
		coordinator: coordinator,
	}
}

func (c *RabbitMQEventsController) HandleEvent(d gorabbit.Delivery) gorabbit.Action {
	var msg Message
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal event")
		return gorabbit.NackDiscard
	}

	if err := c.coordinator.HandleEvent(context.Background(), &msg); err != nil {
		logrus.WithError(err).Error("Failed to handle saga event")
		return gorabbit.NackDiscard
	}

	return gorabbit.Ack
}

func (c *RabbitMQEventsController) HandleFailure(d gorabbit.Delivery) gorabbit.Action {
	var msg Message
	if err := json.Unmarshal(d.Body, &msg); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal failure message")
		return gorabbit.NackDiscard
	}

	logrus.WithField("correlation_id", msg.CorrelationID).
		Infof("Processing failure from DLQ for command: %s", msg.Type)

	if err := c.coordinator.HandleFailure(context.Background(), &msg); err != nil {
		logrus.WithError(err).Error("Failed to compensate saga")
		return gorabbit.NackDiscard
	}

	return gorabbit.Ack
}
