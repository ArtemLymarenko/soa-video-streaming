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
	"soa-video-streaming/pkg/saga"
	"soa-video-streaming/services/orchestrator-service/internal/config"
	amqpcontrollers "soa-video-streaming/services/orchestrator-service/internal/controllers/amqp"
	postgresrepo "soa-video-streaming/services/orchestrator-service/internal/repository/postgres"
	"soa-video-streaming/services/orchestrator-service/internal/service"
	amqptransport "soa-video-streaming/services/orchestrator-service/internal/transport/amqp"
)

func main() {
	flag.Parse()

	app := fx.New(
		config.Module(),
		postgres.Module(),
		rabbitmq.Module(),
		saga.Module(),
		postgresrepo.Module(),
		service.Module(),
		amqpcontrollers.Module(),
		amqptransport.Module(),
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		logrus.WithError(err).Fatal("Failed to start orchestrator service")
		os.Exit(1)
	}

	logrus.Info("Orchestrator Service started successfully")

	<-app.Done()

	if err := app.Stop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop: %v\n", err)
		os.Exit(1)
	}
}
