package client

import (
	"context"

	"github.com/al-revenko/idp/internal/domain/client/dto"
	idpv1 "github.com/al-revenko/idp/proto/gen/idp"
	"google.golang.org/grpc"
)

type Validator interface {
	Struct(interface{}) error
}

type GRPC struct {
	idpv1.UnimplementedClientServiceServer
	service  *Service
	validate Validator
}

func RegisterGRPC(grpc *grpc.Server, clientService *Service, validator Validator) {
	idpv1.RegisterClientServiceServer(grpc, &GRPC{service: clientService, validate: validator})
}

func (g *GRPC) ClientCreate(ctx context.Context, req *idpv1.ClientCreateRequest) (*idpv1.ClientCreateResponse, error) {
	reqDTO := &dto.ClientCreateRequest{Name: req.Name}
	if err := g.validate.Struct(reqDTO); err != nil {
		return nil, err
	}

	clientId, clientSecretToken, err := g.service.CreateClient(ctx, reqDTO.Name)
	if err != nil {
		return nil, err
	}

	return &idpv1.ClientCreateResponse{ClientId: clientId, ClientSecretToken: clientSecretToken}, nil
}

func (g *GRPC) ClientDelete(ctx context.Context, req *idpv1.ClientDeleteRequest) (*idpv1.ClientDeleteResponse, error) {
	reqDTO := &dto.ClientDeleteRequest{ClientId: req.ClientId}
	if err := g.validate.Struct(reqDTO); err != nil {
		return nil, err
	}

	err := g.service.DeleteClient(ctx, req.ClientId)
	if err != nil {
		return nil, err
	}

	return &idpv1.ClientDeleteResponse{Success: true}, nil
}
