package service

import (
	"context"
	userv1 "soa-video-streaming/pkg/pb/user"
	"soa-video-streaming/services/content-service/internal/domain/entity"
	"soa-video-streaming/services/content-service/internal/repository/postgres"
)

type MediaContentService struct {
	repo           *postgres.MediaContentRepository
	userGRPCClient userv1.UserServiceClient
}

func NewMediaContentService(repo *postgres.MediaContentRepository, userClient userv1.UserServiceClient) *MediaContentService {
	return &MediaContentService{
		repo:           repo,
		userGRPCClient: userClient,
	}
}

func (s *MediaContentService) Create(ctx context.Context, m entity.MediaContent) error {
	return s.repo.Create(ctx, m)
}

func (s *MediaContentService) GetByID(ctx context.Context, id string) (*entity.MediaContent, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *MediaContentService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *MediaContentService) GetRecommendations(ctx context.Context, userID string, limit int64) ([]entity.MediaContent, error) {
	resp, err := s.userGRPCClient.GetUserCategories(ctx, &userv1.GetUserCategoriesRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	if len(resp.CategoryIds) == 0 {
		return nil, nil
	}

	mediaList, err := s.repo.GetRandomByCategories(ctx, resp.CategoryIds, limit)
	if err != nil {
		return nil, err
	}

	return mediaList, nil
}
