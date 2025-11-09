package main

import (
	"flag"
	"soa-video-streaming/services/user-service/internal/app/http"
	"soa-video-streaming/services/user-service/internal/config"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		http.Module(),
		fx.Invoke(config.Invoke),
	).Run()
}
