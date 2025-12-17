package tools

import (
	"context"
	"fmt"
	"strings"

	"soa-video-streaming/services/content-service/pkg/client"
	"soa-video-streaming/services/content-service/pkg/dto"

	"go.uber.org/fx"
)

type ContentTools struct {
	client *client.ContentServiceClient
}

func NewContentTools(client *client.ContentServiceClient) *ContentTools {
	return &ContentTools{client: client}
}

func Module() fx.Option {
	return fx.Options(
		fx.Provide(NewContentTools),
	)
}

type CheckCategoryInput struct {
	Name string `json:"name"`
}

type CheckCategoryOutput struct {
	Exists bool   `json:"exists"`
	ID     string `json:"id,omitempty"`
	Name   string `json:"name,omitempty"`
}

func (t *ContentTools) CheckCategory(ctx context.Context, input CheckCategoryInput) (*CheckCategoryOutput, error) {
	cat, err := t.client.GetCategoryByName(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	if cat == nil {
		return &CheckCategoryOutput{Exists: false}, nil
	}

	return &CheckCategoryOutput{
		Exists: true,
		ID:     cat.ID,
		Name:   cat.Name,
	}, nil
}

type CreateCategoryInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateCategoryOutput struct {
	Status string `json:"status"`
	ID     string `json:"id"`
}

func (t *ContentTools) CreateCategory(ctx context.Context, input CreateCategoryInput) (*CreateCategoryOutput, error) {
	req := dto.CreateCategoryRequest{
		Name:        input.Name,
		Description: input.Description,
	}
	cat, err := t.client.CreateCategory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &CreateCategoryOutput{
		Status: "created",
		ID:     cat.ID,
	}, nil
}

type CreateMovieInput struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	CategoryName string `json:"category_name"`
	Type         string `json:"type"`
	Duration     int    `json:"duration"`
}

type CreateMovieOutput struct {
	Status string `json:"status"`
	ID     string `json:"id"`
}

func (t *ContentTools) CreateMovie(ctx context.Context, input CreateMovieInput) (*CreateMovieOutput, error) {
	cat, err := t.client.GetCategoryByName(ctx, input.CategoryName)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if cat == nil {
		return nil, fmt.Errorf("category '%s' does not exist", input.CategoryName)
	}

	req := dto.CreateMediaContentRequest{
		Name:        input.Title,
		Description: input.Description,
		Type:        strings.ToLower(input.Type),
		Duration:    input.Duration,
		Categories:  []string{cat.ID},
	}

	id, err := t.client.CreateMediaContent(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create media content: %w", err)
	}

	return &CreateMovieOutput{
		Status: "created",
		ID:     id,
	}, nil
}
