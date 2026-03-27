package app

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/Be4Die/game-developer-hub/game-server-node/internal/config"
	"google.golang.org/grpc"
)

type App struct {
	log        *slog.Logger
	config     *config.Config
	gRPCServer *grpc.Server
}

func New(log *slog.Logger, config *config.Config) *App {
	return &App{
		log:        log,
		config:     config,
		gRPCServer: grpc.NewServer(),
	}
}

func (a *App) MustRun() {
	if err := a.runGRPCServer(); err != nil {
		panic(err)
	}
}

func (a *App) runGRPCServer() error {
	const op = "App.RunGRPCServer"
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.config.GRPC.Port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	a.log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) MustStop() {
	const op = "App.MustStop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.config.GRPC.Port))

	a.gRPCServer.GracefulStop()
}
