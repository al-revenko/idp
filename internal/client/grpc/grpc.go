package grpc

import (
	"context"

	idpv1 "github.com/al-revenko/idp/proto/gen/idp"
	"google.golang.org/grpc"
)

type ClientProvider interface {
	RegisterClient(ctx context.Context, name string) (string, string, error)
	DeleteClient(ctx context.Context, clientId, clientSecret string) error
}

type Server struct {
	idpv1.UnimplementedClientServiceServer
	client ClientProvider
}

func Register(grpc *grpc.Server, clientService ClientProvider) {
	idpv1.RegisterClientServiceServer(grpc, &Server{client: clientService})
}

func (s *Server) ClientRegister(ctx context.Context, req *idpv1.ClientRegisterRequest) (*idpv1.ClientRegisterResponse, error) {
	clientId, clientSecretToken, err := s.client.RegisterClient(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	return &idpv1.ClientRegisterResponse{ClientId: clientId, ClientSecretToken: clientSecretToken}, nil
}

func (s *Server) ClientDelete(ctx context.Context, req *idpv1.ClientDeleteRequest) (*idpv1.ClientDeleteResponse, error) {
	err := s.client.DeleteClient(ctx, req.ClientId, req.ClientSecretToken)
	if err != nil {
		return nil, err
	}

	return &idpv1.ClientDeleteResponse{Success: true}, nil
}
