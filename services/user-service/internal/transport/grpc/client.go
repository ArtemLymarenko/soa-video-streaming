package grpctransport

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"
	contentpb "soa-video-streaming/pkg/pb/content"
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
