package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"

	"soa-video-streaming/pkg/postgres"
	"soa-video-streaming/pkg/rabbitmq"
	"soa-video-streaming/services/orchestrator-service/internal/config"
	postgresrepo "soa-video-streaming/services/orchestrator-service/internal/repository/postgres"
	"soa-video-streaming/services/orchestrator-service/internal/service"
)

func main() {
	configPath := flag.String("configPath", "./release/config", "Path to config file")
	configName := flag.String("configName", "local", "Config file name")
	flag.Parse()

	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	app := fx.New(
		fx.Provide(
			func() (*config.AppConfig, error) {
				return config.Load(*configPath, *configName)
			},
			func(cfg *config.AppConfig) *postgres.Config {
				return &cfg.Postgres
			},
			func(cfg *config.AppConfig) *rabbitmq.Config {
				return &cfg.RabbitMQ
			},
		),
		postgres.Module(),
		rabbitmq.Module(),
		fx.Provide(
			postgresrepo.NewSagaRepository,
			service.NewCoordinator,
		),
		fx.Invoke(service.RunConsumers),
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		logrus.WithError(err).Fatal("Failed to start orchestrator service")
		os.Exit(1)
	}

	logrus.Info("🚀 Orchestrator Service started successfully")

	<-app.Done()

	if err := app.Stop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop: %v\n", err)
		os.Exit(1)
	}
}
