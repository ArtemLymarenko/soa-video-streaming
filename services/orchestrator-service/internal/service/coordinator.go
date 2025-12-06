package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"

	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/orchestrator-service/internal/repository/postgres"
)

type Coordinator struct {
	sagaRepo  *postgres.SagaRepository
	publisher *gorabbit.Publisher
}

func NewCoordinator(sagaRepo *postgres.SagaRepository, client *rabbitmq.Client) (*Coordinator, error) {
	publisher, err := gorabbit.NewPublisher(
		client.Conn,
		gorabbit.WithPublisherOptionsLogger(logrus.StandardLogger()),
	)
	if err != nil {
		return nil, err
	}

	return &Coordinator{
		sagaRepo:  sagaRepo,
		publisher: publisher,
	}, nil
}

// HandleUserSignUp handles the EventUserSignUp event
func (c *Coordinator) HandleUserSignUp(ctx context.Context, msg *saga.SagaMessage) error {
	logrus.WithField("correlation_id", msg.CorrelationID).Info("📝 Received EventUserSignUp")

	var payload saga.UserSignUpPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Create saga state
	sagaData := map[string]interface{}{
		"user_id":    payload.UserID,
		"email":      payload.Email,
		"first_name": payload.FirstName,
		"last_name":  payload.LastName,
	}

	sagaState, err := c.sagaRepo.Create(ctx, msg.CorrelationID, saga.SagaStateStarted, sagaData)
	if err != nil {
		return fmt.Errorf("create saga state: %w", err)
	}

	// Add initial step
	if err := c.sagaRepo.AddStep(ctx, sagaState.ID, "user_signup", "user-service", saga.StepStatusCompleted); err != nil {
		logrus.WithError(err).Error("Failed to add step")
	}

	// Send CmdCreateBucket to Content Service
	logrus.WithField("correlation_id", msg.CorrelationID).Info("➡️  Sending CmdCreateBucket to Content Service")

	bucketPayload := saga.BucketPayload{
		UserID: payload.UserID,
	}

	cmdMsg, err := saga.NewSagaMessage(msg.CorrelationID, saga.CmdCreateBucket, bucketPayload)
	if err != nil {
		return fmt.Errorf("create command message: %w", err)
	}

	// Add step for bucket creation
	if err := c.sagaRepo.AddStep(ctx, sagaState.ID, "create_bucket", "content-service", saga.StepStatusPending); err != nil {
		logrus.WithError(err).Error("Failed to add step")
	}

	return c.publishMessage(ctx, saga.QueueContentCommands, cmdMsg)
}

// HandleBucketCreated handles the EventBucketCreated event
func (c *Coordinator) HandleBucketCreated(ctx context.Context, msg *saga.SagaMessage) error {
	logrus.WithField("correlation_id", msg.CorrelationID).Info("✅ Received EventBucketCreated")

	var payload saga.BucketPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Get saga state
	sagaState, err := c.sagaRepo.FindByCorrelationID(ctx, msg.CorrelationID)
	if err != nil {
		return fmt.Errorf("find saga state: %w", err)
	}
	if sagaState == nil {
		return fmt.Errorf("saga state not found for correlation_id: %s", msg.CorrelationID)
	}

	// Update step
	if err := c.sagaRepo.UpdateStep(ctx, sagaState.ID, "create_bucket", saga.StepStatusCompleted, ""); err != nil {
		logrus.WithError(err).Error("Failed to update step")
	}

	// Update saga data with bucket name
	sagaState.Data["bucket_name"] = payload.BucketName
	if err := c.sagaRepo.Update(ctx, msg.CorrelationID, saga.SagaStateStarted, sagaState.Data); err != nil {
		logrus.WithError(err).Error("Failed to update saga state")
	}

	// Send CmdSendEmail to Notification Service
	logrus.WithField("correlation_id", msg.CorrelationID).Info("➡️  Sending CmdSendEmail to Notification Service")

	emailPayload := saga.EmailPayload{
		UserID:    payload.UserID,
		Email:     sagaState.Data["email"].(string),
		FirstName: sagaState.Data["first_name"].(string),
		LastName:  sagaState.Data["last_name"].(string),
	}

	cmdMsg, err := saga.NewSagaMessage(msg.CorrelationID, saga.CmdSendEmail, emailPayload)
	if err != nil {
		return fmt.Errorf("create command message: %w", err)
	}

	// Add step for email sending
	if err := c.sagaRepo.AddStep(ctx, sagaState.ID, "send_email", "notification-service", saga.StepStatusPending); err != nil {
		logrus.WithError(err).Error("Failed to add step")
	}

	return c.publishMessage(ctx, saga.QueueNotificationCommands, cmdMsg)
}

