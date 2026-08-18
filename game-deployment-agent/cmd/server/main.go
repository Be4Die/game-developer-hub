package main

import (
	"context"
	"log"

	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/app"
	"github.com/Be4Die/game-developer-hub/game-deployment-agent/internal/infrastructure/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	ctx := context.Background()
	if err := application.Run(ctx); err != nil {
		log.Fatalf("server terminated with error: %v", err)
	}
}
