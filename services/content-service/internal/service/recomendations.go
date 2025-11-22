package service

import (
	"context"
	userv1 "soa-video-streaming/pkg/pb/user"
	"soa-video-streaming/services/content-service/internal/domain/entity"
	"soa-video-streaming/services/content-service/internal/repository/postgres"
)

type Recommendations struct {
	repo           *postgres.MediaContentRepository
	userGRPCClient userv1.UserServiceClient
}

func NewRecommendations(userGRPCClient userv1.UserServiceClient) *Recommendations {
	return &Recommendations{
		userGRPCClient: userGRPCClient,
	}
}

func (s *Recommendations) GetRecommendations(ctx context.Context, userID string, limit int64) ([]entity.MediaContent, error) {
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
