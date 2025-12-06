package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/oagudo/outbox"
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
	WithTx(tx pgx.Tx) Repository
}

type TransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

type OutboxRepository interface {
	Save(ctx context.Context, msg *outbox.Message) error
	WithTx(tx pgx.Tx) OutboxRepository
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
	tm              TransactionManager
	outboxRepo      OutboxRepository
	eventHandlers   map[string]EventHandlerFunc
	failureCommands map[string][]string
	commandDest     map[string]CommandDestination
	eventToStep     map[string]string
}

func NewCoordinator(repo Repository, publisher MessagePublisher, tm TransactionManager, outboxRepo OutboxRepository) *Coordinator {
	return &Coordinator{
		repo:            repo,
		publisher:       publisher,
		tm:              tm,
		outboxRepo:      outboxRepo,
		eventHandlers:   make(map[string]EventHandlerFunc),
		failureCommands: make(map[string][]string),
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
		c.failureCommands[step.FailureEvent] = step.OnFailure
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

func (c *Coordinator) On(event string, handler EventHandlerFunc) *Coordinator {
	c.eventHandlers[event] = handler
	return c
}

func (c *Coordinator) OnFailure(event string, compensationCommands ...string) *Coordinator {
	c.failureCommands[event] = compensationCommands
	return c
}

func (c *Coordinator) HandleEvent(ctx context.Context, msg *Message) error {
	state, err := c.GetOrCreateState(ctx, msg)
	if err != nil {
		return fmt.Errorf("failed to get or create saga state: %w", err)
	}

	if state == nil {
		return fmt.Errorf("saga state not found")
	}

	if state.Status == SagaStateCompleted || state.Status == SagaStateCompensated {
		return nil
	}

	// Early return if failure event is received
	if failCmd, ok := c.failureCommands[msg.Type]; ok {
		return c.ExecuteCompensation(ctx, state, failCmd)
	}

	handler, exists := c.eventHandlers[msg.Type]
	if !exists {
		return nil
	}

	event := &Event{
		ctx:         ctx,
		coordinator: c,
		message:     msg,
		sagaState:   state,
	}

	return c.tm.RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := c.UpdateStepStatus(ctx, tx, state, msg); err != nil {
			return err
		}

		event.tx = tx
		if err := handler(ctx, event); err != nil {
			return err
		}

		return c.repo.WithTx(tx).Update(ctx, state.CorrelationID, state.Status, state.Data)
	})
}

func (c *Coordinator) UpdateStepStatus(ctx context.Context, tx pgx.Tx, state *SagaStateEntity, msg *Message) error {
	stepName, ok := c.eventToStep[msg.Type]
	if !ok {
		return nil
	}

	if _, ok := c.failureCommands[msg.Type]; ok {
		var payload map[string]any
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			if errVal, ok := payload["error"]; ok {
				return c.repo.WithTx(tx).UpdateStep(ctx, state.ID, stepName, StepStatusFailed, fmt.Sprintf("%v", errVal))
			}
		}
	}

	return c.repo.WithTx(tx).UpdateStep(ctx, state.ID, stepName, StepStatusCompleted, "")
}

func (c *Coordinator) GetOrCreateState(ctx context.Context, msg *Message) (*SagaStateEntity, error) {
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

func (c *Coordinator) ExecuteCompensation(ctx context.Context, state *SagaStateEntity, commands []string) error {
	return c.tm.RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if err := c.repo.WithTx(tx).Update(ctx, state.CorrelationID, SagaStateCompensating, state.Data); err != nil {
			return fmt.Errorf("failed to update saga state to COMPENSATING: %w", err)
		}

		for _, cmd := range commands {
			dest, ok := c.commandDest[cmd]
			if !ok {
				continue
			}

			if err := c.PublishOutboxCommand(ctx, tx, dest.Queue, state.CorrelationID, cmd, state.Data); err != nil {
				return fmt.Errorf("compensation failed for command %s: %w", cmd, err)
			}
		}

		if err := c.repo.WithTx(tx).Update(ctx, state.CorrelationID, SagaStateCompensated, state.Data); err != nil {
			return fmt.Errorf("failed to update saga state to COMPENSATED: %w", err)
		}

		return nil
	})
}

func (c *Coordinator) PublishOutboxCommand(ctx context.Context, tx pgx.Tx, queueName string, correlationID, cmdType string, payload any) error {
	msg, err := NewSagaMessage(correlationID, cmdType, payload)
	if err != nil {
		return fmt.Errorf("failed to create saga message: %w", err)
	}

	payloadJSON, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	return c.outboxRepo.WithTx(tx).Save(ctx, outbox.NewMessage(payloadJSON,
		outbox.WithID(uuid.New()),
		outbox.WithCreatedAt(time.Now()),
		outbox.WithMetadata([]byte(queueName)),
	))
}

type Event struct {
	ctx         context.Context
	coordinator *Coordinator
	message     *Message
	sagaState   *SagaStateEntity
	tx          pgx.Tx
}

func (e *Event) SendCommand(cmdType string, payload any) error {
	dest, ok := e.coordinator.commandDest[cmdType]
	if !ok {
		return fmt.Errorf("command destination not found for: %s", cmdType)
	}

	if err := e.coordinator.PublishOutboxCommand(e.ctx, e.tx, dest.Queue, e.message.CorrelationID, cmdType, payload); err != nil {
		return err
	}

	return e.coordinator.repo.WithTx(e.tx).AddStep(e.ctx, e.sagaState.ID, cmdType, dest.Service, StepStatusPending)
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
	if err := e.coordinator.repo.WithTx(e.tx).Complete(e.ctx, e.message.CorrelationID); err != nil {
		return fmt.Errorf("failed to complete saga: %w", err)
	}

	e.sagaState.Status = SagaStateCompleted
	return nil
}

func (e *Event) CorrelationID() string {
	return e.message.CorrelationID
}
