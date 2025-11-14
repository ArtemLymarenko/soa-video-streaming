package grpctransport

import (
	"soa-video-streaming/pkg/grpcsrv"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	pb "soa-video-streaming/pkg/pb/user"
	grpcCtrl "soa-video-streaming/services/user-service/internal/controller/grpc"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(grpcCtrl.NewUserController),
		fx.Provide(
			fx.Annotate(
				newRegistrar,
				fx.As(new(grpcsrv.Registrar)),
				fx.ResultTags(`group:"grpc-registrars"`),
			),
		),
	)
}

type userRegistrar struct {
	svc pb.UserServiceServer
}

func newRegistrar(svc *grpcCtrl.UsersController) *userRegistrar {
	return &userRegistrar{
		svc: svc,
	}
}

func (r *userRegistrar) Register(s *grpc.Server) {
	pb.RegisterUserServiceServer(s, r.svc)
}
