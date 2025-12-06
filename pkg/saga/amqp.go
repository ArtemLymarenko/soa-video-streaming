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
		logrus.WithError(err).Error("Failed to unmarshal message")
		return gorabbit.NackDiscard
	}

	if err := c.coordinator.HandleEvent(context.Background(), &msg); err != nil {
		logrus.WithError(err).Error("Failed to handle event")
		return gorabbit.NackRequeue
	}

	return gorabbit.Ack
}
