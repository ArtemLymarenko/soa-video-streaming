package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	gorabbit "github.com/wagslane/go-rabbitmq"

	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/content-service/internal/domain/entity"
	"soa-video-streaming/services/content-service/internal/repository/postgres"
)

type BucketHandler struct {
	storageRepo *postgres.StorageRepository
	s3Mock      *S3Mock
	publisher   *gorabbit.Publisher
}

func NewBucketHandler(
	storageRepo *postgres.StorageRepository,
	s3Mock *S3Mock,
	client *rabbitmq.Client,
) (*BucketHandler, error) {
	publisher, err := gorabbit.NewPublisher(
		client.Conn,
		gorabbit.WithPublisherOptionsLogger(logrus.StandardLogger()),
	)
	if err != nil {
		return nil, err
	}

	return &BucketHandler{
		storageRepo: storageRepo,
		s3Mock:      s3Mock,
		publisher:   publisher,
	}, nil
}

func (h *BucketHandler) HandleCreateBucket(ctx context.Context, msg *saga.SagaMessage) error {
	logrus.WithField("correlation_id", msg.CorrelationID).Info("Received CmdCreateBucket")

	var payload saga.BucketPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	bucketName, err := h.s3Mock.CreateBucket(payload.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to create bucket")
		return h.publishBucketFailed(ctx, msg.CorrelationID, payload.UserID, err.Error())
	}

	account := entity.StorageAccount{
		UserID:     payload.UserID,
		BucketName: bucketName,
		Status:     entity.StorageAccountStatusActive,
		CreatedAt:  time.Now(),
	}

	if err := h.storageRepo.Create(ctx, account); err != nil {
		logrus.WithError(err).Error("Failed to save storage account")

		_ = h.s3Mock.DeleteBucket(bucketName)

		return h.publishBucketFailed(ctx, msg.CorrelationID, payload.UserID, err.Error())
	}

	logrus.WithFields(logrus.Fields{
		"correlation_id": msg.CorrelationID,
		"user_id":        payload.UserID,
		"bucket_name":    bucketName,
	}).Info("Bucket created successfully")

	return h.publishBucketCreated(ctx, msg.CorrelationID, payload.UserID, bucketName)
}

func (h *BucketHandler) HandleCompensateBucket(ctx context.Context, msg *saga.SagaMessage) error {
	logrus.WithField("correlation_id", msg.CorrelationID).Info("Received CmdCompensateBucket")

	var payload saga.BucketPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	account, err := h.storageRepo.FindByUserID(ctx, payload.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to find storage account")
		return err
	}

	if account == nil {
		logrus.WithField("user_id", payload.UserID).Warn("Storage account not found for compensation")
		return nil
	}

	if err := h.s3Mock.DeleteBucket(account.BucketName); err != nil {
		logrus.WithError(err).Error("Failed to delete bucket from S3")
	}

	if err := h.storageRepo.Delete(ctx, payload.UserID); err != nil {
		logrus.WithError(err).Error("Failed to delete storage account from database")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"correlation_id": msg.CorrelationID,
		"user_id":        payload.UserID,
		"bucket_name":    account.BucketName,
	}).Info("Bucket compensated (deleted) successfully")

	return nil
}
func (h *BucketHandler) publishBucketCreated(ctx context.Context, correlationID, userID, bucketName string) error {
	payload := saga.BucketPayload{
		UserID:     userID,
		BucketName: bucketName,
	}

	msg, err := saga.NewSagaMessage(correlationID, saga.EventBucketCreated, payload)
	if err != nil {
		return fmt.Errorf("create saga message: %w", err)
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	logrus.WithField("correlation_id", correlationID).Info("Publishing EventBucketCreated")

	return h.publisher.PublishWithContext(
		ctx,
		msgBytes,
		[]string{saga.QueueBucketEvents},
		gorabbit.WithPublishOptionsContentType("application/json"),
	)
}

func (h *BucketHandler) publishBucketFailed(ctx context.Context, correlationID, userID, errorMsg string) error {
	payload := saga.BucketPayload{
		UserID: userID,
		Error:  errorMsg,
	}

	msg, err := saga.NewSagaMessage(correlationID, saga.EventBucketFailed, payload)
	if err != nil {
		return fmt.Errorf("create saga message: %w", err)
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"correlation_id": correlationID,
		"error":          errorMsg,
	}).Error("Publishing EventBucketFailed")

	return h.publisher.PublishWithContext(
		ctx,
		msgBytes,
		[]string{saga.QueueBucketEvents},
		gorabbit.WithPublishOptionsContentType("application/json"),
	)
}