// HandleEmailSent handles the EventEmailSent event
func (c *Coordinator) HandleEmailSent(ctx context.Context, msg *saga.SagaMessage) error {
	logrus.WithField("correlation_id", msg.CorrelationID).Info("✅ Received EventEmailSent")

	var payload saga.EmailPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Get saga state
	sagaState, err := c.sagaRepo.FindByCorrelationID(ctx, msg.CorrelationID)
	if err != nil {
		return fmt.Errorf("find saga state: %w", err)
	}
	if sagaState == nil {
		return fmt.Errorf("saga state not found for correlation_id: %s", msg.CorrelationID)
	}

	// Update step
	if err := c.sagaRepo.UpdateStep(ctx, sagaState.ID, "send_email", saga.StepStatusCompleted, ""); err != nil {
		logrus.WithError(err).Error("Failed to update step")
	}

	// Send CmdActivateUser to User Service
	logrus.WithField("correlation_id", msg.CorrelationID).Info("➡️  Sending CmdActivateUser to User Service")

	userPayload := saga.UserPayload{
		UserID: payload.UserID,
	}

	cmdMsg, err := saga.NewSagaMessage(msg.CorrelationID, saga.CmdActivateUser, userPayload)
	if err != nil {
		return fmt.Errorf("create command message: %w", err)
	}

	// Add step for user activation
	if err := c.sagaRepo.AddStep(ctx, sagaState.ID, "activate_user", "user-service", saga.StepStatusPending); err != nil {
		logrus.WithError(err).Error("Failed to add step")
	}

	return c.publishMessage(ctx, saga.QueueUserCommands, cmdMsg)
}

// HandleUserActivated handles the user activation completion (can be called after activation)
func (c *Coordinator) HandleUserActivated(ctx context.Context, correlationID string) error {
	logrus.WithField("correlation_id", correlationID).Info("🎉 User activated, completing saga")

	// Get saga state
	sagaState, err := c.sagaRepo.FindByCorrelationID(ctx, correlationID)
	if err != nil {
		return fmt.Errorf("find saga state: %w", err)
	}
	if sagaState == nil {
		return fmt.Errorf("saga state not found for correlation_id: %s", correlationID)
	}

	// Update step
	if err := c.sagaRepo.UpdateStep(ctx, sagaState.ID, "activate_user", saga.StepStatusCompleted, ""); err != nil {
		logrus.WithError(err).Error("Failed to update step")
	}

	// Complete saga
	if err := c.sagaRepo.Complete(ctx, correlationID); err != nil {
		return fmt.Errorf("complete saga: %w", err)
	}

	logrus.WithField("correlation_id", correlationID).Info("✨ Saga completed successfully")
	return nil
}

