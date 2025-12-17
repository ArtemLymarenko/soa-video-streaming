package config

import (
	"soa-video-streaming/pkg/config"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type Config struct {
	App struct {
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
	} `mapstructure:"app"`
	Services struct {
		ContentServiceURL string `mapstructure:"content_service_url"`
	} `mapstructure:"services"`
	Gemini struct {
		APIKey string `mapstructure:"api_key"`
	} `mapstructure:"gemini"`
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
}

func NewConfig() (*Config, error) {
	return config.NewViper[Config]()
}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewConfig,
		),
		fx.Invoke(func(cfg *Config) {
			logrus.WithField("config", cfg).Info("Config loaded")
		}),
	)
}
