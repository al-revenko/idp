package app

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/al-revenko/idp/internal/lib/config"
	"github.com/al-revenko/idp/internal/lib/sign"
	"github.com/redis/go-redis/v9"

	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
)

var pkg = sign.Pkg("app")

type Addr = string

type App struct {
	ctx        context.Context
	cancelCtx  context.CancelFunc
	config     config.Config
	log        *slog.Logger
	grpcServer *grpc.Server
	httpServer *http.Server
	dbconn     *pgx.Conn
	redisConn  *redis.Client
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) *App {
	appCtx, cancel := context.WithCancel(ctx)

	if log == nil {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return &App{
		ctx:       appCtx,
		cancelCtx: cancel,
		config:    cfg,
		log:       log,
	}
}

func (a *App) Start() error {
	op := pkg.Op("App.Start")

	g, ctx := errgroup.WithContext(a.ctx)
	a.ctx = ctx

	log := a.log.With(slog.String("op", op.Name))

	if a.ctx.Err() != nil {
		return op.Err(a.ctx.Err())
	}

	httpServer, grpcServer, dbconn, redisConn, err := initApp(a.ctx, &a.config, a.log)
	if err != nil {
		return op.Err(err)
	}

	a.dbconn = dbconn
	a.redisConn = redisConn
	a.grpcServer = grpcServer
	a.httpServer = &http.Server{
		Addr:    a.config.HTTP.Addr,
		Handler: httpServer,
	}

	g.Go(func() error {
		netl, err := net.Listen("tcp", a.config.GRPC.Addr)
		if err != nil {
			return op.Err(err)
		}

		log.Info("grpc listening", slog.String("addr", netl.Addr().String()))

		if err := a.grpcServer.Serve(netl); err != nil {
			return op.Err(fmt.Errorf("grpc goroutine: %w", err))
		}

		return nil
	})

	g.Go(func() error {
		netl, err := net.Listen("tcp", a.config.HTTP.Addr)
		if err != nil {
			return op.Err(err)
		}

		log.Info("http listening", slog.String("addr", netl.Addr().String()))

		if err := a.httpServer.Serve(netl); err != nil {
			if err == http.ErrServerClosed {
				return nil
			}

			return op.Err(fmt.Errorf("http goroutine: %w", err))
		}

		return nil
	})

	return g.Wait()
}

func (a *App) MustStart() {
	if err := a.Start(); err != nil {
		panic(err)
	}
}

func (a *App) Stop() error {
	op := pkg.Op("App.Stop")
	log := a.log.With(slog.String("op", op.Name))

	log.Info("stopping grpc server")

	a.grpcServer.GracefulStop()

	log.Info("stopping http server")

	a.httpServer.Shutdown(a.ctx)

	if a.dbconn != nil {
		log.Info("close db connection")
		if err := a.dbconn.Close(a.ctx); err != nil {
			log.Error("failed to close db connection", slog.String("error", op.Err(err).Error()))
		}
	}

	if a.redisConn != nil {
		log.Info("close redis connection")
		if err := a.redisConn.Close(); err != nil {
			log.Error("failed to close redis connection", slog.String("error", op.Err(err).Error()))
		}
	}

	a.cancelCtx()

	return nil
}
