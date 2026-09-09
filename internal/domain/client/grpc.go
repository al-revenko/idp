package client

import (
	"context"

	"github.com/al-revenko/idp/internal/domain/client/dto"
	"github.com/al-revenko/idp/internal/lib/valid"
	idpv1 "github.com/al-revenko/idp/proto/gen/idp"
	"google.golang.org/grpc"
)

type GRPC struct {
	idpv1.UnimplementedClientServiceServer
	service *Service
}

func RegisterGRPC(grpc *grpc.Server, clientService *Service) {
	idpv1.RegisterClientServiceServer(grpc, &GRPC{service: clientService})
}

func (g *GRPC) ClientRegister(ctx context.Context, req *idpv1.ClientRegisterRequest) (*idpv1.ClientRegisterResponse, error) {
	reqDTO := &dto.ClientRegisterRequest{Name: req.Name}
	if err := valid.Struct(reqDTO); err != nil {
		return nil, err
	}

	clientId, clientSecretToken, err := g.service.RegisterClient(ctx, reqDTO.Name)
	if err != nil {
		return nil, err
	}

	return &idpv1.ClientRegisterResponse{ClientId: clientId, ClientSecretToken: clientSecretToken}, nil
}

func (g *GRPC) ClientDelete(ctx context.Context, req *idpv1.ClientDeleteRequest) (*idpv1.ClientDeleteResponse, error) {
	reqDTO := &dto.ClientDeleteRequest{ClientId: req.ClientId, ClientSecretToken: req.ClientSecretToken}
	if err := valid.Struct(reqDTO); err != nil {
		return nil, err
	}

	err := g.service.DeleteClient(ctx, req.ClientId, req.ClientSecretToken)
	if err != nil {
		return nil, err
	}

	return &idpv1.ClientDeleteResponse{Success: true}, nil
}
