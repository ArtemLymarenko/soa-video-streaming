package saga

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewCoordinator,
		),
	)
}

type EventHandlerFunc func(ctx context.Context, event *Event) error

type Repository interface {
	Create(ctx context.Context, correlationID string, status SagaStateStatus, data json.RawMessage) (*SagaStateEntity, error)
	Update(ctx context.Context, correlationID string, status SagaStateStatus, data json.RawMessage) error
	Complete(ctx context.Context, correlationID string) error
	FindByCorrelationID(ctx context.Context, correlationID string) (*SagaStateEntity, error)
	AddStep(ctx context.Context, sagaStateID, stepName, serviceName string, status StepStatus) error
	UpdateStep(ctx context.Context, sagaStateID, stepName string, status StepStatus, errorMessage string) error
	GetSteps(ctx context.Context, sagaStateID string) ([]SagaStep, error)
}

type MessagePublisher interface {
	PublishCommand(ctx context.Context, queue string, msg *Message) error
}

type CommandDestination struct {
	Queue   string
	Service string
}

type StepDefinition struct {
	Name         string
	Command      string
	Queue        string
	Service      string
	SuccessEvent string
	FailureEvent string
	OnFailure    []string
}

type Coordinator struct {
	repo            Repository
	publisher       MessagePublisher
	eventHandlers   map[string]EventHandlerFunc
	failureHandlers map[string][]string
	commandDest     map[string]CommandDestination
	eventToStep     map[string]string
}

func NewCoordinator(repo Repository, publisher MessagePublisher) *Coordinator {
	return &Coordinator{
		repo:            repo,
		publisher:       publisher,
		eventHandlers:   make(map[string]EventHandlerFunc),
		failureHandlers: make(map[string][]string),
		commandDest:     make(map[string]CommandDestination),
		eventToStep:     make(map[string]string),
	}
}

func (c *Coordinator) RegisterStep(step StepDefinition) *Coordinator {
	c.commandDest[step.Command] = CommandDestination{
		Queue:   step.Queue,
		Service: step.Service,
	}

	if step.SuccessEvent != "" {
		c.eventToStep[step.SuccessEvent] = step.Command
	}
	if step.FailureEvent != "" {
		c.eventToStep[step.FailureEvent] = step.Command
	}

	if step.FailureEvent != "" && len(step.OnFailure) > 0 {
		c.failureHandlers[step.FailureEvent] = step.OnFailure
	}

	return c
}

func (c *Coordinator) RegisterCompensationCommand(command, queue, service string) *Coordinator {
	c.commandDest[command] = CommandDestination{
		Queue:   queue,
		Service: service,
	}
	return c
}

func (c *Coordinator) On(eventType string, handler EventHandlerFunc) *Coordinator {
	c.eventHandlers[eventType] = handler
	return c
}

func (c *Coordinator) OnFailure(failureEvent string, compensationCommands ...string) *Coordinator {
	c.failureHandlers[failureEvent] = compensationCommands
	return c
}

func (c *Coordinator) HandleEvent(ctx context.Context, msg *Message) error {
	sagaState, err := c.getOrCreateSagaState(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to get/create saga state: %w", err)
	}

	if sagaState == nil {
		return fmt.Errorf("saga state not found")
	}

	if sagaState.Status == SagaStateCompleted || sagaState.Status == SagaStateCompensated {
		return nil
	}

	if failHandler, ok := c.failureHandlers[msg.Type]; ok {
		return c.executeCompensation(ctx, sagaState, failHandler)
	}

	handler, exists := c.eventHandlers[msg.Type]
	if !exists {
		return nil
	}

	event := &Event{
		ctx:         ctx,
		coordinator: c,
		message:     msg,
		sagaState:   sagaState,
	}

	if err = c.updateStepStatusFromEvent(ctx, sagaState, msg); err != nil {
		logrus.WithError(err).Warn("Failed to update step status")
	}

	if err = handler(ctx, event); err != nil {
		return fmt.Errorf("handler execution failed: %w", err)
	}

	if err = c.repo.Update(ctx, sagaState.CorrelationID, sagaState.Status, sagaState.Data); err != nil {
		return fmt.Errorf("failed to update saga state: %w", err)
	}

	return nil
}

