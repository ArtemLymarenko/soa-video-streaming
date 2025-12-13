package config

import (
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"soa-video-streaming/pkg/config"
	"soa-video-streaming/pkg/httpsrv"
	"soa-video-streaming/services/ai-assitant/pkg/gemini"
)

type AppConfig struct {
	HTTP struct {
		httpsrv.Config `mapstructure:",squash"`
	} `mapstructure:"http"`

	Gemini struct {
		gemini.Config `mapstructure:",squash"`
	} `mapstructure:"gemini"`
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
