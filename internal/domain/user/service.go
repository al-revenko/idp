package user

import (
	"context"

	"github.com/al-revenko/idp/internal/domain/model"
)

type Hasher interface {
	Hash(str string, salt []byte) (string, error)
	Compare(str, hash string) (bool, error)
}

type Service struct {
	store *Store
	hash  Hasher
}

func NewService(store *Store, hash Hasher) *Service {
	return &Service{store: store, hash: hash}
}

func (s *Service) CreateUser(ctx context.Context, username, password string) (userId string, err error) {
	op := pkg.Op("Service.CreateUser")

	passwordHash, err := s.hash.Hash(password, nil)
	if err != nil {
		return "", op.Err(err)
	}

	userId, err = s.store.CreateUser(ctx, username, passwordHash)
	if err != nil {
		return "", op.Err(err)
	}

	return userId, nil
}

func (s *Service) GetUserByUsername(ctx context.Context, username string) (model.User, error) {
	op := pkg.Op("Service.GetUserByUsername")

	user, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		return model.User{}, op.Err(err)
	}

	return user, nil
}

func (s *Service) GetUserById(ctx context.Context, userId string) (model.User, error) {
	op := pkg.Op("Service.GetUserById")

	user, err := s.store.GetUserById(ctx, userId)
	if err != nil {
		return model.User{}, op.Err(err)
	}

	return user, nil
}

func (s *Service) ComparePassword(ctx context.Context, password string, passwordHash string) (bool, error) {
	op := pkg.Op("Service.ComparePassword")

	equal, err := s.hash.Compare(password, passwordHash)
	if err != nil {
		return false, op.Err(err)
	}

	return equal, nil
}
