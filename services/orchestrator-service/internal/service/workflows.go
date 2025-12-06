package service

import (
	"context"

	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/orchestrator-service/domain"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewRegisterUserWorkflow,
		),
		fx.Invoke(func(rw *RegisterUserWorkflow) {
			rw.Register()
		}),
	)
}

type RegisterUserWorkflow struct {
	coordinator *saga.Coordinator
}

type RegisterUserSagaState struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

func NewRegisterUserWorkflow(coordinator *saga.Coordinator) *RegisterUserWorkflow {
	return &RegisterUserWorkflow{
		coordinator: coordinator,
	}
}

func (w *RegisterUserWorkflow) Register() {
	w.coordinator.
		// Крок 1: Створити bucket для користувача
		RegisterStep(saga.StepDefinition{
			Name:         "CreateBucket",
			Command:      domain.CmdCreateBucket,
			Queue:        domain.QueueContentCommands,
			Service:      "content-service",
			SuccessEvent: domain.EventBucketCreated,
			FailureEvent: domain.EventBucketFailed,
			OnFailure:    []string{domain.CmdCompensateUser}, // якщо bucket не створився → видалити користувача
		}).
		// Крок 2: Відправити email користувачу
		RegisterStep(saga.StepDefinition{
			Name:         "SendEmail",
			Command:      domain.CmdSendEmail,
			Queue:        domain.QueueNotificationCommands,
			Service:      "notification-service",
			SuccessEvent: domain.EventEmailSent,
			FailureEvent: domain.EventEmailFailed,
			OnFailure:    []string{domain.CmdCompensateBucket, domain.CmdCompensateUser}, // якщо email не відправився → видалити bucket та користувача
		}).
		// Явно вказуємо destinations для компенсаційних команд
		RegisterCompensationCommand(domain.CmdCompensateUser, domain.QueueUserCommands, "user-service").
		RegisterCompensationCommand(domain.CmdCompensateBucket, domain.QueueContentCommands, "content-service").
		// Обробники подій
		On(domain.EventUserSignUp, w.HandleUserSignUp).
		On(domain.EventBucketCreated, w.HandleBucketCreated).
		On(domain.EventEmailSent, w.HandleEmailSent)
}

func (w *RegisterUserWorkflow) HandleUserSignUp(_ context.Context, event *saga.Event) error {
	var payload domain.UserSignUpPayload
	if err := event.UnmarshalPayload(&payload); err != nil {
		return err
	}

	state := RegisterUserSagaState{
		UserID:    payload.UserID,
		Email:     payload.Email,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
	}
	if err := event.SetState(state); err != nil {
		return err
	}

	return event.SendCommand(domain.CmdCreateBucket, domain.BucketPayload{
		UserID: payload.UserID,
	})
}

func (w *RegisterUserWorkflow) HandleBucketCreated(_ context.Context, event *saga.Event) error {
	var payload domain.BucketPayload
	if err := event.UnmarshalPayload(&payload); err != nil {
		return err
	}

	var state RegisterUserSagaState
	if err := event.GetState(&state); err != nil {
		return err
	}

	return event.SendCommand(domain.CmdSendEmail, domain.EmailPayload{
		UserID:    payload.UserID,
		Email:     state.Email,
		FirstName: state.FirstName,
		LastName:  state.LastName,
	})
}

func (w *RegisterUserWorkflow) HandleEmailSent(_ context.Context, event *saga.Event) error {
	return event.Complete()
}
