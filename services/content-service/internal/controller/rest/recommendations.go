package rest

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"soa-video-streaming/services/content-service/internal/controller/rest/dto"
	"soa-video-streaming/services/content-service/internal/service"
)

type RecommendationsController struct {
	service *service.Recommendations
}

func NewRecommendationsController(service *service.Recommendations) *RecommendationsController {
	return &RecommendationsController{service: service}
}

func (c *RecommendationsController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", c.GetRecommendations)
}

const DefaultRecommendationsLimit = 10

func (c *RecommendationsController) GetRecommendations(ctx *gin.Context) {
	limit := ctx.GetInt64("limit")
	if limit == 0 {
		limit = DefaultRecommendationsLimit
	}

	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	recommendations, err := c.service.GetRecommendations(ctx, userID, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if recommendations == nil {
		ctx.JSON(http.StatusOK, []dto.RecommendationResponse{})
		return
	}

	ctx.JSON(http.StatusOK, dto.ToRecommendationResponses(recommendations))
}
