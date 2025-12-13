package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

// ExampleBasicChat demonstrates basic chat without tools
func ExampleBasicChat(t *testing.T) {
	ctx := context.Background()

	// Create client
	client, err := NewClient(ctx, "your-api-key")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Register a model
	modelConfig := NewModelConfig("gemini-pro").
		WithSystemInstruction("You are a helpful assistant.").
		WithTemperature(0.7)

	if err := client.RegisterModel(modelConfig); err != nil {
		t.Fatalf("failed to register model: %v", err)
	}

	// Create processor
	processor := NewMessageProcessor(client, "gemini-pro")

	// Send a message
	response, err := processor.SendMessage(ctx, "Hello, what is 2+2?")
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	t.Logf("Response: %s", response.Text)
}

// ExampleWithTools demonstrates chat with function tools
func ExampleWithTools(t *testing.T) {
	ctx := context.Background()

	// Create client
	client, err := NewClient(ctx, "your-api-key")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Define tools
	calculatorTool := NewTool(
		"calculate",
		"Performs mathematical calculations",
	).
		AddParameter("expression", "string", "Mathematical expression to calculate").
		MarkRequired("expression")

	getWeatherTool := NewTool(
		"get_weather",
		"Gets the weather for a location",
	).
		AddParameter("location", "string", "City or location name").
		AddEnumParameter("unit", "Temperature unit", "celsius", "fahrenheit").
		MarkRequired("location")

	// Register model with tools
	modelConfig := NewModelConfig("gemini-pro").
		WithSystemInstruction("You are a helpful assistant with access to tools.").
		AddTool(calculatorTool).
		AddTool(getWeatherTool)

	if err := client.RegisterModel(modelConfig); err != nil {
		t.Fatalf("failed to register model: %v", err)
	}

	// Create processor and register handlers
	processor := NewMessageProcessor(client, "gemini-pro")

	processor.RegisterHandler("calculate", func(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
		var params struct {
			Expression string `json:"expression"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}
		// In real scenario, evaluate the expression
		return fmt.Sprintf("Result of '%s' = 4", params.Expression), nil
	})

	processor.RegisterHandler("get_weather", func(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
		var params struct {
			Location string `json:"location"`
			Unit     string `json:"unit"`
		}
		if err := json.Unmarshal(args, &params); err != nil {
			return "", err
		}
		return fmt.Sprintf("Weather in %s: 22°%s, Sunny", params.Location, params.Unit), nil
	})

	// Process message with automatic tool handling
	response, err := processor.ProcessWithToolLoop(ctx, "What is the weather in Kyiv?", 5)
	if err != nil {
		t.Fatalf("failed to process message: %v", err)
	}

	t.Logf("Final response: %s", response.Text)
	t.Logf("Finished: %v, Reason: %s", response.IsFinished, response.FinishReason)
}

// ExampleWithHistory demonstrates chat with conversation history
func ExampleWithHistory(t *testing.T) {
	ctx := context.Background()

	// Create client
	client, err := NewClient(ctx, "your-api-key")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Register model
	modelConfig := NewModelConfig("gemini-pro").
		WithSystemInstruction("You are a helpful assistant.")

	if err := client.RegisterModel(modelConfig); err != nil {
		t.Fatalf("failed to register model: %v", err)
	}

	// Create processor
	processor := NewMessageProcessor(client, "gemini-pro")

	// Build conversation history
	history := []Message{
		{
			Role:    "user",
			Content: "My name is John",
		},
		{
			Role:    "model",
			Content: "Nice to meet you, John!",
		},
	}

	// Continue conversation
	response, err := processor.ProcessWithHistory(
		ctx,
		history,
		"What is my name?",
	)
	if err != nil {
		t.Fatalf("failed to process message: %v", err)
	}

	t.Logf("Response: %s", response.Text)
}

// ExampleCustomModel demonstrates registering a custom model configuration
func ExampleCustomModel(t *testing.T) {
	ctx := context.Background()

	// Create client
	client, err := NewClient(ctx, "your-api-key")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Create custom model with specific parameters
	temp := float32(0.9)
	topK := int32(50)
	maxTokens := int32(4096)

	modelConfig := &ModelConfig{
		Name:              "gemini-pro-vision",
		SystemInstruction: "You are an AI vision expert",
		Temperature:       &temp,
		TopK:              &topK,
		MaxOutputTokens:   &maxTokens,
	}

	if err := client.RegisterModel(modelConfig); err != nil {
		t.Fatalf("failed to register model: %v", err)
	}

	// Create processor
	processor := NewMessageProcessor(client, "gemini-pro-vision")

	// Send message
	response, err := processor.SendMessage(ctx, "Analyze this image")
	if err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	t.Logf("Response: %s", response.Text)
}

// ExampleMultipleTools demonstrates using multiple tools in one interaction
func ExampleMultipleTools(t *testing.T) {
	ctx := context.Background()

	// Create client
	client, err := NewClient(ctx, "your-api-key")
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	defer client.Close()

	// Create multiple tools
	tools := []*Tool{
		NewTool("search", "Search the web").
			AddParameter("query", "string", "Search query").
			MarkRequired("query"),

		NewTool("fetch_url", "Fetch content from URL").
			AddParameter("url", "string", "URL to fetch").
			MarkRequired("url"),

		NewTool("summarize", "Summarize text").
			AddParameter("text", "string", "Text to summarize").
			MarkRequired("text"),
	}

	// Create model with all tools
	modelConfig := NewModelConfig("gemini-pro").
		WithSystemInstruction("You are a research assistant with access to multiple tools.")

	for _, tool := range tools {
		modelConfig.AddTool(tool)
	}

	if err := client.RegisterModel(modelConfig); err != nil {
		t.Fatalf("failed to register model: %v", err)
	}

	// Create processor with handlers
	processor := NewMessageProcessor(client, "gemini-pro")

	processor.RegisterHandler("search", func(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
		var params struct {
			Query string `json:"query"`
		}
		json.Unmarshal(args, &params)
		return fmt.Sprintf("Search results for '%s': ...", params.Query), nil
	})

	processor.RegisterHandler("fetch_url", func(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
		var params struct {
			URL string `json:"url"`
		}
		json.Unmarshal(args, &params)
		return fmt.Sprintf("Content from %s: ...", params.URL), nil
	})

	processor.RegisterHandler("summarize", func(ctx context.Context, toolName string, args json.RawMessage) (string, error) {
		var params struct {
			Text string `json:"text"`
		}
		json.Unmarshal(args, &params)
		return "Summary: ...", nil
	})

	// Process with tool loop
	response, err := processor.ProcessWithToolLoop(ctx, "Find and summarize the latest AI news", 10)
	if err != nil {
		t.Fatalf("failed to process: %v", err)
	}

	t.Logf("Final response: %s", response.Text)
	t.Logf("Total tool calls: %d", len(response.ToolCalls))
}
