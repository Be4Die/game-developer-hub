// SSO-сервис — Single Sign-On для платформы Game Developer Hub.
package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Be4Die/game-developer-hub/sso/internal/app"
	"github.com/Be4Die/game-developer-hub/sso/internal/infrastructure/config"

	// Регистрируем gzip decompressor для gRPC.
	_ "google.golang.org/grpc/encoding/gzip"
)

func main() {
	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	application, err := app.New(log, cfg)
	if err != nil {
		log.Error("failed to initialize application", slog.String("error", err.Error()))
		os.Exit(1)
	}

	go func() { application.MustRun() }()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	application.MustStop()
}

func setupLogger(env string) *slog.Logger {
	var level slog.Level
	switch env {
	case config.EnvLocal:
		level = slog.LevelDebug
	case config.EnvDev:
		level = slog.LevelInfo
	case config.EnvProd:
		level = slog.LevelWarn
	default:
		level = slog.LevelInfo
	}

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
}
