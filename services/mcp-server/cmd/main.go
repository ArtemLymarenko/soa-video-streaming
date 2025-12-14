package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"soa-video-streaming/services/mcp-server/config"
	"strings"

	"soa-video-streaming/services/content-service/pkg/client"
	"soa-video-streaming/services/content-service/pkg/dto"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type CheckCategoryInput struct {
	Name string `json:"name" jsonschema:"Category name"`
}

type CreateCategoryInput struct {
	Name        string  `json:"name" jsonschema:"Category name"`
	Description *string `json:"description,omitempty" jsonschema:"Optional category description"`
}

type MovieType string

const (
	MovieTypeMovie  MovieType = "movie"
	MovieTypeSeries MovieType = "series"
)

type CreateMovieInput struct {
	Title        string    `json:"title" jsonschema:"Movie title"`
	Description  *string   `json:"description,omitempty" jsonschema:"Optional movie description"`
	CategoryName string    `json:"category_name" jsonschema:"Existing category name"`
	Type         MovieType `json:"type" jsonschema:"Movie type: movie or series"`
	Duration     int       `json:"duration" jsonschema:"Duration in minutes"`
}

type EmptyOutput struct{}

func main() {
	flag.Parse()

	cfg, err := config.NewAppConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	contentClient := client.NewContentServiceClient(cfg.Services.ContentServiceAddr)

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "content-mcp-server",
		Version: "1.0.0",
	}, &mcp.ServerOptions{
		Logger: logger,
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "CheckCategory",
		Description: "Check if a movie category exists by name",
	}, func(ctx context.Context, request *mcp.CallToolRequest, input CheckCategoryInput) (*mcp.CallToolResult, EmptyOutput, error) {
		cat, err := contentClient.GetCategoryByName(ctx, input.Name)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to get category: %v", err)}},
			}, EmptyOutput{}, nil
		}

		var resp string
		if cat == nil {
			resp = `{"exists": false}`
		} else {
			resp = fmt.Sprintf(`{"exists": true, "id": "%s", "name": "%s"}`, cat.ID, cat.Name)
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: resp}},
		}, EmptyOutput{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "CreateCategory",
		Description: "Create a new movie category",
	}, func(ctx context.Context, request *mcp.CallToolRequest, input CreateCategoryInput) (*mcp.CallToolResult, EmptyOutput, error) {
		req := dto.CreateCategoryRequest{
			Name:        input.Name,
			Description: *input.Description,
		}
		cat, err := contentClient.CreateCategory(ctx, req)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to create category: %v", err)}},
			}, EmptyOutput{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(`{"status": "created", "id": "%s"}`, cat.ID)}},
		}, EmptyOutput{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "CreateMovie",
		Description: "Create a new movie or series",
	}, func(ctx context.Context, request *mcp.CallToolRequest, input CreateMovieInput) (*mcp.CallToolResult, EmptyOutput, error) {
		cat, err := contentClient.GetCategoryByName(ctx, input.CategoryName)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to get category: %v", err)}},
			}, EmptyOutput{}, nil
		}
		if cat == nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("category '%s' does not exist", input.CategoryName)}},
			}, EmptyOutput{}, nil
		}

		req := dto.CreateMediaContentRequest{
			Name:        input.Title,
			Description: *input.Description,
			Type:        strings.ToLower(string(input.Type)),
			Duration:    input.Duration,
			Categories:  []string{cat.ID},
		}

		id, err := contentClient.CreateMediaContent(ctx, req)
		if err != nil {
			return &mcp.CallToolResult{
				IsError: true,
				Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("failed to create media content: %v", err)}},
			}, EmptyOutput{}, nil
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(`{"status": "created", "id": "%s"}`, id)}},
		}, EmptyOutput{}, nil
	},
	)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
