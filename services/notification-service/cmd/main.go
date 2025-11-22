package main

import (
	"flag"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/notification-service/internal/config"
	"soa-video-streaming/services/notification-service/internal/service"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		httpsrv.Module(),
		rabbitmq.Module(),
		service.Module(),
	).Run()
}
