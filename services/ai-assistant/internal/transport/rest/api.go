package rest

import (
	"go.uber.org/fx"
	"soa-video-streaming/services/ai-assistant/internal/controller/rest"

	"github.com/gin-gonic/gin"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewGinEngine,
		),
	)
}

func NewGinEngine(
	ai *rest.AIController,
) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")

	ai.RegisterRoutes(v1)

	return r
}
