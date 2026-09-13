package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Be4Die/game-developer-hub/ingress-proxy/internal/config"
	"github.com/Be4Die/game-developer-hub/ingress-proxy/internal/proxy"
	"github.com/Be4Die/game-developer-hub/ingress-proxy/internal/store"
)

func main() {
	cfg := config.Load()

	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(log)

	log.Info("starting gdh-ingress-proxy",
		slog.Int("port", cfg.Port),
		slog.String("valkey_addr", cfg.ValkeyAddr),
		slog.Int("valkey_db", cfg.ValkeyDB),
	)

	valkeyStore := store.NewValkeyStore(cfg.ValkeyAddr, cfg.ValkeyPassword, cfg.ValkeyDB)
	defer valkeyStore.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	if err := valkeyStore.Ping(ctx); err != nil {
		log.Warn("failed to ping valkey on startup, will retry on requests", slog.Any("err", err))
	} else {
		log.Info("connected to valkey route store")
	}
	cancel()

	handler := proxy.NewHandler(valkeyStore, log, cfg.DialTimeout)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server failed", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	log.Info("gdh-ingress-proxy is listening", slog.String("addr", server.Addr))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Info("shutting down gdh-ingress-proxy gracefully...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown failed", slog.Any("err", err))
	}
	log.Info("gdh-ingress-proxy stopped")
}
