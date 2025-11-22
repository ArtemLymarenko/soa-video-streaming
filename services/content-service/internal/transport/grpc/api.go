package grpctransport

import (
	"soa-video-streaming/pkg/grpcsrv"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	pb "soa-video-streaming/pkg/pb/content"
	grpcCtrl "soa-video-streaming/services/content-service/internal/controller/grpc"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(grpcCtrl.NewCategoryController),
		fx.Provide(
			fx.Annotate(
				newRegistrar,
				fx.As(new(grpcsrv.Registrar)),
				fx.ResultTags(`group:"grpc-registrars"`),
			),
		),
	)
}

type categoryRegistrar struct {
	svc pb.CategoryServiceServer
}

func newRegistrar(svc *grpcCtrl.CategoryController) *categoryRegistrar {
	return &categoryRegistrar{
		svc: svc,
	}
}

func (r *categoryRegistrar) Register(s *grpc.Server) {
	pb.RegisterCategoryServiceServer(s, r.svc)
}
