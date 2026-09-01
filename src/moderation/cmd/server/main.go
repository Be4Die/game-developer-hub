// main.go — точка входа сервиса moderation.
package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Be4Die/game-developer-hub/moderation/internal/app"
	"github.com/Be4Die/game-developer-hub/moderation/internal/infrastructure/config"
)

func main() {
	cfg := config.MustLoad()

	var handler slog.Handler
	if cfg.Env == config.EnvLocal {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}
	logger := slog.New(handler)
	slog.SetDefault(logger)

	logger.Info("starting moderation service", slog.String("env", cfg.Env), slog.Int("port", cfg.GRPC.Port))

	application, err := app.New(logger, cfg)
	if err != nil {
		logger.Error("failed to initialize moderation application", slog.String("error", err.Error()))
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go application.MustRun()

	sig := <-stop
	logger.Info("received termination signal", slog.String("signal", sig.String()))

	application.MustStop()
}
