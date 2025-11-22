package rest

import (
	"soa-video-streaming/services/content-service/internal/controller/rest"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			rest.NewCategoryController,
			rest.NewMediaContentController,
			NewGinEngine,
		),
	)
}

func NewGinEngine(
	category *rest.CategoryController,
	media *rest.MediaContentController,
) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")

	category.RegisterRoutes(v1.Group("/categories"))
	media.RegisterRoutes(v1.Group("/media-content"))

	return r
}
