package dto

import "soa-video-streaming/services/content-service/internal/domain/entity"

type CreateMediaContentRequest struct {
	ID          string   `json:"id" binding:"required"`
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	Type        string   `json:"type" binding:"required,oneof=movie series"`
	Duration    int      `json:"duration"`
	Categories  []string `json:"categories"` // List of Category IDs
}

type MediaContentResponse struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Type        string             `json:"type"`
	Duration    int                `json:"duration"`
	Categories  []CategoryResponse `json:"categories"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
}

type CategoryResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ToMediaContentResponse(m *entity.MediaContent) MediaContentResponse {
	cats := make([]CategoryResponse, len(m.Categories))
	for i, c := range m.Categories {
		cats[i] = CategoryResponse{
			ID:          string(c.ID),
			Name:        c.Name,
			Description: c.Description,
		}
	}

	return MediaContentResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Type:        string(m.Type),
		Duration:    m.Duration,
		Categories:  cats,
		CreatedAt:   m.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   m.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

type RecommendationResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Duration    int    `json:"duration"`
}

func ToRecommendationResponse(m *entity.MediaContent) RecommendationResponse {
	return RecommendationResponse{
		ID:          m.ID,
		Name:        m.Name,
		Description: m.Description,
		Type:        string(m.Type),
		Duration:    m.Duration,
	}
}

func ToRecommendationResponses(mediaList []entity.MediaContent) []RecommendationResponse {
	recommendations := make([]RecommendationResponse, len(mediaList))
	for i, m := range mediaList {
		recommendations[i] = ToRecommendationResponse(&m)
	}
	return recommendations
}
