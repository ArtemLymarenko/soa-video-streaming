package saga

import (
	"context"
	"encoding/json"
	"fmt"

	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/orchestrator-service/domain"

	"github.com/sirupsen/logrus"
)

type NotificationSagaHandler struct {
	// dependencies
}

func NewNotificationSagaHandler() *NotificationSagaHandler {
	return &NotificationSagaHandler{}
}

func (h *NotificationSagaHandler) HandleSendEmail(ctx context.Context, msg *saga.Message) (any, error) {
	var payload domain.EmailPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	if payload.FirstName == "Artem" {
		return nil, fmt.Errorf("first name is required")
	}

	// Logic to send email would go here.
	logrus.WithFields(logrus.Fields{
		"user_id": payload.UserID,
		"email":   payload.Email,
	}).Info("Sending email (simulated)")

	return domain.EmailPayload{
		UserID: payload.UserID,
		Email:  payload.Email,
	}, nil
}
