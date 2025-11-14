package rest

import (
	"soa-video-streaming/services/user-service/internal/controller/rest"
	"soa-video-streaming/services/user-service/internal/controller/rest/middleware"
	"soa-video-streaming/services/user-service/internal/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
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
	authService *service.AuthService,
) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")

	pubAuth := v1.Group("/auth", middleware.JWTAuthMiddleware(authService))
	{
		pubAuth.POST("/sign-in", auth.SignIn)
		pubAuth.POST("/sign-up", auth.SignUp)
	}

	privateUsers := v1.Group("/users", middleware.JWTAuthMiddleware(authService))
	{
		privateUsers.POST("/preferences/categories", users.AddPreferenceCategories)
	}

	return r
}
