package client

import (
	"context"

	"github.com/al-revenko/idp/internal/domain/model"
	"github.com/al-revenko/idp/internal/lib/crypt"
)

type Hasher interface {
	Hash(secret string, salt []byte) (string, error)
	Compare(secret, hash string) (bool, error)
}

type Service struct {
	store *Store
	hash  Hasher
}

func NewService(store *Store, hash Hasher) *Service {
	return &Service{store: store, hash: hash}
}

func (s *Service) GetClientById(ctx context.Context, clientId string) (model.Client, error) {
	op := pkg.Op("Service.GetClientById")

	client, err := s.store.GetClientById(ctx, clientId)
	if err != nil {
		return model.Client{}, op.Err(err)
	}

	return client, nil
}

func (s *Service) CreateClient(ctx context.Context, name string) (clientId, clientSecret string, err error) {
	op := pkg.Op("Service.CreateClient")

	clientSecret, err = crypt.GenerateOpagueToken(32)
	if err != nil {
		return "", "", op.Err(err)
	}

	secretHash, err := s.hash.Hash(clientSecret, nil)
	if err != nil {
		return "", "", op.Err(err)
	}

	clientId, err = s.store.CreateClient(ctx, name, secretHash)
	if err != nil {
		return "", "", op.Err(err)
	}

	return clientId, clientSecret, nil
}

func (s *Service) DeleteClient(ctx context.Context, clientId string) error {
	op := pkg.Op("Service.DeleteClient")

	_, err := s.store.GetClientById(ctx, clientId)
	if err != nil {
		return op.Err(err)
	}

	err = s.store.DeleteClient(ctx, clientId)
	if err != nil {
		return op.Err(err)
	}

	return nil
}

func (s *Service) CompareSecret(ctx context.Context, secret string, secretHash string) (bool, error) {
	op := pkg.Op("Service.CompareSecret")

	equal, err := s.hash.Compare(secret, secretHash)
	if err != nil {
		return false, op.Err(err)
	}

	return equal, nil
}
