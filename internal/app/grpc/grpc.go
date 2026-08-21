package grpc

import (
	"context"

	idpv1 "github.com/al-revenko/idp/proto/gen/idp"
	"google.golang.org/grpc"
)

type ServiceProvider interface {
	GetPubKeyPemBlock() ([]byte, error)
}

type Server struct {
	idpv1.UnimplementedAppServiceServer
	service ServiceProvider
}

func Register(grpc *grpc.Server, service ServiceProvider) {
	idpv1.RegisterAppServiceServer(grpc, &Server{service: service})
}

func (s *Server) AppPublicKey(ctx context.Context, req *idpv1.AppPublicKeyRequest) (*idpv1.AppPublicKeyResponse, error) {
	pem, err := s.service.GetPubKeyPemBlock()
	if err != nil {
		return nil, err
	}

	return &idpv1.AppPublicKeyResponse{RsaPubPem: pem}, nil
}
