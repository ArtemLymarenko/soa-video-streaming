package config

import (
	"soa-video-streaming/pkg/config"
)

type AppConfig struct {
	Services struct {
		ContentServiceAddr string `mapstructure:"content_service_addr"`
	} `mapstructure:"services"`
}

func NewAppConfig() (*AppConfig, error) {
	return config.NewViper[AppConfig]()
}
