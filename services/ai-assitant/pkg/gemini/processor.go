package gemini

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

// Message represents a message in the conversation
type Message struct {
	Role    string `json:"role"`    // "user" or "model"
	Content string `json:"content"` // text content
}

// ToolResponse represents a response from a tool execution
type ToolResponse struct {
	ToolName  string `json:"tool_name"`
	ToolUseID string `json:"tool_use_id"`
	Output    string `json:"output"`
}

// ProcessedResponse represents the final response from processing
type ProcessedResponse struct {
	Text         string
	ToolCalls    []*ToolCall
	IsFinished   bool
	FinishReason string
}

// Handler is a function that handles tool calls
// It receives the tool name and arguments and returns the result
type Handler func(ctx context.Context, toolName string, args json.RawMessage) (string, error)

// MessageProcessor handles the message processing loop with tools
type MessageProcessor struct {
	client    *Client
	modelName string
	handlers  map[string]Handler
}

// NewMessageProcessor creates a new message processor
func NewMessageProcessor(client *Client, modelName string) *MessageProcessor {
	return &MessageProcessor{
		client:    client,
		modelName: modelName,
		handlers:  make(map[string]Handler),
	}
}

// RegisterHandler registers a handler for a specific tool
func (mp *MessageProcessor) RegisterHandler(toolName string, handler Handler) {
	mp.handlers[toolName] = handler
}

// SendMessage sends a message and processes the response
func (mp *MessageProcessor) SendMessage(ctx context.Context, userMessage string) (*ProcessedResponse, error) {
	model, err := mp.client.GetGenerativeModel(ctx, mp.modelName)
	if err != nil {
		return nil, err
	}

	// Start chat session
	session := model.StartChat()

	// Send user message
	resp, err := session.SendMessage(ctx, genai.Text(userMessage))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return mp.processResponse(resp)
}

// ProcessWithHistory sends a message with conversation history
func (mp *MessageProcessor) ProcessWithHistory(
	ctx context.Context,
	messages []Message,
	userMessage string,
) (*ProcessedResponse, error) {
	model, err := mp.client.GetGenerativeModel(ctx, mp.modelName)
	if err != nil {
		return nil, err
	}

	// Start chat session
	session := model.StartChat()

	// Add history
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

	// Send new message
	resp, err := session.SendMessage(ctx, genai.Text(userMessage))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return mp.processResponse(resp)
}

// ProcessWithToolLoop processes a message with automatic tool calling
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

	// Send initial message
	resp, err := session.SendMessage(ctx, genai.Text(userMessage))
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Process response and handle tool calls in a loop
	for i := 0; i < maxIterations; i++ {
		processedResp := mp.extractResponse(resp)

		// If no tool calls, we're done
		if len(processedResp.ToolCalls) == 0 {
			return processedResp, nil
		}

		// Execute tool calls
		toolResults := make([]genai.Part, 0)
		for _, toolCall := range processedResp.ToolCalls {
			result, err := mp.executeToolCall(ctx, toolCall)
			if err != nil {
				// Include error in tool result
				result = fmt.Sprintf("Error: %v", err)
			}

			toolResults = append(toolResults, &genai.FunctionResponse{
				Name:     toolCall.Name,
				Response: map[string]interface{}{"result": result},
			})
		}

		// Send tool results back to model
		resp, err = session.SendMessage(ctx, toolResults...)
		if err != nil {
			return nil, fmt.Errorf("failed to send tool results: %w", err)
		}
	}

	// Max iterations reached
	return &ProcessedResponse{
		IsFinished:   false,
		FinishReason: "max_iterations_reached",
	}, nil
}

// executeToolCall executes a single tool call
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

// processResponse extracts tool calls and text from response
func (mp *MessageProcessor) processResponse(resp *genai.GenerateContentResponse) (*ProcessedResponse, error) {
	return mp.extractResponse(resp), nil
}

// extractResponse extracts the response details
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

	// Extract text and function calls from parts
	for _, part := range candidate.Content.Parts {
		switch v := part.(type) {
		case genai.Text:
			processed.Text += string(v)

		case *genai.FunctionCall:
			// Convert Args to JSON raw message
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

// BuildUserContent creates a user message content
func BuildUserContent(text string) *genai.Content {
	return &genai.Content{
		Role: "user",
		Parts: []genai.Part{
			genai.Text(text),
		},
	}
}

// BuildAssistantContent creates an assistant message content
func BuildAssistantContent(text string) *genai.Content {
	return &genai.Content{
		Role: "model",
		Parts: []genai.Part{
			genai.Text(text),
		},
	}
}
