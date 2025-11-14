package rest

import (
	"errors"
	"net/http"
	"soa-video-streaming/services/user-service/internal/controller/rest/dto"
	"soa-video-streaming/services/user-service/internal/service"

	"github.com/gin-gonic/gin"
)

type UsersController struct {
	usersService *service.UsersService
}

func NewUsersController(u *service.UsersService) *UsersController {
	return &UsersController{
		usersService: u,
	}
}

func (c *UsersController) AddPreferenceCategories(gc *gin.Context) {
	var req dto.AddPreferenceCategoriesRequest

	if err := gc.ShouldBindJSON(&req); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}

	if err := dto.GetValidator().Struct(req); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	userID := gc.GetString("user_id")
	if userID == "" {
		gc.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := c.usersService.AddPreferenceCategories(gc, userID, req.CategoryIDs); err != nil {
		if errors.Is(err, service.ErrCategoryNotFound) {
			gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		gc.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	gc.JSON(http.StatusOK, dto.AddPreferenceCategoriesResponse{
		Status: "ok",
	})
}
