package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/fx"
	"strings"

	"soa-video-streaming/services/ai-assistant/pkg/gemini"
	"soa-video-streaming/services/content-service/pkg/client"
	"soa-video-streaming/services/content-service/pkg/dto"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewAssistant,
		),
	)
}

type Assistant struct {
	client        *gemini.Client
	contentClient *client.ContentServiceClient
	processor     *gemini.MessageProcessor
}

func NewAssistant(client *gemini.Client, contentClient *client.ContentServiceClient) (*Assistant, error) {
	a := &Assistant{
		client:        client,
		contentClient: contentClient,
	}

	var temp float32 = 0.7
	var maxTokens int32 = 2048

	modelConfig := &gemini.ModelConfig{
		Name:            "models/gemini-1.5-flash",
		Temperature:     &temp,
		MaxOutputTokens: &maxTokens,
	}

	checkCategoryTool := gemini.NewTool("CheckCategory", "Check if a movie category exists by name").
		AddParameter("name", "string", "Name of the category to check").
		MarkRequired("name")

	createCategoryTool := gemini.NewTool("CreateCategory", "Create a new movie category").
		AddParameter("name", "string", "Name of the category to create").
		AddParameter("description", "string", "Description of the category").
		MarkRequired("name", "description")

	createMovieTool := gemini.NewTool("CreateMovie", "Create a new movie or series").
		AddParameter("title", "string", "Title of the movie/series").
		AddParameter("description", "string", "Description of the content").
		AddParameter("category_name", "string", "Name of the category (must exist)").
		AddEnumParameter("type", "Type of content (movie or series)", "movie", "series").
		AddParameter("duration", "integer", "Duration in minutes").
		MarkRequired("title", "category_name", "type")

	modelConfig.AddTool(checkCategoryTool)
	modelConfig.AddTool(createCategoryTool)
	modelConfig.AddTool(createMovieTool)

	if err := client.RegisterModel(modelConfig); err != nil {
		return nil, fmt.Errorf("failed to register model: %w", err)
	}

	a.processor = gemini.NewMessageProcessor(client, "models/gemini-1.5-flash")

	a.processor.RegisterHandler("CheckCategory", a.HandleCheckCategory)
	a.processor.RegisterHandler("CreateCategory", a.HandleCreateCategory)
	a.processor.RegisterHandler("CreateMovie", a.HandleCreateMovie)

	return a, nil
}

func (a *Assistant) ProcessCommand(ctx context.Context, prompt string) (string, error) {
	if err := a.TokenGuardrail(prompt); err != nil {
		return "", err
	}

	if err := a.PromptInjectionGuardrail(prompt); err != nil {
		return "", err
	}

	resp, err := a.processor.ProcessWithToolLoop(ctx, prompt, 10)
	if err != nil {
		return "", err
	}

	return resp.Text, nil
}

const MaxInputTokens = 1000

func (a *Assistant) TokenGuardrail(prompt string) error {
	estimatedTokens := len(prompt) / 4
	if estimatedTokens > MaxInputTokens {
		return fmt.Errorf("token limit exceeded: input too long")
	}

	return nil
}

// Can be better
func (a *Assistant) PromptInjectionGuardrail(prompt string) error {
	forbiddenPhrases := []string{
		"ignore previous instructions",
		"system override",
		"delete database",
	}

	lowerPrompt := strings.ToLower(prompt)
	for _, phrase := range forbiddenPhrases {
		if strings.Contains(lowerPrompt, phrase) {
			return fmt.Errorf("potential prompt injection detected")
		}
	}
	return nil
}

func (a *Assistant) HandleCheckCategory(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
	var params struct {
		Name string `json:"name"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	cat, err := a.contentClient.GetCategoryByName(ctx, params.Name)
	if err != nil {
		return "", err
	}

	if cat == nil {
		return `{"exists": false}`, nil
	}

	return fmt.Sprintf(`{"exists": true, "id": "%s", "name": "%s"}`, cat.ID, cat.Name), nil
}

func (a *Assistant) HandleCreateCategory(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
	var params struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	req := dto.CreateCategoryRequest{
		Name:        params.Name,
		Description: params.Description,
	}
	cat, err := a.contentClient.CreateCategory(ctx, req)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`{"status": "created", "id": "%s"}`, cat.ID), nil
}

func (a *Assistant) HandleCreateMovie(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
	var params struct {
		Title        string          `json:"title"`
		Description  string          `json:"description"`
		CategoryName string          `json:"category_name"`
		Type         string          `json:"type"`
		Duration     json.RawMessage `json:"duration"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	var duration int
	var dInt int
	var dStr string
	if err := json.Unmarshal(params.Duration, &dInt); err == nil {
		duration = dInt
	}

	if err := json.Unmarshal(params.Duration, &dStr); err == nil {
		fmt.Sscanf(dStr, "%d", &duration)
	}

	cat, err := a.contentClient.GetCategoryByName(ctx, params.CategoryName)
	if err != nil {
		return "", err
	}
	if cat == nil {
		return "", fmt.Errorf("category '%s' does not exist", params.CategoryName)
	}

	req := dto.CreateMediaContentRequest{
		Name:        params.Title,
		Description: params.Description,
		Type:        strings.ToLower(params.Type),
		Duration:    duration,
		Categories:  []string{cat.ID},
	}

	id, err := a.contentClient.CreateMediaContent(ctx, req)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(`{"status": "created", "id": "%s"}`, id), nil
}
