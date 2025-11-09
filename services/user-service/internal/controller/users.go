package controller

import (
	"net/http"
	"soa-video-streaming/services/user-service/internal/config"

	"github.com/gin-gonic/gin"
)

type Users struct {
	cfg *config.AppConfig
}

func NewUsers(cfg *config.AppConfig) *Users {
	return &Users{
		cfg: cfg,
	}
}

func (s *Users) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"addr": s.cfg.HTTP.Addr,
	})
}
