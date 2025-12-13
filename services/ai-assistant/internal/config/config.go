package config

import (
	"soa-video-streaming/pkg/config"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/services/ai-assistant/pkg/gemini"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type AppConfig struct {
	HTTP struct {
		httpsrv.Config `mapstructure:",squash"`
	} `mapstructure:"http"`

	Gemini struct {
		gemini.Config `mapstructure:",squash"`
	} `mapstructure:"gemini"`

	Services struct {
		ContentServiceURL string `mapstructure:"content_service_url"`
	} `mapstructure:"services"`
}

func NewAppConfig() (*AppConfig, error) {
	return config.NewViper[AppConfig]()
}

func ProvideHTTPConfig(ac *AppConfig) *httpsrv.Config {
	return &ac.HTTP.Config
}

func ProvideGeminiConfig(ac *AppConfig) *gemini.Config {
	return &ac.Gemini.Config
}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewAppConfig,
			ProvideHTTPConfig,
			ProvideGeminiConfig,
		),
		fx.Invoke(func(cfg *AppConfig) {
			logrus.WithField("config", cfg).Info("Config loaded")
		}),
	)
}
