package httpsrv

import (
	"context"
	"net/http"
	"soa-video-streaming/pkg/config"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Options(
		fx.Provide(
			NewHTTPServer,
		),
		fx.Invoke(runServer),
	)
}

func NewHTTPServer(cfg *config.BaseHTTPServerConfig, eng *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    cfg.HTTP.Addr,
		Handler: eng,
	}
}

func runServer(lc fx.Lifecycle, srv *http.Server) {
	lc.Append(fx.Hook{
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
}
