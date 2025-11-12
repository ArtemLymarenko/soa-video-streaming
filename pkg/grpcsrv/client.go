package grpcsrv

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func ClientModule() fx.Option {
	return fx.Options(
		fx.Provide(NewClientConfig),
		fx.Provide(NewClientConn),
	)
}

type ClientConfig struct {
	Target string
}

func NewClientConfig() ClientConfig {
	target := os.Getenv("GRPC_TARGET")
	if target == "" {
		target = "127.0.0.1:50051"
	}

	return ClientConfig{Target: target}
}

func NewClientConn(lc fx.Lifecycle, cfg ClientConfig) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		cfg.Target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error { return conn.Close() },
	})

	logrus.WithField("target", cfg.Target).Info("gRPC client connected")

	return conn, nil
}
