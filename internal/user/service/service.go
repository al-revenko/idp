package service

import (
	"context"

	"github.com/al-revenko/idp/internal/client"
	"github.com/al-revenko/idp/internal/lib/apperr"
	"github.com/al-revenko/idp/internal/lib/pkgmark"
	"github.com/al-revenko/idp/internal/user"
)

var pkg = pkgmark.New("user/service")

type UserStore interface {
	CreateUser(ctx context.Context, username, passwordHash string) (string, error)
	GetUserByUsername(ctx context.Context, username string) (user.UserWithPasswordHash, error)
}

type ClientProvider interface {
	GetClientById(ctx context.Context, clientId string) (client.Client, error)
}

type HashProvider interface {
	Create(str string) (string, error)
	Compare(str, hash string) (bool, error)
}

type Service struct {
	store  UserStore
	client ClientProvider
	hash   HashProvider
}

func New(store UserStore, client ClientProvider, h HashProvider) *Service {
	return &Service{store: store, client: client, hash: h}
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
		return "", "", op.Err(apperr.From(apperr.CodeNotFound, "user not found", err))
	}

	match, err := s.hash.Compare(password, u.PasswordHash)
	if err != nil {
		return "", "", op.Err(err)
	}
	if !match {
		return "", "", op.Err(apperr.From(apperr.CodeUnauthorized, "invalid password", nil))
	}

	return "access_token", "refresh_token", nil
}
