package main

import (
	"flag"
	"soa-video-streaming/pkg/grpcsrv"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/pkg/postgres"
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/user-service/internal/cache"
	"soa-video-streaming/services/user-service/internal/config"
	postgresRepos "soa-video-streaming/services/user-service/internal/repository/postgres"
	"soa-video-streaming/services/user-service/internal/service"
	grpcTransport "soa-video-streaming/services/user-service/internal/transport/grpc"
	restTransport "soa-video-streaming/services/user-service/internal/transport/rest"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		grpcsrv.Module(),
		grpcsrv.ClientModule(),
		httpsrv.Module(),
		rabbitmq.Module(),
		postgres.Module(),
		grpcTransport.Module(),
		grpcTransport.ClientModule(),
		restTransport.Module(),
		postgresRepos.Module(),
		service.Module(),
		cache.Module(),
	).Run()
}
