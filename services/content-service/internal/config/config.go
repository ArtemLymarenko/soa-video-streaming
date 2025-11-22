package config

import (
	"soa-video-streaming/pkg/config"
	"soa-video-streaming/pkg/httpsrv"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type AppConfig struct {
	httpsrv.Config `mapstructure:",squash"`
}

func NewAppConfig() (*AppConfig, error) {
	return config.NewViper[AppConfig]()
}

func ProvideHTTPConfig(ac *AppConfig) *httpsrv.Config {
	return &ac.Config
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
