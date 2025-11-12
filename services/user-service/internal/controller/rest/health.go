package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewGinEngine,
		),
	)
}

const (
	HealthRoute = "/health"
)

func NewGinEngine() *gin.Engine {
	r := gin.Default()

	r.GET(HealthRoute, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	return r
}
