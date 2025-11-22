package rest

import (
	"context"
	"net/http"
	"soa-video-streaming/services/content-service/internal/controller/rest/dto"
	"soa-video-streaming/services/content-service/internal/domain/entity"
	"time"

	"github.com/gin-gonic/gin"
)

type MediaContentService interface {
	Create(ctx context.Context, m entity.MediaContent) error
	GetByID(ctx context.Context, id string) (*entity.MediaContent, error)
	Delete(ctx context.Context, id string) error
	GetRecommendations(ctx context.Context, userID string, limit int64) ([]entity.MediaContent, error)
}

type MediaContentController struct {
	service MediaContentService
}

func NewMediaContentController(service MediaContentService) *MediaContentController {
	return &MediaContentController{service: service}
}

func (c *MediaContentController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", c.Create)
	rg.GET("/:id", c.GetByID)
	rg.DELETE("/:id", c.Delete)
	rg.GET("/recommendations/:userId", c.GetRecommendations)
}

func (c *MediaContentController) Create(ctx *gin.Context) {
	var req dto.CreateMediaContentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	categories := make([]entity.Category, len(req.Categories))
	for i, catID := range req.Categories {
		categories[i] = entity.Category{ID: entity.CategoryID(catID)}
	}

	media := entity.MediaContent{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Type:        entity.MediaContentType(req.Type),
		Duration:    req.Duration,
		Categories:  categories,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := c.service.Create(ctx, media); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusCreated)
}

func (c *MediaContentController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	media, err := c.service.GetByID(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if media == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, dto.ToMediaContentResponse(media))
}

func (c *MediaContentController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.service.Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

const DefaultRecommendationsLimit = 10

func (c *MediaContentController) GetRecommendations(ctx *gin.Context) {
	userID := ctx.Param("userId")
	limit := ctx.GetInt64("limit")
	if limit == 0 {
		limit = DefaultRecommendationsLimit
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
