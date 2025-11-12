package main

import (
	"flag"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/services/user-service/internal/config"
	ctrlV1 "soa-video-streaming/services/user-service/internal/controller/v1"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		httpsrv.Module(),
		ctrlV1.Module(),
		fx.Invoke(config.Invoke),
		fx.Invoke(httpsrv.Invoke),
	).Run()
}
