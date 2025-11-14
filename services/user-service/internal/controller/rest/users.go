package rest

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type UsersController struct{}

func NewUsersController() *UsersController {
	return &UsersController{}
}

func (c *UsersController) GetUserByID(gc *gin.Context) {
	gc.JSON(http.StatusOK, gin.H{
		"userID": "ok",
	})
}
