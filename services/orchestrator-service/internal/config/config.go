package config

import (
	"time"

	"github.com/spf13/viper"

	"soa-video-streaming/pkg/postgres"
	"soa-video-streaming/pkg/rabbitmq"
)

type AppConfig struct {
	Postgres postgres.Config `mapstructure:"postgres"`
	RabbitMQ rabbitmq.Config `mapstructure:"rabbitmq"`
}

func Load(configPath, configName string) (*AppConfig, error) {
	v := viper.New()
	v.AddConfigPath(configPath)
	v.SetConfigName(configName)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &AppConfig{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Set defaults if not specified
	if cfg.Postgres.MaxConns == 0 {
		cfg.Postgres.MaxConns = 25
	}
	if cfg.Postgres.MinConns == 0 {
		cfg.Postgres.MinConns = 5
	}
	if cfg.Postgres.ConnMaxIdle == 0 {
		cfg.Postgres.ConnMaxIdle = 30 * time.Minute
	}
	if cfg.Postgres.HealthCheckInt == 0 {
		cfg.Postgres.HealthCheckInt = 1 * time.Minute
	}
	if cfg.RabbitMQ.ReconnectDelay == 0 {
		cfg.RabbitMQ.ReconnectDelay = 5 * time.Second
	}

	return cfg, nil
}
