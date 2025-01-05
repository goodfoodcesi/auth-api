package main

import (
	"context"
	"io"
	"net/http"

	"github.com/goodfoodcesi/auth-api/config"
	"github.com/goodfoodcesi/auth-api/logging"
	"github.com/goodfoodcesi/auth-api/server"
	"go.uber.org/zap"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	logManager := logging.NewLogManager(cfg.Env)
	loggerFromCtx := logging.FromContext(context.Background())

	if cfg.Env != "dev" {
		tracer.Start(
			tracer.WithService("auth-api"),
			tracer.WithEnv(cfg.Env),
			tracer.WithServiceVersion("0.0.1"),
		)
		defer tracer.Stop()
		gin.DefaultWriter = io.Discard
	}

	api := server.SetupApi(cfg, logManager)

	loggerFromCtx.Info("Starting server on port 8080", zap.String("env", cfg.Env))
	err := http.ListenAndServe(":8080", api)
	if err != nil {
		loggerFromCtx.Fatal("Failed to start server", zap.Error(err))
	}
}
