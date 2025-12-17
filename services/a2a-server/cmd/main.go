package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/a2aproject/a2a-go/a2a"
	"net"
	"net/http"

	"github.com/a2aproject/a2a-go/a2asrv"
	"github.com/sirupsen/logrus"
	"soa-video-streaming/services/a2a-server/internal/agent"
	"soa-video-streaming/services/a2a-server/internal/config"
	"soa-video-streaming/services/a2a-server/internal/tools"
	"soa-video-streaming/services/ai-assistant/pkg/gemini"
	"soa-video-streaming/services/content-service/pkg/client"

	"go.uber.org/fx"
)

func main() {
	flag.Parse()

	fx.New(
		config.Module(),
		tools.Module(),
		agent.Module(),
		fx.Provide(
			func(cfg *config.Config) *client.ContentServiceClient {
				return client.NewContentServiceClient(cfg.Services.ContentServiceURL)
			},
			func(cfg *config.Config) (*gemini.Client, error) {
				return gemini.NewClient(context.Background(), cfg.Gemini.APIKey)
			},
			func() *logrus.Logger {
				return logrus.New()
			},
		),
		fx.Invoke(registerServer),
	).Run()
}

func registerServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	executor a2asrv.AgentExecutor,
	logger *logrus.Logger,
) {
	mux := http.NewServeMux()

	handler := a2asrv.NewHandler(executor)

	jsonRPC := a2asrv.NewJSONRPCHandler(handler)

	mux.Handle("/v1/execute", jsonRPC)

	mux.HandleFunc("/v1/agent-card", func(w http.ResponseWriter, r *http.Request) {
		card := a2a.AgentCard{
			Name:        "ContentProducer",
			Description: "I can create movies and categories in the content service.",
		}
		fmt.Fprintf(w, `{"topic": "%s", "description": "%s"}`, card.Name, card.Description)
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: mux,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logger.Infof("Starting A2A Server on %s", srv.Addr)
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}
