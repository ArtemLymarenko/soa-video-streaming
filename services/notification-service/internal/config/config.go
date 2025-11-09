package config

import (
	"soa-video-streaming/pkg/config"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type AppConfig struct {
	config.BaseHTTPServerConfig `mapstructure:",squash"`
}

func NewAppConfig() (*AppConfig, error) {
	return config.NewViper[AppConfig]()
}

func ProvideHTTPConfig(ac *AppConfig) *config.BaseHTTPServerConfig {
	return &ac.BaseHTTPServerConfig
}

func Module() fx.Option {
	return fx.Provide(
		NewAppConfig,
		ProvideHTTPConfig,
	)
}

func Invoke(cfg *AppConfig) {
	logrus.WithField("config", cfg).Info("Config loaded")
}
