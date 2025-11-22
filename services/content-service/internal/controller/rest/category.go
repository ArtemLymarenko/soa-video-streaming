package rest

import (
	"context"
	"net/http"
	"soa-video-streaming/services/content-service/internal/controller/rest/dto"
	"soa-video-streaming/services/content-service/internal/domain/entity"

	"github.com/gin-gonic/gin"
)

type CategoryService interface {
	Create(ctx context.Context, c entity.Category) error
	GetByID(ctx context.Context, id entity.CategoryID) (*entity.Category, error)
	Update(ctx context.Context, c entity.Category) error
	Delete(ctx context.Context, id entity.CategoryID) error
	GetByTimestamp(ctx context.Context, from, to int64) ([]entity.Category, error)
}

type CategoryController struct {
	service CategoryService
}

func NewCategoryController(service CategoryService) *CategoryController {
	return &CategoryController{service: service}
}

func (c *CategoryController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", c.Create)
	rg.GET("/:id", c.GetByID)
	rg.PUT("/:id", c.Update)
	rg.DELETE("/:id", c.Delete)
	rg.GET("", c.GetByTimestamp)
}

func (c *CategoryController) Create(ctx *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := entity.Category{
		ID:          entity.CategoryID(req.ID),
		Name:        req.Name,
		Description: req.Description,
	}

	if err := c.service.Create(ctx, category); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusCreated)
}

func (c *CategoryController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")
	category, err := c.service.GetByID(ctx, entity.CategoryID(id))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if category == nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, category)
}

func (c *CategoryController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var req dto.UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category := entity.Category{
		ID:          entity.CategoryID(id),
		Name:        req.Name,
		Description: req.Description,
	}

	if err := c.service.Update(ctx, category); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusOK)
}

func (c *CategoryController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.service.Delete(ctx, entity.CategoryID(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *CategoryController) GetByTimestamp(ctx *gin.Context) {
	from := ctx.GetInt64("from")
	to := ctx.GetInt64("to")

	categories, err := c.service.GetByTimestamp(ctx, from, to)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, categories)
}
