package app

import (
	"context"
	"log/slog"
	"net"
	"os"

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
	dbconn     *pgx.Conn
	redisConn  *redis.Client
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) *App {
	appCtx, cancel := context.WithCancel(ctx)

	if log == nil {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return &App{
		ctx:        appCtx,
		cancelCtx:  cancel,
		config:     cfg,
		log:        log,
		grpcServer: nil,
		dbconn:     nil,
		redisConn:  nil,
	}
}

func (a *App) Start() error {
	op := pkg.Op("App.Start")
	log := a.log.With(slog.String("op", op.Name))

	if a.ctx.Err() != nil {
		return op.Err(a.ctx.Err())
	}

	grpcServer, dbconn, redisConn, err := initApp(a.ctx, &a.config, a.grpcServer, a.log)
	if err != nil {
		return op.Err(err)
	}
	a.grpcServer = grpcServer
	a.dbconn = dbconn
	a.redisConn = redisConn

	netl, err := net.Listen("tcp", a.config.GRPC.Addr)
	if err != nil {
		return op.Err(err)
	}

	log.Info("grpc listening", slog.String("addr", netl.Addr().String()))

	if err := a.grpcServer.Serve(netl); err != nil {
		return op.Err(err)
	}

	return nil
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

	a.cancelCtx()

	if a.dbconn != nil {
		log.Info("close db connection")
		if err := a.dbconn.Close(a.ctx); err != nil {
			log.Error("failed to close db connection", op.Err(err))
		}
	}

	if a.redisConn != nil {
		log.Info("close redis connection")
		if err := a.redisConn.Close(); err != nil {
			log.Error("failed to close redis connection", op.Err(err))
		}
	}

	return nil
}
