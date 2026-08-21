package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/al-revenko/idp/internal/app"
	"github.com/al-revenko/idp/internal/lib/config"
	"github.com/lmittmann/tint"
)

func main() {
	cfg := config.MustLoad()

	var log *slog.Logger

	switch cfg.Env {
	case config.EnvLocal:
		log = slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))
	case config.EnvDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	default:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	application := app.New(context.Background(), cfg, log)

	log.Info("application starting", slog.String("op", "main"), slog.String("env", cfg.Env))

	go application.MustStart()

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGTERM, syscall.SIGINT)

	<-stopSignal

	err := application.Stop()
	if err != nil {
		log.Error("failed to gracefully stop application", slog.Any("error", err))
		os.Exit(0)
	}
}
