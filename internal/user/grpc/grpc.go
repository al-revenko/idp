package grpc

import (
	"context"

	"github.com/al-revenko/idp/internal/lib/valid"
	idpv1 "github.com/al-revenko/idp/proto/gen/idp"
	"google.golang.org/grpc"
)

type UserServiceProvider interface {
	RegisterUser(ctx context.Context, username, password string) (string, error)
	LoginUser(ctx context.Context, username, password, clientID string) (accessToken, refreshToken string, err error)
}

type Server struct {
	idpv1.UnimplementedUserServiceServer
	service UserServiceProvider
}

func Register(grpc *grpc.Server, service UserServiceProvider) {
	idpv1.RegisterUserServiceServer(grpc, &Server{
		service: service,
	})
}

func (s *Server) UserRegister(ctx context.Context, req *idpv1.UserRegisterRequest) (*idpv1.UserRegisterResponse, error) {
	reqDTO := &UserRegisterRequest{
		Username: req.Username,
		Password: req.Password,
	}
	if err := valid.RequestDTO(reqDTO); err != nil {
		return nil, err
	}

	userId, err := s.service.RegisterUser(ctx, req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	return &idpv1.UserRegisterResponse{
		UserId: userId,
	}, nil
}

func (s *Server) UserLogin(ctx context.Context, req *idpv1.UserLoginRequest) (*idpv1.UserLoginResponse, error) {
	reqDTO := &UserLoginRequest{
		Username: req.Username,
		Password: req.Password,
		ClientId: req.ClientId,
	}
	if err := valid.RequestDTO(reqDTO); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.service.LoginUser(ctx, req.Username, req.Password, req.ClientId)
	if err != nil {
		return nil, err
	}

	return &idpv1.UserLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
