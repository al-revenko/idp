package client

import (
	"context"

	"github.com/al-revenko/idp/internal/domain"
	"github.com/al-revenko/idp/internal/lib/crypt"
)

type HashProvider interface {
	Create(str string) (string, error)
	Compare(str, hash string) (bool, error)
}

type Service struct {
	store *Store
	hash  HashProvider
}

func NewService(store *Store, hash HashProvider) *Service {
	return &Service{store: store, hash: hash}
}

func (s *Service) GetClientById(ctx context.Context, clientId string) (domain.Client, error) {
	op := pkg.Op("Service.GetClientById")

	client, err := s.store.GetClientById(ctx, clientId)
	if err != nil {
		return domain.Client{}, op.Err(err)
	}

	return domain.Client{ID: client.ID, Name: client.Name}, nil
}

func (s *Service) RegisterClient(ctx context.Context, name string) (clientId, clientSecret string, err error) {
	op := pkg.Op("Service.RegisterClient")

	clientSecret, err = crypt.GenerateOpagueToken()
	if err != nil {
		return "", "", op.Err(err)
	}

	secretHash, err := s.hash.Create(clientSecret)
	if err != nil {
		return "", "", op.Err(err)
	}

	clientId, err = s.store.CreateClient(ctx, name, secretHash)
	if err != nil {
		return "", "", op.Err(err)
	}

	return clientId, clientSecret, nil
}

func (s *Service) DeleteClient(ctx context.Context, clientId, clientSecret string) error {
	op := pkg.Op("Service.DeleteClient")

	client, err := s.store.GetClientById(ctx, clientId)
	if err != nil {
		return op.Err(err)
	}

	isValidToken, err := s.hash.Compare(clientSecret, client.SecretHash)
	if err != nil {
		return op.Err(err)
	}
	if !isValidToken {
		return op.Err(domain.Error(domain.CodeUnauthorized, "invalid token", nil))
	}

	err = s.store.DeleteClient(ctx, clientId)
	if err != nil {
		return op.Err(err)
	}

	return nil
}
