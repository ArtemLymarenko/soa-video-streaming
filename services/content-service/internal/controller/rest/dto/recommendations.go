package dto

import "soa-video-streaming/services/content-service/internal/domain/entity"

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
