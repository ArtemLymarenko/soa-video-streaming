package main

import (
	"flag"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/services/ai-assistant/internal/config"
	"soa-video-streaming/services/ai-assistant/internal/controller/rest"
	"soa-video-streaming/services/ai-assistant/internal/service"
	resttransport "soa-video-streaming/services/ai-assistant/internal/transport/rest"
	"soa-video-streaming/services/ai-assistant/pkg/gemini"
	"soa-video-streaming/services/content-service/pkg/client"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		gemini.Module(),
		httpsrv.Module(),
		service.Module(),
		rest.Module(),
		resttransport.Module(),
		fx.Provide(
			func(cfg *config.AppConfig) *client.ContentServiceClient {
				return client.NewContentServiceClient(cfg.Services.ContentServiceURL)
			},
		),
	).Run()
}
