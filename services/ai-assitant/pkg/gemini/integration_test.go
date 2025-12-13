package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// TestBasicInitialization tests basic client initialization
func TestBasicInitialization(t *testing.T) {
	ctx := context.Background()

	// Create client
	client, err := NewClient(ctx, "test-key")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Register model
	config := NewModelConfig("gemini-pro").
		WithSystemInstruction("You are a helpful assistant")

	if err := client.RegisterModel(config); err != nil {
		t.Fatalf("failed to register model: %v", err)
	}

	// Verify model registration
	model, err := client.GetModel("gemini-pro")
	if err != nil {
		t.Fatalf("failed to get model: %v", err)
	}

	if model.Name != "gemini-pro" {
		t.Errorf("expected model name 'gemini-pro', got '%s'", model.Name)
	}

	if model.SystemInstruction != "You are a helpful assistant" {
		t.Errorf("expected system instruction, got '%s'", model.SystemInstruction)
	}
}

// TestToolConfiguration tests tool configuration
func TestToolConfiguration(t *testing.T) {
	// Create tool
	tool := NewTool(
		"search",
		"Search the web for information",
	).
		AddParameter("query", "string", "Search query").
		AddParameter("maxResults", "integer", "Maximum number of results").
		MarkRequired("query")

	// Verify tool configuration
	if tool.Name != "search" {
		t.Errorf("expected tool name 'search', got '%s'", tool.Name)
	}

	if len(tool.Parameters.Properties) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(tool.Parameters.Properties))
	}

	if len(tool.Parameters.Required) != 1 {
		t.Errorf("expected 1 required parameter, got %d", len(tool.Parameters.Required))
	}

	if tool.Parameters.Required[0] != "query" {
		t.Errorf("expected 'query' to be required")
	}
}

// TestEnumParameter tests enum parameter
func TestEnumParameter(t *testing.T) {
	tool := NewTool("format", "Format data").
		AddEnumParameter("type", "Output format", "json", "xml", "csv").
		MarkRequired("type")

	if len(tool.Parameters.Properties) != 1 {
		t.Errorf("expected 1 parameter")
	}

	prop := tool.Parameters.Properties["type"]
	if len(prop.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(prop.Enum))
	}
}

// TestModelConfigBuilder tests model config builder pattern
func TestModelConfigBuilder(t *testing.T) {
	config := NewModelConfig("gemini-pro").
		WithSystemInstruction("Test instruction").
		WithTemperature(0.8).
		WithTopK(50).
		WithTopP(0.9).
		WithMaxOutputTokens(4096)

	if config.Name != "gemini-pro" {
		t.Errorf("expected name 'gemini-pro'")
	}

	if config.SystemInstruction != "Test instruction" {
		t.Errorf("expected system instruction")
	}

	if *config.Temperature != 0.8 {
		t.Errorf("expected temperature 0.8")
	}

	if *config.TopK != 50 {
		t.Errorf("expected TopK 50")
	}

	if *config.MaxOutputTokens != 4096 {
		t.Errorf("expected max tokens 4096")
	}
}

// TestMessageProcessorHandlers tests message processor handler registration
func TestMessageProcessorHandlers(t *testing.T) {
	ctx := context.Background()
	client, _ := NewClient(ctx, "test-key")
	defer client.Close()

	processor := NewMessageProcessor(client, "gemini-pro")

	// Register handlers
	processor.RegisterHandler("tool1", func(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
		return "result1", nil
	})

	processor.RegisterHandler("tool2", func(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
		return "result2", nil
	})

	if len(processor.handlers) != 2 {
		t.Errorf("expected 2 handlers, got %d", len(processor.handlers))
	}

	if _, ok := processor.handlers["tool1"]; !ok {
		t.Errorf("expected tool1 handler")
	}

	if _, ok := processor.handlers["tool2"]; !ok {
		t.Errorf("expected tool2 handler")
	}
}

// TestToolCallParsing tests tool call parsing
func TestToolCallParsing(t *testing.T) {
	toolCall := &ToolCall{
		Name:      "calculate",
		Arguments: json.RawMessage(`{"expression":"2+2"}`),
	}

	var params struct {
		Expression string `json:"expression"`
	}

	if err := json.Unmarshal(toolCall.Arguments, &params); err != nil {
		t.Fatalf("failed to unmarshal arguments: %v", err)
	}

	if params.Expression != "2+2" {
		t.Errorf("expected expression '2+2', got '%s'", params.Expression)
	}
}

// ExampleIntegrationWithService demonstrates integration with AI assistant service
func ExampleIntegrationWithService(t *testing.T) {
	ctx := context.Background()

	// 1. Create and initialize client
	client, _ := NewClient(ctx, "your-api-key")
	defer client.Close()

	// 2. Define tools for the service
	searchTool := NewTool(
		"search_database",
		"Search media in our database",
	).
		AddParameter("query", "string", "Search query").
		AddParameter("category", "string", "Category filter").
		MarkRequired("query")

	recommendTool := NewTool(
		"get_recommendations",
		"Get recommendations based on content",
	).
		AddParameter("user_id", "string", "User ID").
		AddParameter("limit", "integer", "Number of recommendations").
		MarkRequired("user_id")

	// 3. Configure model with tools
	modelConfig := NewModelConfig("gemini-pro").
		WithSystemInstruction(`You are an AI assistant for a video streaming service.
You help users find content and get personalized recommendations.
Use the available tools to assist users.`).
		WithTemperature(0.7).
		AddTool(searchTool).
		AddTool(recommendTool)

	_ = client.RegisterModel(modelConfig)

	// 4. Create processor with tool handlers
	processor := NewMessageProcessor(client, "gemini-pro")

	processor.RegisterHandler("search_database", func(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
		var params struct {
			Query    string `json:"query"`
			Category string `json:"category"`
		}
		json.Unmarshal(args, &params)

		// In real service, query your database
		return fmt.Sprintf("Found 5 results for '%s' in category '%s'", params.Query, params.Category), nil
	})

	processor.RegisterHandler("get_recommendations", func(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
		var params struct {
			UserID string `json:"user_id"`
			Limit  int    `json:"limit"`
		}
		json.Unmarshal(args, &params)

		// In real service, get recommendations from your service
		return fmt.Sprintf("Got %d recommendations for user %s", params.Limit, params.UserID), nil
	})

	// 5. Process user queries with automatic tool calling
	userQuery := "Find me some action movies and give me personalized recommendations"

	response, _ := processor.ProcessWithToolLoop(ctx, userQuery, 5)

	t.Logf("Response: %s", response.Text)
	t.Logf("Tool calls made: %d", len(response.ToolCalls))
	t.Logf("Finished: %v", response.IsFinished)
}
