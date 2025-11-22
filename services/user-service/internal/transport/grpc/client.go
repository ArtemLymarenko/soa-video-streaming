package grpctransport

import (
	contentpb "soa-video-streaming/pkg/pb/content"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

func ClientModule() fx.Option {
	return fx.Options(
		fx.Provide(
			NewCategoryGrpcClient,
		),
	)
}

func NewCategoryGrpcClient(conn *grpc.ClientConn) contentpb.CategoryServiceClient {
	return contentpb.NewCategoryServiceClient(conn)
}
