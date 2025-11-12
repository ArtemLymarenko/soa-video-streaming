package main

import (
	"flag"
	httpsrv "soa-video-streaming/services/user-service/internal/app/http"
	"soa-video-streaming/services/user-service/internal/config"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		httpsrv.Module(),
		fx.Invoke(config.Invoke),
		fx.Invoke(httpsrv.Invoke),
	).Run()
}
