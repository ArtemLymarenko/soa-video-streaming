package service

import (
	"context"
	"encoding/json"

	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/orchestrator-service/domain"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewRegisterUserWorkflow,
			NewRabbitMQPublisher,
		),
		fx.Invoke(func(rw *RegisterUserWorkflow) {
			rw.Register()
		}),
	)
}

type RegisterUserWorkflow struct {
	coordinator *saga.Coordinator
}

func NewRegisterUserWorkflow(coordinator *saga.Coordinator) *RegisterUserWorkflow {
	return &RegisterUserWorkflow{
		coordinator: coordinator,
	}
}

func (w *RegisterUserWorkflow) Register() {
	w.coordinator.
		RegisterCommandQueue(map[string]string{
			domain.CmdCreateBucket: domain.QueueContentCommands,
			domain.CmdSendEmail:    domain.QueueNotificationCommands,
		}).
		On(domain.EventUserSignUp, w.HandleUserSignUp).
		On(domain.EventBucketCreated, w.HandleBucketCreated).
		On(domain.EventEmailSent, w.HandleEmailSent).
		OnFailure(domain.EventBucketFailed, domain.CmdCompensateUser).
		OnFailure(domain.EventEmailFailed, domain.CmdCompensateBucket, domain.CmdCompensateUser)
}

func (w *RegisterUserWorkflow) HandleUserSignUp(ctx context.Context, event *saga.Event) error {
	var payload domain.UserSignUpPayload
	if err := event.UnmarshalPayload(&payload); err != nil {
		return err
	}

	userData, _ := json.Marshal(map[string]any{
		"user_id":    payload.UserID,
		"email":      payload.Email,
		"first_name": payload.FirstName,
		"last_name":  payload.LastName,
	})
	event.SetData(userData)

	return event.SendCommand(domain.CmdCreateBucket, "content-service", domain.QueueContentCommands, domain.BucketPayload{
		UserID: payload.UserID,
	})
}

func (w *RegisterUserWorkflow) HandleBucketCreated(ctx context.Context, event *saga.Event) error {
	var payload domain.BucketPayload
	if err := event.UnmarshalPayload(&payload); err != nil {
		return err
	}

	var userData map[string]any
	json.Unmarshal(event.GetData(), &userData)

	return event.SendCommand(domain.CmdSendEmail, "notification-service", domain.QueueNotificationCommands, domain.EmailPayload{
		UserID:    payload.UserID,
		Email:     userData["email"].(string),
		FirstName: userData["first_name"].(string),
		LastName:  userData["last_name"].(string),
	})
}

func (w *RegisterUserWorkflow) HandleEmailSent(ctx context.Context, event *saga.Event) error {
	var payload domain.EmailPayload
	if err := event.UnmarshalPayload(&payload); err != nil {
		return err
	}

	return event.Complete()
}
