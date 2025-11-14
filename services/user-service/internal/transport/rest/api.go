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
	GetUserByID = "/users/:userID"
)

type EngineController struct {
	fx.In

	Users *rest.UsersController
}

func NewGinEngine(e EngineController) *gin.Engine {
	r := gin.Default()

	r.GET(GetUserByID, e.Users.GetUserByID)

	return r
}
