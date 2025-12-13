package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"soa-video-streaming/services/content-service/internal/controller/rest/dto"
	"soa-video-streaming/services/content-service/internal/domain/entity"
	"time"
)

type ContentServiceClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewContentServiceClient(baseURL string) *ContentServiceClient {
	return &ContentServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ContentServiceClient) GetAllCategories(ctx context.Context) ([]entity.Category, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/categories", c.baseURL), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	var categories []entity.Category
	if err := json.Unmarshal(body, &categories); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return categories, nil
}

func (c *ContentServiceClient) CreateMediaContent(ctx context.Context, req dto.CreateMediaContentRequest) (string, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/api/v1/media-content", c.baseURL), bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, string(body))
	}

	var response struct {
		ID string `json:"id"`
	}

	if err = json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return response.ID, nil
}
