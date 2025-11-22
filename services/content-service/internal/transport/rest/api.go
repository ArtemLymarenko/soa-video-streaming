package rest

import (
	"soa-video-streaming/pkg/middleware"
	"soa-video-streaming/services/content-service/internal/controller/rest"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			rest.NewCategoryController,
			rest.NewMediaContentController,
			rest.NewRecommendationsController,
			NewGinEngine,
		),
	)
}

func NewGinEngine(
	category *rest.CategoryController,
	media *rest.MediaContentController,
	recommendations *rest.RecommendationsController,
) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1", middleware.Auth())

	category.RegisterRoutes(v1.Group("/categories"))
	media.RegisterRoutes(v1.Group("/media-content"))
	recommendations.RegisterRoutes(v1.Group("/recommendations"))

	return r
}
