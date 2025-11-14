package rest

import (
	"soa-video-streaming/services/user-service/internal/service"

	"github.com/gin-gonic/gin"
)

type UsersController struct {
	authService *service.AuthService
}

func NewUsersController(authService *service.AuthService) *UsersController {
	return &UsersController{
		authService: authService,
	}
}

func (c *UsersController) SignUp(gc *gin.Context) {

}

func (c *UsersController) SignIn(gc *gin.Context) {

}
