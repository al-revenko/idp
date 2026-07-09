package app

import (
	"context"
	"log/slog"
	"os"

	"github.com/al-revenko/sso-service/internal/lib/config"
)

type Addr = string

type App struct {
	ctx    context.Context
	cancel context.CancelFunc
	config config.Config
	log    *slog.Logger
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) *App {
	appCtx, cancel := context.WithCancel(ctx)

	if log == nil {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return &App{
		ctx:    appCtx,
		cancel: cancel,
		config: cfg,
		log:    log,
	}
}

func (a *App) Start() error {
	if a.ctx.Err() != nil {
		a.cancel()
		return a.ctx.Err()
	}

	return nil
}

func (a *App) MustStart() {
	if err := a.Start(); err != nil {
		panic(err)
	}
}

func (a *App) Stop() error {
	a.cancel()

	return nil
}
