package service

import (
	"context"
	"soa-video-streaming/services/content-service/internal/domain/entity"
	"soa-video-streaming/services/content-service/internal/repository/postgres"
)

type CategoryService struct {
	repo *postgres.CategoryRepository
}

func NewCategoryService(repo *postgres.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) Create(ctx context.Context, c entity.Category) error {
	return s.repo.Create(ctx, c)
}

func (s *CategoryService) GetByID(ctx context.Context, id entity.CategoryID) (*entity.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *CategoryService) Update(ctx context.Context, c entity.Category) error {
	return s.repo.Update(ctx, c)
}

func (s *CategoryService) Delete(ctx context.Context, id entity.CategoryID) error {
	return s.repo.Delete(ctx, id)
}

func (s *CategoryService) GetByTimestamp(ctx context.Context, from, to int64) ([]entity.Category, error) {
	return s.repo.GetByTimestamp(ctx, from, to)
}
