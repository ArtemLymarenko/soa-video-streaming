package gemini

import (
	"context"
	"fmt"

	"google.golang.org/genai"
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

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &Client{
		client: client,
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

type GenerativeModel struct {
	client *genai.Client
	config *genai.GenerateContentConfig
	name   string
}

type ChatSession struct {
	model   *GenerativeModel
	History []*genai.Content
}

func (c *Client) GetGenerativeModel(_ context.Context, modelName string) (*GenerativeModel, error) {
	config, err := c.GetModel(modelName)
	if err != nil {
		return nil, err
	}

	genaiConfig := &genai.GenerateContentConfig{}

	if config.Temperature != nil {
		genaiConfig.Temperature = config.Temperature
	}

	if config.TopK != nil {
		val := float32(*config.TopK)
		genaiConfig.TopK = &val
	}

	if config.TopP != nil {
		genaiConfig.TopP = config.TopP
	}

	if config.MaxOutputTokens != nil {
		genaiConfig.MaxOutputTokens = *config.MaxOutputTokens
	}

	// System Instruction тепер частина конфігу
	if config.SystemInstruction != "" {
		genaiConfig.SystemInstruction = &genai.Content{
			Parts: []*genai.Part{
				{Text: config.SystemInstruction},
			},
		}
	}

	if len(config.Tools) > 0 {
		tools := make([]*genai.Tool, len(config.Tools))
		for i, tool := range config.Tools {
			tools[i] = tool.ToGenaiTool()
		}
		genaiConfig.Tools = tools
	}

	return &GenerativeModel{
		client: c.client,
		config: genaiConfig,
		name:   config.Name,
	}, nil
}

func (c *Client) Close() error {
	return nil
}

func (m *GenerativeModel) StartChat() *ChatSession {
	return &ChatSession{
		model:   m,
		History: make([]*genai.Content, 0),
	}
}

func (s *ChatSession) SendMessage(ctx context.Context, parts ...*genai.Part) (*genai.GenerateContentResponse, error) {
	if len(parts) > 0 {
		msg := &genai.Content{
			Role:  "user",
			Parts: parts,
		}

		if parts[0].FunctionResponse != nil {
			msg.Role = "tool"
		}

		s.History = append(s.History, msg)
	}

	resp, err := s.model.client.Models.GenerateContent(ctx, s.model.name, s.History, s.model.config)
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) > 0 && resp.Candidates[0].Content != nil {
		s.History = append(s.History, resp.Candidates[0].Content)
	}

	return resp, nil
}
