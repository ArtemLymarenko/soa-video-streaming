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
		fx.Provide(NewCoordinator),
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
	WithTx(tx pgx.Tx) Repository
}

type TransactionManager interface {
	RunInTransaction(ctx context.Context, fn func(ctx context.Context, tx pgx.Tx) error) error
}

type OutboxRepository interface {
	Save(ctx context.Context, msg *outbox.Message) error
	WithTx(tx pgx.Tx) OutboxRepository
}

type CommandDestination struct {
	Queue   string
	Service string
}

type StepDefinition struct {
	Command       string
	Queue         string
	Service       string
	SuccessEvent  string
	Compensations []string
}

type Coordinator struct {
	repo           Repository
	tm             TransactionManager
	outboxRepo     OutboxRepository
	eventHandlers  map[string]EventHandlerFunc
	commandDest    map[string]CommandDestination
	eventToCommand map[string]string
	compensations  map[string][]string
}

func NewCoordinator(repo Repository, tm TransactionManager, outboxRepo OutboxRepository) *Coordinator {
	return &Coordinator{
		repo:           repo,
		tm:             tm,
		outboxRepo:     outboxRepo,
		eventHandlers:  make(map[string]EventHandlerFunc),
		commandDest:    make(map[string]CommandDestination),
		eventToCommand: make(map[string]string),
		compensations:  make(map[string][]string),
	}
}

func (c *Coordinator) RegisterStep(step StepDefinition) *Coordinator {
	c.commandDest[step.Command] = CommandDestination{
		Queue:   step.Queue,
		Service: step.Service,
	}

	if step.SuccessEvent != "" {
		c.eventToCommand[step.SuccessEvent] = step.Command
	}

	if len(step.Compensations) > 0 {
		c.compensations[step.Command] = step.Compensations
	}

	return c
}

func (c *Coordinator) RegisterCompensationQueue(cmd string, queue string) *Coordinator {
	c.commandDest[cmd] = CommandDestination{
		Queue: queue,
	}

	return c
}

func (c *Coordinator) On(event string, handler EventHandlerFunc) *Coordinator {
	c.eventHandlers[event] = handler
	return c
}

func (c *Coordinator) HandleEvent(ctx context.Context, msg *Message) error {
	state, err := c.GetOrCreateState(ctx, msg)
	if err != nil {
		return fmt.Errorf("get/create state: %w", err)
	}

	// Ігноруємо, якщо сага вже завершена або компенсована
	if state.Status == SagaStateCompleted || state.Status == SagaStateCompensated {
		return nil
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
		if cmd, ok := c.eventToCommand[msg.Type]; ok {
			if err := c.repo.WithTx(tx).UpdateStep(ctx, state.ID, cmd, StepStatusCompleted, ""); err != nil {
				return err
			}
		}

		event.tx = tx

		if err := handler(ctx, event); err != nil {
			return err
		}

		return c.repo.WithTx(tx).Update(ctx, state.CorrelationID, state.Status, state.Data)
	})
}

func (c *Coordinator) HandleFailure(ctx context.Context, msg *Message) error {
	state, err := c.repo.FindByCorrelationID(ctx, msg.CorrelationID)
	if err != nil || state == nil {
		return fmt.Errorf("state not found for failure handling: %w", err)
	}

	if state.Status == SagaStateCompleted || state.Status == SagaStateCompensated {
		return nil
	}

	comps, ok := c.compensations[msg.Type]
	if !ok {
		return fmt.Errorf("no compensation defined for failed command: %s", msg.Type)
	}

	return c.tm.RunInTransaction(ctx, func(ctx context.Context, tx pgx.Tx) error {
		errMsg := "Command failed and moved to DLQ"
		if err := c.repo.WithTx(tx).UpdateStep(ctx, state.ID, msg.Type, StepStatusFailed, errMsg); err != nil {
			return err
		}

		if err := c.repo.WithTx(tx).Update(ctx, state.CorrelationID, SagaStateCompensating, state.Data); err != nil {
			return err
		}

		for _, cmd := range comps {
			dest, ok := c.commandDest[cmd]
			if !ok {
				continue
			}

			if err := c.publishOutboxCommand(ctx, tx, dest.Queue, state.CorrelationID, cmd, state.Data); err != nil {
				return err
			}
		}

		return c.repo.WithTx(tx).Update(ctx, state.CorrelationID, SagaStateCompensated, state.Data)
	})
}

func (c *Coordinator) GetOrCreateState(ctx context.Context, msg *Message) (*SagaStateEntity, error) {
	state, err := c.repo.FindByCorrelationID(ctx, msg.CorrelationID)
	if err != nil {
		return nil, err
	}
	if state != nil {
		return state, nil
	}
	return c.repo.Create(ctx, msg.CorrelationID, SagaStateStarted, nil)
}

func (c *Coordinator) publishOutboxCommand(ctx context.Context, tx pgx.Tx, queueName, correlationID, cmdType string, payload any) error {
	msg, err := NewSagaMessage(correlationID, cmdType, payload)
	if err != nil {
		return err
	}
	payloadJSON, _ := json.Marshal(msg)

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
		return fmt.Errorf("destination not found for: %s", cmdType)
	}
	if err := e.coordinator.publishOutboxCommand(e.ctx, e.tx, dest.Queue, e.message.CorrelationID, cmdType, payload); err != nil {
		return err
	}
	return e.coordinator.repo.WithTx(e.tx).AddStep(e.ctx, e.sagaState.ID, cmdType, dest.Service, StepStatusPending)
}

func (e *Event) Complete() error {
	if err := e.coordinator.repo.WithTx(e.tx).Complete(e.ctx, e.message.CorrelationID); err != nil {
		return err
	}

	e.sagaState.Status = SagaStateCompleted
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

func (e *Event) CorrelationID() string {
	return e.message.CorrelationID
}
