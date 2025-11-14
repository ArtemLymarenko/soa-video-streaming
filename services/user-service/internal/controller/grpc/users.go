package grpc

import (
	"context"
	pb "soa-video-streaming/pkg/pb/user"
	"time"
)

type UsersController struct {
	pb.UnsafeUserServiceServer
}

func NewUserController() *UsersController {
	return &UsersController{}
}

func (u *UsersController) GetUserInfoByID(ctx context.Context, req *pb.GetUserInfoByIDRequest) (*pb.GetUserInfoByIDResponse, error) {
	return &pb.GetUserInfoByIDResponse{
		User: &pb.User{
			Id:        req.GetId(),
			Email:     "some@email.com",
			Name:      "name",
			CreatedAt: time.Now().Unix(),
		},
	}, nil
}
