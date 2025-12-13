package gemini

import (
	"context"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewClientFromConfig,
		),
	)
}

type Config struct {
	APIKey string `mapstructure:"api_key"`
}

func NewClientFromConfig(ctx context.Context, lc fx.Lifecycle, config *Config) (*Client, error) {
	client, err := NewClient(ctx, config.APIKey)
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})

	return client, nil
}
