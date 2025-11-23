package rest

import (
	"github.com/google/uuid"
	"net/http"
	"soa-video-streaming/services/content-service/internal/controller/rest/dto"
	"soa-video-streaming/services/content-service/internal/domain/entity"
	"soa-video-streaming/services/content-service/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

type MediaContentController struct {
	service *service.MediaContentService
}

func NewMediaContentController(service *service.MediaContentService) *MediaContentController {
	return &MediaContentController{service: service}
}

func (c *MediaContentController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", c.Create)
	rg.GET("/:id", c.GetByID)
	rg.DELETE("/:id", c.Delete)
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

	id := uuid.NewString()
	media := entity.MediaContent{
		ID:          id,
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

	ctx.JSON(http.StatusCreated, gin.H{"id": id})
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
