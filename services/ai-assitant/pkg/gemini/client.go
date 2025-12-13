package gemini

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	client *genai.Client
	models map[string]*ModelConfig
	apiKey string
}

func NewClient(ctx context.Context, apiKey string) (*Client, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	geminiClient, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		client: geminiClient,
		models: make(map[string]*ModelConfig),
		apiKey: apiKey,
	}, nil
}

func (c *Client) RegisterModel(config *ModelConfig) error {
	if config.Name == "" {
		return fmt.Errorf("model name is required")
	}
	c.models[config.Name] = config
	return nil
}

func (c *Client) GetModel(name string) (*ModelConfig, error) {
	model, ok := c.models[name]
	if !ok {
		return nil, fmt.Errorf("model %s not registered", name)
	}
	return model, nil
}

func (c *Client) GetGenerativeModel(_ context.Context, modelName string) (*genai.GenerativeModel, error) {
	config, err := c.GetModel(modelName)
	if err != nil {
		return nil, err
	}

	model := c.client.GenerativeModel(config.Name)

	if config.Temperature != nil {
		model.SetTemperature(*config.Temperature)
	}

	if config.TopK != nil {
		model.SetTopK(*config.TopK)
	}

	if config.TopP != nil {
		model.SetTopP(*config.TopP)
	}

	if config.MaxOutputTokens != nil {
		model.SetMaxOutputTokens(*config.MaxOutputTokens)
	}

	if config.SystemInstruction != "" {
		model.SystemInstruction = &genai.Content{
			Role: "user",
			Parts: []genai.Part{
				genai.Text(config.SystemInstruction),
			},
		}
	}

	if len(config.Tools) > 0 {
		tools := make([]*genai.Tool, len(config.Tools))
		for i, tool := range config.Tools {
			tools[i] = tool.ToGenaiTool()
		}

		model.Tools = tools
	}

	return model, nil
}

func (c *Client) Close() error {
	return c.client.Close()
}
