package auth

import (
	"context"

	"github.com/al-revenko/idp/internal/domain"
	"github.com/al-revenko/idp/internal/domain/auth/dto"
	"github.com/al-revenko/idp/internal/lib/meta"
	idpv1 "github.com/al-revenko/idp/proto/gen/idp"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Validator interface {
	Struct(interface{}) error
}

type GRPC struct {
	idpv1.UnimplementedAuthServiceServer
	service  *Service
	validate Validator
}

func RegisterGRPC(grpc *grpc.Server, service *Service, validator Validator) {
	idpv1.RegisterAuthServiceServer(grpc, &GRPC{service: service, validate: validator})
}

func (g *GRPC) PublicKey(ctx context.Context, req *emptypb.Empty) (*idpv1.PublicKeyResponse, error) {
	pem, err := g.service.GetPubKeyPemBlock()
	if err != nil {
		return nil, err
	}

	return &idpv1.PublicKeyResponse{RsaPubPem: pem}, nil
}

func (g *GRPC) Register(ctx context.Context, req *idpv1.RegisterRequest) (*idpv1.RegisterResponse, error) {
	reqDTO := &dto.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
	}
	if err := g.validate.Struct(reqDTO); err != nil {
		return nil, err
	}

	userId, err := g.service.Register(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return &idpv1.RegisterResponse{
		UserId: userId,
	}, nil
}

func (g *GRPC) Login(ctx context.Context, req *idpv1.LoginRequest) (*idpv1.LoginResponse, error) {
	reqDTO := &dto.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}
	if err := g.validate.Struct(reqDTO); err != nil {
		return nil, err
	}

	clientId, ok := ctx.Value(meta.CtxClientIDKey{}).(string)
	if !ok {
		return nil, domain.Error(domain.CodeInternal, "ctx: clientId not provided", nil)
	}

	accessToken, refreshToken, err := g.service.Login(ctx, clientId, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return &idpv1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (g *GRPC) RefreshToken(ctx context.Context, req *idpv1.RefreshTokenRequest) (*idpv1.RefreshTokenResponse, error) {
	accessToken, refreshToken, err := g.service.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &idpv1.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (g *GRPC) Logout(ctx context.Context, req *idpv1.LogoutRequest) (*emptypb.Empty, error) {
	err := g.service.Logout(ctx, req.RefreshToken)
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
