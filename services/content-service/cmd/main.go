package main

import (
	"flag"
	grpcsrv "soa-video-streaming/pkg/grpcsrv"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/services/content-service/internal/config"
	"soa-video-streaming/services/content-service/internal/repository/postgres"
	"soa-video-streaming/services/content-service/internal/service"
	grpctransport "soa-video-streaming/services/content-service/internal/transport/grpc"
	"soa-video-streaming/services/content-service/internal/transport/rest"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		httpsrv.Module(),
		postgres.Module(),
		grpcsrv.Module(),
		service.Module(),
		rest.Module(),
		grpctransport.ClientModule(),
		grpctransport.Module(),
	).Run()
}
