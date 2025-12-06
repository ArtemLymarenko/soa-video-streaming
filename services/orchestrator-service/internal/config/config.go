package config

import (
	"soa-video-streaming/pkg/config"
	"soa-video-streaming/pkg/postgres"
	"soa-video-streaming/pkg/rabbitmq"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type AppConfig struct {
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
			ProvideRabbitMQConfig,
			ProvidePostgresConfig,
		),
		fx.Invoke(func(cfg *AppConfig) {
			logrus.WithField("config", cfg).Info("Config loaded")
		}),
	)
}
