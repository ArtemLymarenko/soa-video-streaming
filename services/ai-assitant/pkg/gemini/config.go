package gemini

import (
	"context"
	"os"
)

type Config struct {
	APIKey string
}

func LoadConfig() *Config {
	return &Config{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	}
}

func NewClientFromConfig(ctx context.Context, config *Config) (*Client, error) {
	return NewClient(ctx, config.APIKey)
}
