package main

import (
	"flag"
	"soa-video-streaming/pkg/grpcsrv"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/services/user-service/internal/config"
	grpcTransport "soa-video-streaming/services/user-service/internal/transport/grpc"
	restTransport "soa-video-streaming/services/user-service/internal/transport/rest"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		grpcsrv.Module(),
		httpsrv.Module(),

		grpcTransport.Module(),
		restTransport.Module(),
		//rabbitmq.Module(),
	).Run()
}
