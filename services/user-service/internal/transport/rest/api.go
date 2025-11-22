package rest

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"soa-video-streaming/pkg/middleware"
	"soa-video-streaming/services/user-service/internal/controller/rest"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			rest.NewAuthController,
			rest.NewUsersController,
			NewGinEngine,
		),
	)
}

func NewGinEngine(
	auth *rest.AuthController,
	users *rest.UsersController,
) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")

	pubAuth := v1.Group("/auth")
	{
		pubAuth.POST("/sign-in", auth.SignIn)
		pubAuth.POST("/sign-up", auth.SignUp)
	}

	privateUsers := v1.Group("/users", middleware.Auth())
	{
		privateUsers.POST("/preferences/categories", users.AddPreferenceCategories)
	}

	return r
}
