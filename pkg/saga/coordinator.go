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

type SagaRepository interface {
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
type Coordinator struct {
	repo            SagaRepository
	publisher       MessagePublisher
	eventHandlers   map[string]EventHandlerFunc
	failureHandlers map[string][]string
	commandsQueues  map[string]string
}

func NewCoordinator(repo SagaRepository, publisher MessagePublisher) *Coordinator {
	return &Coordinator{
		repo:            repo,
		publisher:       publisher,
		eventHandlers:   make(map[string]EventHandlerFunc),
		failureHandlers: make(map[string][]string),
	}
}

func (c *Coordinator) RegisterCommandQueue(queue map[string]string) *Coordinator {
	c.commandsQueues = queue
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

	if compensationCommands, isFailure := c.failureHandlers[msg.Type]; isFailure {
		return c.executeCompensation(ctx, sagaState, compensationCommands)
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

	if err := handler(ctx, event); err != nil {
		return fmt.Errorf("handler execution failed: %w", err)
	}

	if err := c.repo.Update(ctx, sagaState.CorrelationID, SagaStateCompleted, sagaState.Data); err != nil {
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
		queue, ok := c.commandsQueues[cmd]
		if !ok {
			logrus.WithField("command", cmd).Warn("Command queue not found")
			continue
		}

		if err := c.publishCommand(ctx, queue, sagaState.CorrelationID, cmd, sagaState.Data); err != nil {
			logrus.WithError(err).WithField("command", cmd).Error("Failed to publish compensation command")
			return err
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

	if err := c.publisher.PublishCommand(ctx, queueName, msg); err != nil {
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

func (e *Event) SendCommand(cmdType, serviceName, queueName string, payload any) error {
	if err := e.coordinator.publishCommand(e.ctx, queueName, e.message.CorrelationID, cmdType, payload); err != nil {
		return err
	}

	if err := e.coordinator.repo.AddStep(e.ctx, e.sagaState.ID, cmdType, serviceName, StepStatusPending); err != nil {
		logrus.WithError(err).Warn("Failed to add saga step")
	}

	return nil
}

func (e *Event) SetData(raw json.RawMessage) {
	e.sagaState.Data = raw
}

func (e *Event) GetData() json.RawMessage {
	return e.sagaState.Data
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
