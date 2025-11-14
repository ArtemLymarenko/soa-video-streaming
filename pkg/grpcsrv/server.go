package grpcsrv

import (
	"context"
	"net"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

type Config struct {
	Addr string `mapstructure:"addr"`
}

type Registrar interface {
	Register(s *grpc.Server)
}

type Params struct {
	fx.In
	Registrars []Registrar `group:"grpc-registrars"`
}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewGRPCServer),
		fx.Invoke(StartServer),
	)
}

func NewGRPCServer(p Params) *grpc.Server {
	s := grpc.NewServer()

	for _, r := range p.Registrars {
		r.Register(s)
	}

	return s
}

func StartServer(
	lc fx.Lifecycle,
	cfg *Config,
	srv *grpc.Server,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			addr := cfg.Addr
			lis, err := net.Listen("tcp", addr)
			if err != nil {
				return err
			}

			logrus.Infof("gRPC server listening at %s", addr)

			go func() {
				if err := srv.Serve(lis); err != nil {
					logrus.Errorf("gRPC server stopped: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logrus.Info("Stopping gRPC server")
			srv.GracefulStop()
			return nil
		},
	})
}