func (c *Coordinator) getOrCreateSagaState(ctx context.Context, msg *Message) (*SagaStateEntity, error) {
	sagaState, err := c.repo.FindByCorrelationID(ctx, msg.CorrelationID)
	if err != nil {
		return nil, err
	}

	if sagaState != nil {
		return sagaState, nil
	}

	sagaState, err = c.repo.Create(ctx, msg.CorrelationID, SagaStateStarted, nil)
	if err != nil {
		return nil, err
	}

	return sagaState, nil
}

func (c *Coordinator) executeCompensation(ctx context.Context, sagaState *SagaStateEntity, commands []string) error {
	if err := c.repo.Update(ctx, sagaState.CorrelationID, SagaStateCompensating, sagaState.Data); err != nil {
		return fmt.Errorf("failed to update saga state to COMPENSATING: %w", err)
	}

	for _, cmd := range commands {
		dest, ok := c.commandDest[cmd]
		if !ok {
			continue
		}

		if err := c.publishCommand(ctx, dest.Queue, sagaState.CorrelationID, cmd, sagaState.Data); err != nil {
			return fmt.Errorf("compensation failed for command %s: %w", cmd, err)
		}
	}

	if err := c.repo.Update(ctx, sagaState.CorrelationID, SagaStateCompensated, sagaState.Data); err != nil {
		return fmt.Errorf("failed to update saga state to COMPENSATED: %w", err)
	}

	return nil
}

func (c *Coordinator) publishCommand(ctx context.Context, queueName string, correlationID, cmdType string, payload any) error {
	msg, err := NewSagaMessage(correlationID, cmdType, payload)
	if err != nil {
		return fmt.Errorf("failed to create saga message: %w", err)
	}

	if err = c.publisher.PublishCommand(ctx, queueName, msg); err != nil {
		return fmt.Errorf("failed to publish command: %w", err)
	}

	return nil
}

type Event struct {
	ctx         context.Context
	coordinator *Coordinator
	message     *Message
	sagaState   *SagaStateEntity
}

func (e *Event) SendCommand(cmdType string, payload any) error {
	dest, ok := e.coordinator.commandDest[cmdType]
	if !ok {
		return fmt.Errorf("command destination not found for: %s", cmdType)
	}

	if err := e.coordinator.publishCommand(e.ctx, dest.Queue, e.message.CorrelationID, cmdType, payload); err != nil {
		return err
	}

	if err := e.coordinator.repo.AddStep(e.ctx, e.sagaState.ID, cmdType, dest.Service, StepStatusPending); err != nil {
		logrus.WithError(err).Warn("Failed to add saga step")
	}

	return nil
}

func (e *Event) SetState(state any) error {
	raw, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}
	e.sagaState.Data = raw
	return nil
}

func (e *Event) GetState(target any) error {
	if len(e.sagaState.Data) == 0 {
		return nil
	}
	if err := json.Unmarshal(e.sagaState.Data, target); err != nil {
		return fmt.Errorf("unmarshal state: %w", err)
	}
	return nil
}

func (e *Event) UnmarshalPayload(target any) error {
	if err := json.Unmarshal(e.message.Payload, target); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return nil
}

func (e *Event) Complete() error {
	if err := e.coordinator.repo.Complete(e.ctx, e.message.CorrelationID); err != nil {
		return fmt.Errorf("failed to complete saga: %w", err)
	}

	e.sagaState.Status = SagaStateCompleted
	return nil
}

func (e *Event) CorrelationID() string {
	return e.message.CorrelationID
}

func (c *Coordinator) updateStepStatusFromEvent(ctx context.Context, sagaState *SagaStateEntity, msg *Message) error {
	stepName, ok := c.eventToStep[msg.Type]
	if !ok {
		return nil
	}

	status := StepStatusCompleted
	errorMsg := ""

	if _, ok := c.failureHandlers[msg.Type]; ok {
		status = StepStatusFailed

		var payload map[string]any
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			if errVal, ok := payload["error"]; ok {
				errorMsg = fmt.Sprintf("%v", errVal)
			}
		}
	}

	return c.repo.UpdateStep(ctx, sagaState.ID, stepName, status, errorMsg)
}
