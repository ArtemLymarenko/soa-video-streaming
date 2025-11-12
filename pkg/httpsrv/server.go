package httpsrv

import (
	"context"
	"net/http"
	"soa-video-streaming/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type ServerOptions struct {
	fx.In
	Lc  fx.Lifecycle
	Cfg *config.BaseHTTPServerConfig
	Eng *gin.Engine
}

func NewHTTPServer(o ServerOptions) *http.Server {
	srv := &http.Server{
		Addr:    o.Cfg.HTTP.Addr,
		Handler: o.Eng,
	}

	o.Lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logrus.Infof("Starting HTTP server on %s", srv.Addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logrus.WithError(err).Error("HTTP server error")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logrus.Infof("Stopping HTTP server...")
			return srv.Shutdown(ctx)
		},
	})

	return srv
}

func Module() fx.Option {
	return fx.Provide(
		NewHTTPServer,
	)
}

func Invoke(s *http.Server) {}
