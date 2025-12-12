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
			NewOutboxPublisher,
		),
		fx.Invoke(func(rw *RegisterUserWorkflow) {
			rw.Register()
		}),
		fx.Invoke(RunOutboxReader),
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
		RegisterStep(saga.StepDefinition{
			Service:      "content-service",
			Command:      domain.CmdCreateBucket,
			Queue:        domain.QueueContentCommands,
			SuccessEvent: domain.EventBucketCreated,
			Compensations: []string{
				domain.CmdCompensateUser,
			},
		}).
		RegisterStep(saga.StepDefinition{
			Service:      "notification-service",
			Command:      domain.CmdSendEmail,
			Queue:        domain.QueueNotificationCommands,
			SuccessEvent: domain.EventEmailSent,
			Compensations: []string{
				domain.CmdCompensateBucket,
				domain.CmdCompensateUser,
			},
		}).
		RegisterCompensationQueue(domain.CmdCompensateUser, domain.QueueUserCommands).
		RegisterCompensationQueue(domain.CmdCompensateBucket, domain.QueueContentCommands).
		On(domain.EventUserSignUp, w.HandleUserSignUp).
		On(domain.EventBucketCreated, w.HandleBucketCreated).
		On(domain.EventEmailSent, w.HandleEmailSent)
}

func (w *RegisterUserWorkflow) HandleUserSignUp(_ context.Context, event *saga.Event) error {
	var payload domain.UserSignUpPayload
	if err := event.UnmarshalPayload(&payload); err != nil {
		return err
	}

	if err := event.SetState(payload); err != nil {
		return err
	}

	return event.SendCommand(domain.CmdCreateBucket, domain.BucketPayload{
		UserID: payload.UserID,
	})
}

func (w *RegisterUserWorkflow) HandleCompensateUserSignUp(_ context.Context, event *saga.Event) error {
	var state domain.UserSignUpPayload
	if err := event.GetState(&state); err != nil {
		return err
	}

	event.SendCommand(domain.CmdCompensateBucket, domain.CompensateUserSignUpPayload{
		UserID: state.UserID,
	})

	event.SendCommand(domain.CmdCompensateUser, domain.CompensateUserSignUpPayload{
		UserID: state.UserID,
	})

	return nil
}

func (w *RegisterUserWorkflow) HandleBucketCreated(_ context.Context, event *saga.Event) error {
	var payload domain.BucketPayload
	if err := event.UnmarshalPayload(&payload); err != nil {
		return err
	}

	var state domain.UserSignUpPayload
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
