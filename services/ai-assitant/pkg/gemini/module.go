package gemini

import (
	"context"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			LoadConfig,
			provideClient,
		),
	)
}

func provideClient(lc fx.Lifecycle, config *Config) (*Client, error) {
	ctx := context.Background()
	client, err := NewClientFromConfig(ctx, config)
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

func ProvideModelConfig(name string, config *ModelConfig) fx.Option {
	return fx.Provide(
		fx.Annotate(
			func(client *Client) (*MessageProcessor, error) {
				if err := client.RegisterModel(config); err != nil {
					return nil, err
				}
				return NewMessageProcessor(client, name), nil
			},
		),
	)
}
