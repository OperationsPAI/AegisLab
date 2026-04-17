package httpinterface

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type ServerConfig struct {
	Addr string
}

func NewServer(config ServerConfig, engine *gin.Engine) *http.Server {
	return &http.Server{
		Addr:    config.Addr,
		Handler: engine,
	}
}

func registerServerLifecycle(lc fx.Lifecycle, server *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				logrus.Infof("Starting HTTP server on %s", server.Addr)
				if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					logrus.Errorf("HTTP server error: %v", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logrus.Info("Stopping HTTP server")
			return server.Shutdown(ctx)
		},
	})
}
