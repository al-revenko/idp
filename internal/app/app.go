package app

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/al-revenko/idp/internal/app/grpc/interceptor"
	"github.com/al-revenko/idp/internal/lib/config"
	"github.com/al-revenko/idp/internal/lib/pkgmark"

	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
)

var pkg = pkgmark.New("app")

type Addr = string

type App struct {
	ctx        context.Context
	cancelCtx  context.CancelFunc
	config     config.Config
	log        *slog.Logger
	grpcServer *grpc.Server
	dbconn     *pgx.Conn
}

func New(ctx context.Context, cfg config.Config, log *slog.Logger) *App {
	appCtx, cancel := context.WithCancel(ctx)

	if log == nil {
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptor.TraceInterceptor(log), interceptor.ErrorInterceptor(log)),
	)

	return &App{
		ctx:        appCtx,
		cancelCtx:  cancel,
		config:     cfg,
		grpcServer: grpcServer,
		log:        log,
		dbconn:     nil,
	}
}

func (a *App) Start() error {
	op := pkg.Op("App.Start")
	log := a.log.With(slog.String("op", op.Name))

	if a.ctx.Err() != nil {
		return op.Err(a.ctx.Err())
	}

	err := a.init()
	if err != nil {
		return op.Err(err)
	}

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

	return nil
}
