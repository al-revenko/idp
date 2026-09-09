package app

import (
	"context"
	"log/slog"

	"github.com/al-revenko/idp/db/sql/gen/dbstore"
	"github.com/al-revenko/idp/internal/app/grpc/interceptor"
	"github.com/al-revenko/idp/internal/domain/auth"
	"github.com/al-revenko/idp/internal/domain/client"
	"github.com/al-revenko/idp/internal/domain/user"
	"github.com/al-revenko/idp/internal/lib/config"
	"github.com/al-revenko/idp/internal/lib/crypt"
	"github.com/al-revenko/idp/internal/lib/crypt/hash/argon2"
	"github.com/al-revenko/idp/internal/lib/crypt/hash/blake3"
	"github.com/al-revenko/idp/internal/lib/crypt/rsa"
	"github.com/al-revenko/idp/internal/lib/valid"
	"github.com/jackc/pgx/v5"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

func initApp(ctx context.Context, cfg *config.Config, grpcServer *grpc.Server, log *slog.Logger) (*grpc.Server, *pgx.Conn, *redis.Client, error) {
	op := pkg.Op("App.init")

	validator := valid.New()

	rsaManager, argon2Hasher, blake3Hasher, err := cryptInit(cfg)
	if err != nil {
		return nil, nil, nil, op.Err(err)
	}

	dbconn, redisConn, err := dbInit(ctx, cfg, log)
	if err != nil {
		return nil, nil, nil, op.Err(err)
	}
	db := dbstore.New(dbconn)

	clientStore := client.NewStore(db)
	clientService := client.NewService(clientStore, argon2Hasher)

	userStore := user.NewStore(db)
	userService := user.NewService(userStore, argon2Hasher)

	authTokenParams := &auth.TokenParams{
		Signer:                      rsaManager,
		Hash:                        blake3Hasher,
		Issuer:                      cfg.AppName,
		AccessTokenTTL:              cfg.Auth.AccessTokenTTL,
		RefreshTokenTTL:             cfg.Auth.RefreshTokenTTL,
		RefreshTokenRevokedStoreTTL: cfg.Auth.RefreshTokenRevokedStoreTTL,
		RefreshTokenGracePeriod:     cfg.Auth.RefreshTokenGracePeriod,
	}
	authStore := auth.NewStore(redisConn)
	authService := auth.NewService(authStore, clientService, userService, authTokenParams)

	grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptor.TraceInterceptor(log), interceptor.AuthInterceptor(clientService, PublicRPCs), interceptor.ErrorInterceptor(log)),
	)
	client.RegisterGRPC(grpcServer, clientService, validator)
	auth.RegisterGRPC(grpcServer, authService, validator)

	return grpcServer, dbconn, redisConn, nil
}

func cryptInit(cfg *config.Config) (*rsa.Manager, *argon2.Hasher, *blake3.Hasher, error) {
	op := pkg.Op("App.cryptInit")

	privRsa, err := rsa.DecodePrivatePemFile(cfg.Crypt.JWTSignCertPath)
	if err != nil {
		return nil, nil, nil, op.Err(err)
	}

	rsaManager := rsa.NewManager(privRsa)

	argon2Hasher := &argon2.Hasher{
		Memory:      cfg.Crypt.Hash.Memory,
		Iterations:  cfg.Crypt.Hash.Iterations,
		Parallelism: cfg.Crypt.Hash.Parallelism,
		KeyLength:   cfg.Crypt.Hash.KeyLength,
		SaltGenerator: func() ([]byte, error) {
			return crypt.GenerateRandomBytes(cfg.Crypt.Hash.SaltLength)
		},
	}

	blake3Hasher := &blake3.Hasher{
		SaltGenerator: func() ([]byte, error) {
			return crypt.GenerateRandomBytes(cfg.Crypt.Hash.SaltLength)
		},
	}

	return rsaManager, argon2Hasher, blake3Hasher, nil
}

func dbInit(ctx context.Context, cfg *config.Config, log *slog.Logger) (*pgx.Conn, *redis.Client, error) {
	op := pkg.Op("App.dbInit")

	dbconn, err := pgx.ConnectConfig(ctx, cfg.Store.PGX)
	if err != nil {
		return nil, nil, op.Err(err)
	}

	redis.SetLogger(&redisSlogAdapter{
		log: log,
	})
	redisConn := redis.NewClient(cfg.Store.Redis)
	err = redisConn.Ping(ctx).Err()
	if err != nil {
		return nil, nil, op.Err(err)
	}

	return dbconn, redisConn, nil
}
