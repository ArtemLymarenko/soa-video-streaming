package grpc

import (
	"context"
	pb "soa-video-streaming/pkg/pb/user"
	"time"
)

type UserController struct {
	pb.UnsafeUserServiceServer
}

func NewUserController() *UserController {
	return &UserController{}
}

func (u *UserController) GetUserInfoByID(ctx context.Context, req *pb.GetUserInfoByIDRequest) (*pb.GetUserInfoByIDResponse, error) {
	return &pb.GetUserInfoByIDResponse{
		User: &pb.User{
			Id:        req.GetId(),
			Email:     "some@email.com",
			Name:      "name",
			CreatedAt: time.Now().Unix(),
		},
	}, nil
}
