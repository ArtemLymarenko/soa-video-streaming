package main

import (
	"flag"
	"go.uber.org/fx"
	"soa-video-streaming/services/ai-assitant/internal/config"
	"soa-video-streaming/services/ai-assitant/pkg/gemini"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		gemini.Module(),
	).Run()
}
