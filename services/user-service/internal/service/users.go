package service

import (
	"context"
	"errors"
	"fmt"
	"soa-video-streaming/services/user-service/internal/cache"
	"soa-video-streaming/services/user-service/internal/domain/entity"
	"soa-video-streaming/services/user-service/internal/repository/postgres"
)

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrUserCategoriesNotSet = errors.New("failed to set categories for user")
	ErrCategoryNotFound     = errors.New("category not found")
)

type UsersService struct {
	usersRepo       *postgres.UsersRepository
	userInfoRepo    *postgres.UserInfoRepository
	userPreference  *postgres.UserPreference
	categoriesCache *cache.CategoryCache
}

func NewUsersService(
	u *postgres.UsersRepository,
	ui *postgres.UserInfoRepository,
	up *postgres.UserPreference,
	cc *cache.CategoryCache,
) *UsersService {
	return &UsersService{
		usersRepo:       u,
		userInfoRepo:    ui,
		userPreference:  up,
		categoriesCache: cc,
	}
}

func (u *UsersService) GetUserByID(ctx context.Context, id string) (entity.User, error) {
	user, err := u.usersRepo.FindById(ctx, id)
	if err != nil {
		return entity.User{}, fmt.Errorf("%w: %v", ErrUserNotFound, err)
	}

	return user, nil
}

func (u *UsersService) AddPreferenceCategories(ctx context.Context, userID string, categoryIds []string) error {
	for _, id := range categoryIds {
		if _, ok := u.categoriesCache.Get(id); !ok {
			return fmt.Errorf("%w: id: %s", ErrCategoryNotFound, id)
		}
	}

	err := u.userPreference.AddPreferredCategories(ctx, userID, categoryIds)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUserCategoriesNotSet, err)
	}

	return nil
}

func (u *UsersService) GetUserCategories(ctx context.Context, userID string) ([]string, error) {
	categories, err := u.userPreference.GetUserPreferredCategories(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUserCategoriesNotSet, err)
	}

	return categories, nil
}
