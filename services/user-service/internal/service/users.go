package service

import (
	"context"
	"errors"
	"fmt"
	"soa-video-streaming/services/user-service/internal/domain/entity"
	"soa-video-streaming/services/user-service/internal/repository/postgres"
)

var ErrUserNotFound = errors.New("user not found")

type UsersService struct {
	usersRepo    *postgres.UsersRepository
	userInfoRepo *postgres.UserInfoRepository
}

func NewUsersService(u *postgres.UsersRepository, ui *postgres.UserInfoRepository) *UsersService {
	return &UsersService{
		usersRepo:    u,
		userInfoRepo: ui,
	}
}

func (u *UsersService) GetUserByID(ctx context.Context, id string) (entity.User, error) {
	user, err := u.usersRepo.FindById(ctx, id)
	if err != nil {
		return entity.User{}, fmt.Errorf("%w: %v", ErrUserNotFound, err)
	}

	return user, nil
}
