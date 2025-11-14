package rest

import (
	"net/http"
	"soa-video-streaming/pkg/cookie"
	"soa-video-streaming/services/user-service/internal/config"
	"soa-video-streaming/services/user-service/internal/controller/rest/dto"
	"soa-video-streaming/services/user-service/internal/domain/entity"
	"soa-video-streaming/services/user-service/internal/service"

	"github.com/gin-gonic/gin"
)

type UsersController struct {
	authService *service.AuthService
	cfg         *config.AppConfig
}

func NewUsersController(authService *service.AuthService, cfg *config.AppConfig) *UsersController {
	return &UsersController{
		authService: authService,
		cfg:         cfg,
	}
}

func (c *UsersController) SignUp(gc *gin.Context) {
	var req dto.SignUpRequest
	if err := gc.ShouldBindJSON(&req); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user := entity.User{
		Email:    req.Email,
		Password: req.Password,
		UserInfo: entity.UserInfo{
			FirstName: req.FirstName,
			LastName:  req.LastName,
		},
	}

	authRes, err := c.authService.SignUp(gc, user)
	if err != nil {
		gc.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	cookie.SetAccessToken(gc, authRes.AccessToken, c.cfg.Auth.JwtTTL)

	gc.JSON(http.StatusOK, dto.SignUpResponse{
		AccessToken: authRes.AccessToken,
	})
}

func (c *UsersController) SignIn(gc *gin.Context) {
	var req dto.SignInRequest
	if err := gc.ShouldBindJSON(&req); err != nil {
		gc.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	authRes, err := c.authService.SignIn(gc, req.Email, req.Password)
	if err != nil {
		gc.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	cookie.SetAccessToken(gc, authRes.AccessToken, c.cfg.Auth.JwtTTL)

	gc.JSON(http.StatusOK, dto.SignInResponse{
		AccessToken: authRes.AccessToken,
	})
}
