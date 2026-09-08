package service

import (
	"context"

	cl "github.com/al-revenko/idp/internal/client"
	"github.com/al-revenko/idp/internal/lib/apperr"
	"github.com/al-revenko/idp/internal/lib/crypt"
	"github.com/al-revenko/idp/internal/lib/pkgmark"
)

var pkg = pkgmark.New("client/service")

type HashProvider interface {
	Create(str string) (string, error)
	Compare(str, hash string) (bool, error)
}

type Store interface {
	CreateClient(ctx context.Context, name, secretHash string) (string, error)
	GetClientById(ctx context.Context, id string) (cl.ClientWithSecretHash, error)
	DeleteClient(ctx context.Context, id string) error
}

type Service struct {
	store Store
	hash  HashProvider
}

func New(store Store, hash HashProvider) *Service {
	return &Service{
		store: store,
		hash:  hash,
	}
}

func (s *Service) GetClientById(ctx context.Context, clientId string) (cl.Client, error) {
	op := pkg.Op("Service.GetClientById")

	client, err := s.store.GetClientById(ctx, clientId)
	if err != nil {
		return cl.Client{}, op.Err(err)
	}

	return cl.Client{ID: client.ID, Name: client.Name}, nil
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
		return op.Err(apperr.New(apperr.CodeUnauthorized, "invalid token"))
	}

	err = s.store.DeleteClient(ctx, clientId)
	if err != nil {
		return op.Err(err)
	}

	return nil
}
