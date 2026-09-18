package client

import (
	"context"

	"github.com/al-revenko/idp/internal/domain/model"
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) GetClientById(ctx context.Context, clientId string) (model.Client, error) {
	op := pkg.Op("Service.GetClientById")

	client, err := s.store.GetClientById(ctx, clientId)
	if err != nil {
		return model.Client{}, op.Err(err)
	}

	return client, nil
}

func (s *Service) CreateClient(ctx context.Context, name, pubKeyUrl string) (clientId string, err error) {
	op := pkg.Op("Service.CreateClient")

	clientId, err = s.store.CreateClient(ctx, name, pubKeyUrl)
	if err != nil {
		return "", op.Err(err)
	}

	return clientId, nil
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