// HandleBucketFailed handles the EventBucketFailed event
func (c *Coordinator) HandleBucketFailed(ctx context.Context, msg *saga.SagaMessage) error {
	logrus.WithField("correlation_id", msg.CorrelationID).Error("❌ Received EventBucketFailed, starting compensation")

	var payload saga.BucketPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Get saga state
	sagaState, err := c.sagaRepo.FindByCorrelationID(ctx, msg.CorrelationID)
	if err != nil {
		return fmt.Errorf("find saga state: %w", err)
	}
	if sagaState == nil {
		return fmt.Errorf("saga state not found for correlation_id: %s", msg.CorrelationID)
	}

	// Update saga state to compensating
	if err := c.sagaRepo.Update(ctx, msg.CorrelationID, saga.SagaStateCompensating, sagaState.Data); err != nil {
		logrus.WithError(err).Error("Failed to update saga state")
	}

	// Update step with error
	if err := c.sagaRepo.UpdateStep(ctx, sagaState.ID, "create_bucket", saga.StepStatusFailed, payload.Error); err != nil {
		logrus.WithError(err).Error("Failed to update step")
	}

	// Compensate user (delete user)
	return c.compensateUser(ctx, msg.CorrelationID, sagaState.Data["user_id"].(string))
}

// HandleEmailFailed handles the EventEmailFailed event
func (c *Coordinator) HandleEmailFailed(ctx context.Context, msg *saga.SagaMessage) error {
	logrus.WithField("correlation_id", msg.CorrelationID).Error("❌ Received EventEmailFailed, starting compensation")

	var payload saga.EmailPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// Get saga state
	sagaState, err := c.sagaRepo.FindByCorrelationID(ctx, msg.CorrelationID)
	if err != nil {
		return fmt.Errorf("find saga state: %w", err)
	}
	if sagaState == nil {
		return fmt.Errorf("saga state not found for correlation_id: %s", msg.CorrelationID)
	}

	// Update saga state to compensating
	if err := c.sagaRepo.Update(ctx, msg.CorrelationID, saga.SagaStateCompensating, sagaState.Data); err != nil {
		logrus.WithError(err).Error("Failed to update saga state")
	}

	// Update step with error
	if err := c.sagaRepo.UpdateStep(ctx, sagaState.ID, "send_email", saga.StepStatusFailed, payload.Error); err != nil {
		logrus.WithError(err).Error("Failed to update step")
	}

	// Compensate bucket and user
	if err := c.compensateBucket(ctx, msg.CorrelationID, payload.UserID); err != nil {
		logrus.WithError(err).Error("Failed to compensate bucket")
	}

	return c.compensateUser(ctx, msg.CorrelationID, payload.UserID)
}

// compensateBucket sends CmdCompensateBucket
func (c *Coordinator) compensateBucket(ctx context.Context, correlationID, userID string) error {
	logrus.WithField("correlation_id", correlationID).Info("🔄 Compensating bucket")

	bucketPayload := saga.BucketPayload{
		UserID: userID,
	}

	cmdMsg, err := saga.NewSagaMessage(correlationID, saga.CmdCompensateBucket, bucketPayload)
	if err != nil {
		return fmt.Errorf("create command message: %w", err)
	}

	return c.publishMessage(ctx, saga.QueueContentCommands, cmdMsg)
}

// compensateUser sends CmdCompensateUser
func (c *Coordinator) compensateUser(ctx context.Context, correlationID, userID string) error {
	logrus.WithField("correlation_id", correlationID).Info("🔄 Compensating user")

	userPayload := saga.UserPayload{
		UserID: userID,
	}

	cmdMsg, err := saga.NewSagaMessage(correlationID, saga.CmdCompensateUser, userPayload)
	if err != nil {
		return fmt.Errorf("create command message: %w", err)
	}

	return c.publishMessage(ctx, saga.QueueUserCommands, cmdMsg)
}

// publishMessage publishes a message to RabbitMQ
func (c *Coordinator) publishMessage(ctx context.Context, queue string, msg *saga.SagaMessage) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	return c.publisher.PublishWithContext(
		ctx,
		msgBytes,
		[]string{queue},
		gorabbit.WithPublishOptionsContentType("application/json"),
	)
}
