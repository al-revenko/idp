package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/al-revenko/idp/db/sql/gen/dbstore"
	"github.com/al-revenko/idp/internal/client"
	"github.com/al-revenko/idp/internal/lib/pkgmark"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
			return client.ClientWithSecretHash{}, op.Err(client.ClientNotFoundError)
		}
		return client.ClientWithSecretHash{}, op.Err(fmt.Errorf("getting client %d: %w", id, err))
	}
	return toDomainClient(c), nil
}

func (s *Store) CreateClient(ctx context.Context, name, secretHash string) (string, error) {
	op := pkg.Op("Store.CreateClient")

	id, err := s.queries.CreateClient(ctx, dbstore.CreateClientParams{
		Name: name, SecretHash: secretHash,
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return "", op.Err(client.ConflictError)
		}
		return "", op.Err(fmt.Errorf("creating client: %w", err))
	}
	return id, nil
}

func (s *Store) DeleteClient(ctx context.Context, id string) error {
	op := pkg.Op("Store.DeleteClient")

	err := s.queries.DeleteClient(ctx, id)
	if err != nil {
		return op.Err(fmt.Errorf("deleting client %d: %w", id, err))
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
