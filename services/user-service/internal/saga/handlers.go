package saga

import (
	"context"
	"encoding/json"
	"fmt"

	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/user-service/internal/repository/postgres"

	"github.com/sirupsen/logrus"
)

type UserSagaHandler struct {
	usersRepo *postgres.UsersRepository
}

func NewUserSagaHandler(usersRepo *postgres.UsersRepository) *UserSagaHandler {
	return &UserSagaHandler{
		usersRepo: usersRepo,
	}
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
