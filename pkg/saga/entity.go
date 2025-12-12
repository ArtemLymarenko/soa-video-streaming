package saga

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SagaStateStatus string

const (
	SagaStateStarted      SagaStateStatus = "STARTED"
	SagaStateCompleted    SagaStateStatus = "COMPLETED"
	SagaStateCompensating SagaStateStatus = "COMPENSATING"
	SagaStateCompensated  SagaStateStatus = "COMPENSATED"
)

type StepStatus string

const (
	StepStatusPending   StepStatus = "PENDING"
	StepStatusCompleted StepStatus = "COMPLETED"
	StepStatusFailed    StepStatus = "FAILED"
)

type SagaStateEntity struct {
	ID            string
	CorrelationID string
	Status        SagaStateStatus
	Data          json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CompletedAt   *time.Time
}

type SagaStep struct {
	ID            string
	SagaStateID   string
	StepName      string
	ServiceName   string
	Status        StepStatus
	ExecutedAt    *time.Time
	CompensatedAt *time.Time
	ErrorMessage  string
	CreatedAt     time.Time
}

type Message struct {
	CorrelationID string          `json:"correlation_id"`
	Type          string          `json:"type"`
	Payload       json.RawMessage `json:"payload"`
	Timestamp     time.Time       `json:"timestamp"`
}

type MessageConfig struct {
	WithAutoCorrelationID bool
}

func WithAutoCorrelationID() func(m *MessageConfig) {
	return func(m *MessageConfig) {
		m.WithAutoCorrelationID = true
	}
}

func NewSagaMessage(correlationID, msgType string, payload any, opts ...func(m *MessageConfig)) (*Message, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	config := &MessageConfig{}
	for _, opt := range opts {
		opt(config)
	}

	if config.WithAutoCorrelationID {
		correlationID = uuid.NewString()
	}

	return &Message{
		CorrelationID: correlationID,
		Type:          msgType,
		Payload:       payloadBytes,
		Timestamp:     time.Now(),
	}, nil
}
