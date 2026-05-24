package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Be4Die/game-developer-hub/project-manager/internal/app"
	"github.com/Be4Die/game-developer-hub/project-manager/internal/infrastructure/config"
)

func main() {
	cfgPath := os.Getenv("CONFIG_PATH")
	if cfgPath == "" {
		cfgPath = "config/local.yaml"
	}

	cfg := config.MustLoad(cfgPath)
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	application, err := app.New(log, cfg)
	if err != nil {
		log.Error("failed to initialize app", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Graceful shutdown.
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		application.MustStop()
	}()

	application.MustRun()
}
