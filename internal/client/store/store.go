package store

import (
	"context"
	"errors"

	"github.com/al-revenko/idp/db/sql/gen/dbstore"
	"github.com/al-revenko/idp/internal/client"
	"github.com/al-revenko/idp/internal/lib/apperr"
	"github.com/al-revenko/idp/internal/lib/pkgmark"
	"github.com/jackc/pgx/v5"
)

var pkg = pkgmark.New("client/store")

type Store struct {
	queries *dbstore.Queries
}

func New(db dbstore.DBTX) *Store {
	return &Store{queries: dbstore.New(db)}
}

func (s *Store) GetClientById(ctx context.Context, id string) (client.ClientWithSecretHash, error) {
	op := pkg.Op("Store.GetClientById")

	c, err := s.queries.GetClientById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return client.ClientWithSecretHash{}, op.Err(apperr.From(apperr.CodeNotFound, "client not found", err))
		}
		return client.ClientWithSecretHash{}, op.Err(apperr.From(apperr.CodeInternal, err.Error(), err))
	}
	return toDomainClient(c), nil
}

func (s *Store) CreateClient(ctx context.Context, name, secretHash string) (clientId string, err error) {
	op := pkg.Op("Store.CreateClient")

	clientId, err = s.queries.CreateClient(ctx, dbstore.CreateClientParams{
		Name: name, SecretHash: secretHash,
	})
	if err != nil {
		return "", op.Err(apperr.From(apperr.CodeInternal, err.Error(), err))
	}
	return clientId, nil
}

func (s *Store) DeleteClient(ctx context.Context, id string) error {
	op := pkg.Op("Store.DeleteClient")

	err := s.queries.DeleteClient(ctx, id)
	if err != nil {
		return op.Err(apperr.From(apperr.CodeInternal, err.Error(), err))
	}
	return nil
}

func toDomainClient(s dbstore.Client) client.ClientWithSecretHash {
	return client.ClientWithSecretHash{
		ID:         s.ID,
		Name:       s.Name,
		SecretHash: s.SecretHash,
	}
}
