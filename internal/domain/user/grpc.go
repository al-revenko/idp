package user

import (
	"context"

	"github.com/al-revenko/idp/internal/domain/user/dto"
	"github.com/al-revenko/idp/internal/lib/valid"
	idpv1 "github.com/al-revenko/idp/proto/gen/idp"
	"google.golang.org/grpc"
)

type GRPC struct {
	idpv1.UnimplementedUserServiceServer
	service *Service
}

func RegisterGRPC(grpc *grpc.Server, service *Service) {
	idpv1.RegisterUserServiceServer(grpc, &GRPC{
		service: service,
	})
}

func (g *GRPC) UserRegister(ctx context.Context, req *idpv1.UserRegisterRequest) (*idpv1.UserRegisterResponse, error) {
	reqDTO := &dto.UserRegisterRequest{
		Username: req.Username,
		Password: req.Password,
	}
	if err := valid.Struct(reqDTO); err != nil {
		return nil, err
	}

	userId, err := g.service.RegisterUser(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return &idpv1.UserRegisterResponse{
		UserId: userId,
	}, nil
}

func (g *GRPC) UserLogin(ctx context.Context, req *idpv1.UserLoginRequest) (*idpv1.UserLoginResponse, error) {
	reqDTO := &dto.UserLoginRequest{
		Username: req.Username,
		Password: req.Password,
		ClientId: req.ClientId,
	}
	if err := valid.Struct(reqDTO); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := g.service.LoginUser(ctx, req.Username, req.Password, req.ClientId)
	if err != nil {
		return nil, err
	}

	return &idpv1.UserLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
