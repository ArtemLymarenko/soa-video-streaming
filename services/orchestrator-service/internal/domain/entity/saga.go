package entity

import (
	"time"

	"soa-video-streaming/pkg/saga"
)

// SagaState represents the state of a saga execution
type SagaState struct {
	ID            string
	CorrelationID string
	State         saga.SagaState
	Data          map[string]interface{}
	CreatedAt     time.Time
	UpdatedAt     time.Time
	CompletedAt   *time.Time
}

// SagaStep represents a single step in a saga
type SagaStep struct {
	ID            string
	SagaStateID   string
	StepName      string
	ServiceName   string
	Status        saga.StepStatus
	ExecutedAt    *time.Time
	CompensatedAt *time.Time
	ErrorMessage  string
	CreatedAt     time.Time
}
