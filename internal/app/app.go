package app

import (
	"context"
	"log/slog"
	"net"
	"os"

	appgrpc "github.com/al-revenko/idp/internal/app/grpc"
	"github.com/al-revenko/idp/internal/app/grpc/interceptor"
	appservice "github.com/al-revenko/idp/internal/app/service"
	clientgrpc "github.com/al-revenko/idp/internal/client/grpc"
	clientservice "github.com/al-revenko/idp/internal/client/service"
	clientstore "github.com/al-revenko/idp/internal/client/store"
	"github.com/al-revenko/idp/internal/lib/config"
	"github.com/al-revenko/idp/internal/lib/crypt"
	"github.com/al-revenko/idp/internal/lib/crypt/hash"
	"github.com/al-revenko/idp/internal/lib/crypt/rsa"
	"github.com/al-revenko/idp/internal/lib/crypt/signer"
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
	op := pkg.Op("Start")
	log := a.log.With(slog.String("op", op.Name))

	if a.ctx.Err() != nil {
		return op.Err(a.ctx.Err())
	}

	signer, hasher, err := a.cryptInit()
	if err != nil {
		return op.Err(err)
	}

	dbconn, err := a.dbInit()
	if err != nil {
		return op.Err(err)
	}
	a.dbconn = dbconn

	a.servicesInit(dbconn, signer, hasher)

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
	op := pkg.Op("Stop")
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

func (a *App) cryptInit() (*signer.Signer, *hash.Argon2, error) {
	op := pkg.Op("cryptInit")

	privRsa, pubRsa, err := rsa.DecodeBase64Keys(a.config.Secrets.RSAPrivateBase64, a.config.Secrets.RSAPublicBase64)
	if err != nil {
		return nil, nil, op.Err(err)
	}

	signer := signer.New(privRsa, pubRsa)

	hasher := &hash.Argon2{
		Memory:      uint32(a.config.Hash.Memory),
		Iterations:  uint32(a.config.Hash.Iterations),
		Parallelism: uint8(a.config.Hash.Parallelism),
		KeyLength:   a.config.Hash.KeyLength,
		SaltGenerator: func() ([]byte, error) {
			return crypt.GenerateRandomBytes(a.config.Hash.SaltLength)
		},
	}

	return signer, hasher, nil
}

func (a *App) dbInit() (*pgx.Conn, error) {
	op := pkg.Op("dbInit")

	dbconn, err := pgx.Connect(a.ctx, a.config.Secrets.DBString)
	if err != nil {
		return nil, op.Err(err)
	}

	return dbconn, nil
}

func (a *App) servicesInit(conn *pgx.Conn, signer *signer.Signer, hasher *hash.Argon2) {
	appService := appservice.New(signer)
	appgrpc.Register(a.grpcServer, appService)

	clientStore := clientstore.New(conn)
	clientService := clientservice.New(clientStore, hasher)
	clientgrpc.Register(a.grpcServer, clientService)
}
