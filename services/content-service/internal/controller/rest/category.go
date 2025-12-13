package rest

import (
	"net/http"
	"soa-video-streaming/services/content-service/internal/controller/rest/dto"
	"soa-video-streaming/services/content-service/internal/domain/entity"
	"soa-video-streaming/services/content-service/internal/service"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

type CategoryController struct {
	service *service.CategoryService
}

func NewCategoryController(service *service.CategoryService) *CategoryController {
	return &CategoryController{service: service}
}

func (c *CategoryController) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("", c.Create)
	rg.GET("/:id", c.GetByID)
	rg.PUT("/:id", c.Update)
	rg.DELETE("/:id", c.Delete)
	rg.GET("", c.GetByTimestamp)
	rg.GET("/all", c.GetAll)
}

func (c *CategoryController) GetAll(ctx *gin.Context) {
	categories, err := c.service.GetAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, categories)
}

func (c *CategoryController) Create(ctx *gin.Context) {
	var req dto.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := entity.CategoryID(uuid.NewString())

	category := entity.Category{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := c.service.Create(ctx, category); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, category)
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

	ctx.Status(http.StatusOK)
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
