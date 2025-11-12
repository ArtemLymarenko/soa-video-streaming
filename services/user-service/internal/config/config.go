package config

import (
	"soa-video-streaming/pkg/config"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type AppConfig struct {
	config.BaseHTTPServerConfig `mapstructure:",squash"`
	config.BaseGRPCServerConfig `mapstructure:",squash"`
}

func NewAppConfig() (*AppConfig, error) {
	return config.NewViper[AppConfig]()
}

func ProvideHTTPConfig(ac *AppConfig) *config.BaseHTTPServerConfig {
	return &ac.BaseHTTPServerConfig
}

func ProvideGRPCConfig(ac *AppConfig) *config.BaseGRPCServerConfig {
	return &ac.BaseGRPCServerConfig
}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewAppConfig,
			ProvideHTTPConfig,
			ProvideGRPCConfig,
		),
		fx.Invoke(func(cfg *AppConfig) {
			logrus.WithField("config", cfg).Info("Config loaded")
		}),
	)
}
