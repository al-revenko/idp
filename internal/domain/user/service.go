package user

import (
	"context"

	"github.com/al-revenko/idp/internal/domain"
)

type ClientProvider interface {
	GetClientById(ctx context.Context, clientId string) (domain.Client, error)
}

type HashProvider interface {
	Create(str string) (string, error)
	Compare(str, hash string) (bool, error)
}

type Service struct {
	store  *Store
	client ClientProvider
	hash   HashProvider
}

func NewService(store *Store, client ClientProvider, hash HashProvider) *Service {
	return &Service{store: store, client: client, hash: hash}
}

func (s *Service) RegisterUser(ctx context.Context, username, password string) (userId string, err error) {
	op := pkg.Op("Service.RegisterUser")

	passwordHash, err := s.hash.Create(password)
	if err != nil {
		return "", op.Err(err)
	}

	userId, err = s.store.CreateUser(ctx, username, passwordHash)
	if err != nil {
		return "", op.Err(err)
	}

	return userId, nil
}

func (s *Service) LoginUser(ctx context.Context, username, password, clientID string) (accessToken, refreshToken string, err error) {
	op := pkg.Op("Service.LoginUser")

	_, err = s.client.GetClientById(ctx, clientID)
	if err != nil {
		return "", "", op.Err(err)
	}

	u, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		return "", "", op.Err(domain.Error(domain.CodeNotFound, "user not found", err))
	}

	match, err := s.hash.Compare(password, u.PasswordHash)
	if err != nil {
		return "", "", op.Err(err)
	}
	if !match {
		return "", "", op.Err(domain.Error(domain.CodeUnauthorized, "invalid password", nil))
	}

	return "access_token", "refresh_token", nil
}
