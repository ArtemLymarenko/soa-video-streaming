package rest

import (
	"soa-video-streaming/services/user-service/internal/controller/rest"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			rest.NewUsersController,
			NewGinEngine,
		),
	)
}

const (
	SignInEndpoint = "/signin"
	SignUpEndpoint = "/signup"
)

func NewGinEngine(users *rest.UsersController) *gin.Engine {
	r := gin.Default()

	r.POST(SignInEndpoint, users.SignIn)
	r.POST(SignUpEndpoint, users.SignUp)

	return r
}
