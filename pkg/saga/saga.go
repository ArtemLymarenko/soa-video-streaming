package saga

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// SagaState represents the state of a saga
type SagaState string

const (
	SagaStateStarted      SagaState = "STARTED"
	SagaStateCompleted    SagaState = "COMPLETED"
	SagaStateFailed       SagaState = "FAILED"
	SagaStateCompensating SagaState = "COMPENSATING"
	SagaStateCompensated  SagaState = "COMPENSATED"
)

// StepStatus represents the status of a saga step
type StepStatus string

const (
	StepStatusPending      StepStatus = "PENDING"
	StepStatusCompleted    StepStatus = "COMPLETED"
	StepStatusFailed       StepStatus = "FAILED"
	StepStatusCompensating StepStatus = "COMPENSATING"
	StepStatusCompensated  StepStatus = "COMPENSATED"
)

// Event types
const (
	EventUserSignUp    = "saga.user.signup"
	EventBucketCreated = "saga.bucket.created"
	EventBucketFailed  = "saga.bucket.failed"
	EventEmailSent     = "saga.email.sent"
	EventEmailFailed   = "saga.email.failed"
)

// Command types
const (
	CmdCreateBucket     = "saga.cmd.create_bucket"
	CmdCompensateBucket = "saga.cmd.compensate_bucket"
	CmdSendEmail        = "saga.cmd.send_email"
	CmdActivateUser     = "saga.cmd.activate_user"
	CmdCompensateUser   = "saga.cmd.compensate_user"
)

// Queue names for RabbitMQ
const (
	QueueUserSignUp           = "saga.user.signup"
	QueueBucketEvents         = "saga.bucket.events"
	QueueEmailEvents          = "saga.email.events"
	QueueContentCommands      = "saga.content.commands"
	QueueUserCommands         = "saga.user.commands"
	QueueNotificationCommands = "saga.notification.commands"
)

// SagaMessage represents a message in the saga
type SagaMessage struct {
	CorrelationID string          `json:"correlation_id"`
	Type          string          `json:"type"`
	Payload       json.RawMessage `json:"payload"`
	ReplyTo       string          `json:"reply_to,omitempty"`
	Timestamp     time.Time       `json:"timestamp"`
}

// NewSagaMessage creates a new saga message with a given correlation ID
func NewSagaMessage(correlationID, msgType string, payload interface{}) (*SagaMessage, error) {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &SagaMessage{
		CorrelationID: correlationID,
		Type:          msgType,
		Payload:       payloadBytes,
		Timestamp:     time.Now(),
	}, nil
}

// NewSagaMessageWithNewCorrelation creates a new saga message with a new correlation ID
func NewSagaMessageWithNewCorrelation(msgType string, payload interface{}) (*SagaMessage, error) {
	return NewSagaMessage(uuid.NewString(), msgType, payload)
}

// UserSignUpPayload is the payload for EventUserSignUp
type UserSignUpPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// BucketPayload is the payload for bucket-related events/commands
type BucketPayload struct {
	UserID     string `json:"user_id"`
	BucketName string `json:"bucket_name,omitempty"`
	Error      string `json:"error,omitempty"`
}

// EmailPayload is the payload for email-related events/commands
type EmailPayload struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Error     string `json:"error,omitempty"`
}

// UserPayload is the payload for user-related commands
type UserPayload struct {
	UserID string `json:"user_id"`
}
