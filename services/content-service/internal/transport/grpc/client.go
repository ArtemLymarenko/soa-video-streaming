package grpctransport

import (
	userpb "soa-video-streaming/pkg/pb/user"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func ClientModule() fx.Option {
	return fx.Options(
		fx.Provide(
			NewUserServiceClient,
		),
	)
}

func NewUserServiceClient(conn *grpc.ClientConn) userpb.UserServiceClient {
	return userpb.NewUserServiceClient(conn)
}
