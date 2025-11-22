package main

import (
	"flag"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/services/content-service/internal/config"
	"soa-video-streaming/services/content-service/internal/repository/postgres"
	"soa-video-streaming/services/content-service/internal/service"
	"soa-video-streaming/services/content-service/internal/transport/rest"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		httpsrv.Module(),
		postgres.Module(),
		service.Module(),
		rest.Module(),
		fx.Invoke(config.Invoke),
	).Run()
}
