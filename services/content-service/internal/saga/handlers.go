package saga

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/content-service/internal/domain/entity"
	"soa-video-streaming/services/content-service/internal/mocks"
	"soa-video-streaming/services/content-service/internal/repository/postgres"
	"soa-video-streaming/services/orchestrator-service/domain"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewBucketsService,
		),
		fx.Invoke(RegisterBucketsActor),
	)
}

type BucketsService struct {
	storageRepo *postgres.StorageRepository
	s3Mock      *mocks.S3Mock
}

func NewBucketsService(
	storageRepo *postgres.StorageRepository,
	s3Mock *mocks.S3Mock,
) *BucketsService {
	return &BucketsService{
		storageRepo: storageRepo,
		s3Mock:      s3Mock,
	}
}

func (h *BucketsService) HandleCreateBucket(ctx context.Context, msg *saga.Message) (any, error) {
	var payload domain.BucketPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	bucketName, err := h.s3Mock.CreateBucket(payload.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to create bucket")
		return nil, err
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
		return nil, err
	}

	return domain.BucketPayload{
		UserID:     payload.UserID,
		BucketName: bucketName,
	}, nil
}

func (h *BucketsService) HandleCompensateBucket(ctx context.Context, msg *saga.Message) (any, error) {
	var payload domain.CompensateUserSignUpPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, fmt.Errorf("unmarshal payload: %w", err)
	}

	account, err := h.storageRepo.FindByUserID(ctx, payload.UserID)
	if err != nil {
		logrus.WithError(err).Error("Failed to find storage account")
		return nil, err
	}

	if account == nil {
		logrus.WithField("user_id", payload.UserID).Warn("Storage account not found for compensation")
		return nil, nil
	}

	if err := h.s3Mock.DeleteBucket(account.BucketName); err != nil {
		logrus.WithError(err).Error("Failed to delete bucket from S3")
	}

	if err := h.storageRepo.Delete(ctx, payload.UserID); err != nil {
		return nil, err
	}

	return nil, nil
}
