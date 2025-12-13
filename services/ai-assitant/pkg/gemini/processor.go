package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ToolResponse struct {
	ToolName  string `json:"tool_name"`
	ToolUseID string `json:"tool_use_id"`
	Output    string `json:"output"`
}

type ProcessedResponse struct {
	Text         string
	ToolCalls    []*ToolCall
	IsFinished   bool
	FinishReason string
}

type Handler func(ctx context.Context, toolName string, args json.RawMessage) (string, error)

type MessageProcessor struct {
	client    *Client
	modelName string
	handlers  map[string]Handler
}

func NewMessageProcessor(client *Client, modelName string) *MessageProcessor {
	return &MessageProcessor{
		client:    client,
		modelName: modelName,
		handlers:  make(map[string]Handler),
	}
}

func (mp *MessageProcessor) RegisterHandler(toolName string, handler Handler) {
	mp.handlers[toolName] = handler
}

func (mp *MessageProcessor) SendMessage(ctx context.Context, userMessage string) (*ProcessedResponse, error) {
	model, err := mp.client.GetGenerativeModel(ctx, mp.modelName)
	if err != nil {
		return nil, err
	}

	session := model.StartChat()

	resp, err := session.SendMessage(ctx, genai.Text(userMessage))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return mp.processResponse(resp)
}

func (mp *MessageProcessor) ProcessWithHistory(
	ctx context.Context,
	messages []Message,
	userMessage string,
) (*ProcessedResponse, error) {
	model, err := mp.client.GetGenerativeModel(ctx, mp.modelName)
	if err != nil {
		return nil, err
	}

	session := model.StartChat()

	for _, msg := range messages {
		role := msg.Role
		if role == "user" {
			role = "user"
		} else {
			role = "model"
		}

		session.History = append(session.History, &genai.Content{
			Role: role,
			Parts: []genai.Part{
				genai.Text(msg.Content),
			},
		})
	}

	resp, err := session.SendMessage(ctx, genai.Text(userMessage))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return mp.processResponse(resp)
}

func (mp *MessageProcessor) ProcessWithToolLoop(
	ctx context.Context,
	userMessage string,
	maxIterations int,
) (*ProcessedResponse, error) {
	if maxIterations <= 0 {
		maxIterations = 5
	}

	model, err := mp.client.GetGenerativeModel(ctx, mp.modelName)
	if err != nil {
		return nil, err
	}

	session := model.StartChat()

	resp, err := session.SendMessage(ctx, genai.Text(userMessage))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	for i := 0; i < maxIterations; i++ {
		processedResp := mp.extractResponse(resp)

		if len(processedResp.ToolCalls) == 0 {
			return processedResp, nil
		}

		toolResults := make([]genai.Part, 0)
		for _, toolCall := range processedResp.ToolCalls {
			result, err := mp.executeToolCall(ctx, toolCall)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			toolResults = append(toolResults, &genai.FunctionResponse{
				Name:     toolCall.Name,
				Response: map[string]interface{}{"result": result},
			})
		}

		resp, err = session.SendMessage(ctx, toolResults...)
		if err != nil {
			return nil, fmt.Errorf("failed to send tool results: %w", err)
		}
	}

	return &ProcessedResponse{
		IsFinished:   false,
		FinishReason: "max_iterations_reached",
	}, nil
}

func (mp *MessageProcessor) executeToolCall(ctx context.Context, toolCall *ToolCall) (string, error) {
	handler, ok := mp.handlers[toolCall.Name]
	if !ok {
		return "", fmt.Errorf("no handler registered for tool: %s", toolCall.Name)
	}

	result, err := handler(ctx, toolCall.Name, toolCall.Arguments)
	if err != nil {
		return "", err
	}

	return result, nil
}

func (mp *MessageProcessor) processResponse(resp *genai.GenerateContentResponse) (*ProcessedResponse, error) {
	return mp.extractResponse(resp), nil
}

func (mp *MessageProcessor) extractResponse(resp *genai.GenerateContentResponse) *ProcessedResponse {
	processed := &ProcessedResponse{
		ToolCalls:  make([]*ToolCall, 0),
		IsFinished: true,
	}

	if resp == nil || len(resp.Candidates) == 0 {
		return processed
	}

	candidate := resp.Candidates[0]
	processed.FinishReason = string(candidate.FinishReason)

	if candidate.Content == nil {
		return processed
	}

	for _, part := range candidate.Content.Parts {
		switch v := part.(type) {
		case genai.Text:
			processed.Text += string(v)

		case *genai.FunctionCall:
			argsData, _ := json.Marshal(v.Args)
			toolCall := &ToolCall{
				Name:      v.Name,
				Arguments: argsData,
			}

			processed.ToolCalls = append(processed.ToolCalls, toolCall)
		}
	}

	return processed
}

func BuildUserContent(text string) *genai.Content {
	return &genai.Content{
		Role: "user",
		Parts: []genai.Part{
			genai.Text(text),
		},
	}
}

func BuildAssistantContent(text string) *genai.Content {
	return &genai.Content{
		Role: "model",
		Parts: []genai.Part{
			genai.Text(text),
		},
	}
}
