package main

import (
	"flag"
	"soa-video-streaming/pkg/grpcsrv"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/services/user-service/internal/config"
	ctrlV1 "soa-video-streaming/services/user-service/internal/controller/rest"
	grpcRegistrator "soa-video-streaming/services/user-service/internal/transport/grpc"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		httpsrv.Module(),
		ctrlV1.Module(),
		grpcsrv.Module(),
		grpcRegistrator.Module(),
	).Run()
}
