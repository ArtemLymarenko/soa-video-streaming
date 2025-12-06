package service

import (
	"context"
	"encoding/json"
	"fmt"

	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/orchestrator-service/domain"
	"soa-video-streaming/services/user-service/internal/repository/postgres"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type UserSagaHandler struct {
	usersRepo *postgres.UsersRepository
}

func NewUserSagaHandler(usersRepo *postgres.UsersRepository) *UserSagaHandler {
	return &UserSagaHandler{
		usersRepo: usersRepo,
	}
}

func RunSagaConsumer(
	lc fx.Lifecycle,
	client *rabbitmq.Client,
	handler *UserSagaHandler,
	outboxRepo *postgres.OutboxRepository,
) *saga.Actor {
	actor := saga.NewActor(
		lc,
		client.Conn,
		nil, // Passing nil for OutboxRepository as Actor doesn't fully support it yet
		domain.QueueUserCommands,
	)

	actor.Register(
		domain.CmdCompensateUser,
		handler.HandleCompensateUser,
		domain.EventUserCompensated, // Assuming this event exists or we define it
		domain.EventUserCompensationFailed,
		domain.QueueUserEvents, // Reply queue
	)

	return actor
}

func (h *UserSagaHandler) HandleCompensateUser(ctx context.Context, msg *saga.Message) (any, error) {
	var payload struct {
		UserID string `json:"user_id"`
	}

	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	if err := h.usersRepo.Delete(ctx, payload.UserID); err != nil {
		logrus.WithError(err).Error("Failed to delete user for compensation")
		return nil, err
	}

	return nil, nil
}
