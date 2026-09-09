package app

import (
	"github.com/al-revenko/idp/db/sql/gen/dbstore"
	"github.com/al-revenko/idp/internal/domain/client"
	"github.com/al-revenko/idp/internal/domain/user"
	"github.com/al-revenko/idp/internal/lib/crypt"
	"github.com/al-revenko/idp/internal/lib/crypt/hash"
	"github.com/al-revenko/idp/internal/lib/crypt/rsa"
	"github.com/al-revenko/idp/internal/lib/crypt/signer"
	"github.com/jackc/pgx/v5"
)

func (a *App) init() error {
	op := pkg.Op("App.init")

	signer, hasher, err := a.cryptInit()
	if err != nil {
		return op.Err(err)
	}

	dbconn, err := a.dbInit()
	if err != nil {
		return op.Err(err)
	}
	a.dbconn = dbconn

	a.srvsInit(dbconn, signer, hasher)

	return nil
}

func (a *App) srvsInit(dbconn *pgx.Conn, signer *signer.Signer, hasher *hash.Argon2) {
	db := dbstore.New(a.dbconn)

	appService := NewService(signer)
	RegisterGRPC(a.grpcServer, appService)

	clientStore := client.NewStore(db)
	clientService := client.NewService(clientStore, hasher)
	client.RegisterGRPC(a.grpcServer, clientService)

	userStore := user.NewStore(db)
	userService := user.NewService(userStore, clientService, hasher)
	user.RegisterGRPC(a.grpcServer, userService)
}

func (a *App) cryptInit() (*signer.Signer, *hash.Argon2, error) {
	op := pkg.Op("App.cryptInit")

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
	op := pkg.Op("App.dbInit")

	dbconn, err := pgx.Connect(a.ctx, a.config.Secrets.DBString)
	if err != nil {
		return nil, op.Err(err)
	}

	return dbconn, nil
}
