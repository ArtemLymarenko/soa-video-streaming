package service

import (
	"context"
	"soa-video-streaming/services/content-service/internal/domain/entity"
	"soa-video-streaming/services/content-service/internal/repository/postgres"
)

type MediaContentService struct {
	repo *postgres.MediaContentRepository
}

func NewMediaContentService(repo *postgres.MediaContentRepository) *MediaContentService {
	return &MediaContentService{
		repo: repo,
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
