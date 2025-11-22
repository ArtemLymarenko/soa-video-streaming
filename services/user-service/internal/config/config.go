package config

import (
	"soa-video-streaming/pkg/config"
	"soa-video-streaming/pkg/grpcsrv"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/pkg/postgres"
	"soa-video-streaming/pkg/rabbitmq"
	"time"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type AppConfig struct {
	Auth struct {
		JwtSecretKey string        `mapstructure:"jwt_secret_key"`
		JwtTTL       time.Duration `mapstructure:"jwt_ttl"`
	} `mapstructure:"auth"`

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

	GPRCClient struct {
		Categories struct {
			grpcsrv.ClientConfig `mapstructure:",squash"`
		} `mapstructure:"categories"`
	} `mapstructure:"grpc_client"`
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

func ProvideGRPCCategoriesConfig(ac *AppConfig) *grpcsrv.ClientConfig {
	return &ac.GPRCClient.Categories.ClientConfig
}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewAppConfig,
			ProvideHTTPConfig,
			ProvideGRPCConfig,
			ProvideRabbitMQConfig,
			ProvidePostgresConfig,
			ProvideGRPCCategoriesConfig,
		),
		fx.Invoke(func(cfg *AppConfig) {
			logrus.WithField("config", cfg).Info("Config loaded")
		}),
	)
}
