package grpc

import (
	"context"
	pb "soa-video-streaming/pkg/pb/user"
	"soa-video-streaming/services/user-service/internal/service"
)

type UsersController struct {
	pb.UnsafeUserServiceServer

	userService *service.UsersService
}

func NewUserController(userService *service.UsersService) *UsersController {
	return &UsersController{
		userService: userService,
	}
}

func (u *UsersController) GetUserInfoByID(ctx context.Context, req *pb.GetUserInfoByIDRequest) (*pb.GetUserInfoByIDResponse, error) {
	user, err := u.userService.GetUserByID(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &pb.GetUserInfoByIDResponse{
		User: &pb.User{
			Id:        user.Id,
			Email:     user.Email,
			Name:      user.FirstName + " " + user.LastName,
			CreatedAt: user.CreatedAt.Unix(),
		},
	}, nil
}
