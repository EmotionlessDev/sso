package app

import (
	"log/slog"
	"time"

	grpcapp "github.com/EmotionlessDev/sso/internal/app/grpc"
	"github.com/EmotionlessDev/sso/internal/services/auth"
	"github.com/EmotionlessDev/sso/internal/storage/postgres"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, tokenTTL time.Duration, dsn string) *App {
	storage, err := postgres.New(dsn)
	if err != nil {
		panic(err)
	}

	authService := auth.New(log, storage, storage, storage, tokenTTL)
	grpcApp := grpcapp.New(log, grpcPort, authService)

	return &App{
		GRPCServer: grpcApp,
	}
}
