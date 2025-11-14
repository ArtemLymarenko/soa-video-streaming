package service

import (
	"context"
	"soa-video-streaming/services/user-service/internal/domain/entity"
	"soa-video-streaming/services/user-service/internal/repository/postgres"
)

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
	return entity.User{}, nil
}
