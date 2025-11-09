package config

import (
	"soa-video-streaming/pkg/config"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type AppConfig struct {
	HTTP struct {
		Addr string `yaml:"addr"`
	} `yaml:"http"`
}

func Module() fx.Option {
	return fx.Provide(
		func() (*AppConfig, error) {
			return config.NewViper[AppConfig]()
		},
	)
}

func Invoke(cfg *AppConfig) {
	logrus.WithField("config", cfg).Info("Config loaded")
}
