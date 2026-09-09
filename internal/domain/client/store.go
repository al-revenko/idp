package client

import (
	"context"
	"errors"

	"github.com/al-revenko/idp/db/sql/gen/dbstore"
	"github.com/al-revenko/idp/internal/domain"
	"github.com/jackc/pgx/v5"
)

type Store struct {
	db *dbstore.Queries
}

func NewStore(db *dbstore.Queries) *Store {
	return &Store{db: db}
}

func (s *Store) GetClientById(ctx context.Context, id string) (domain.Client, error) {
	op := pkg.Op("Store.GetClientById")

	c, err := s.db.GetClientById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Client{}, op.Err(domain.Error(domain.CodeNotFound, "client not found", err))
		}

		return domain.Client{}, op.Err(domain.Error(domain.CodeInternal, err.Error(), err))
	}

	return s.toDomainClient(c), nil
}

func (s *Store) CreateClient(ctx context.Context, name, secretHash string) (clientId string, err error) {
	op := pkg.Op("Store.CreateClient")

	clientId, err = s.db.CreateClient(ctx, dbstore.CreateClientParams{
		Name: name, SecretHash: secretHash,
	})
	if err != nil {
		return "", op.Err(domain.Error(domain.CodeInternal, err.Error(), err))
	}

	return clientId, nil
}

func (s *Store) DeleteClient(ctx context.Context, id string) error {
	op := pkg.Op("Store.DeleteClient")

	err := s.db.DeleteClient(ctx, id)
	if err != nil {
		return op.Err(domain.Error(domain.CodeInternal, err.Error(), err))
	}

	return nil
}

func (s *Store) toDomainClient(cdb dbstore.Client) domain.Client {
	return domain.Client{
		ID:         cdb.ID,
		Name:       cdb.Name,
		SecretHash: cdb.SecretHash,
	}
}
