package config

import (
	"soa-video-streaming/pkg/config"
	"soa-video-streaming/pkg/grpcsrv"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/pkg/postgres"
	"soa-video-streaming/pkg/rabbitmq"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type AppConfig struct {
	GRPC struct {
		grpcsrv.Config `mapstructure:",squash"`
	} `mapstructure:"grpc"`

	HTTP struct {
		httpsrv.Config `mapstructure:",squash"`
	} `mapstructure:"http"`

	RabbitMQ struct {
		rabbitmq.Config `mapstructure:",squash"`
	} `mapstructure:"rabbitmq"`

	Postgres struct {
		postgres.Config `mapstructure:",squash"`
	} `mapstructure:"postgres"`
}

func NewAppConfig() (*AppConfig, error) {
	return config.NewViper[AppConfig]()
}

func ProvideHTTPConfig(ac *AppConfig) *httpsrv.Config {
	return &ac.HTTP.Config
}

func ProvideGRPCConfig(ac *AppConfig) *grpcsrv.Config {
	return &ac.GRPC.Config
}

func ProvideRabbitMQConfig(ac *AppConfig) *rabbitmq.Config {
	return &ac.RabbitMQ.Config
}

func ProvidePostgresConfig(ac *AppConfig) *postgres.Config {
	return &ac.Postgres.Config
}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewAppConfig,
			ProvideHTTPConfig,
			ProvideGRPCConfig,
			ProvideRabbitMQConfig,
			ProvidePostgresConfig,
		),
		fx.Invoke(func(cfg *AppConfig) {
			logrus.WithField("config", cfg).Info("Config loaded")
		}),
	)
}
