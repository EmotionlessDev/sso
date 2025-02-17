package grpcapp

import (
	"fmt"
	"log/slog"
	"net"

	authgrpc "github.com/EmotionlessDev/sso/internal/grpc/auth"
	"google.golang.org/grpc"
)

// grpcapp.App is a struct that represents the gRPC server application

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, port int, authService authgrpc.Auth) *App {
	gRPCServer := grpc.NewServer()
	authgrpc.Register(gRPCServer, authService)

	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.App.Run"
	log := a.log.With(slog.String("op", op))
	log.Info("Starting gRPC server", slog.Int("port", a.port))

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	log.Info("grpc server is running", slog.String("address", l.Addr().String()))
	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.App.Stop"

	log := a.log.With(slog.String("op", op))
	log.Info("Stopping gRPC server")

	a.gRPCServer.GracefulStop()
}
