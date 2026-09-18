package client

import (
	"context"

	"github.com/al-revenko/idp/internal/domain"
	"github.com/al-revenko/idp/internal/domain/client/dto"
	"github.com/al-revenko/idp/internal/lib/meta"
	idpv1 "github.com/al-revenko/idp/proto/gen/idp"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
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
	reqDTO := &dto.ClientCreateRequest{Name: req.Name, PubKeyUrl: req.JwksUrl}
	if err := g.validate.Struct(reqDTO); err != nil {
		return nil, err
	}

	clientId, err := g.service.CreateClient(ctx, reqDTO.Name, reqDTO.PubKeyUrl)
	if err != nil {
		return nil, err
	}

	return &idpv1.ClientCreateResponse{ClientId: clientId}, nil
}

func (g *GRPC) ClientDelete(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	clientId := meta.ExtractClientID(ctx)
	if clientId == "" {
		return nil, domain.Error(domain.CodeInternal, "clientId not found in ctx", nil)
	}

	err := g.service.DeleteClient(ctx, clientId)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
