package main

import (
	"flag"
	"soa-video-streaming/pkg/grpcsrv"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/pkg/postgres"
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/content-service/internal/config"
	"soa-video-streaming/services/content-service/internal/mocks"
	postgresRepos "soa-video-streaming/services/content-service/internal/repository/postgres"
	"soa-video-streaming/services/content-service/internal/saga"
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
		rabbitmq.Module(),
		postgresRepos.Module(),
		grpcsrv.Module(),
		grpcsrv.ClientModule(),
		service.Module(),
		rest.Module(),
		saga.Module(),
		grpctransport.ClientModule(),
		grpctransport.Module(),
		mocks.Module(),
	).Run()
}
