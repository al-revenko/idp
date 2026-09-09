package interceptor

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/al-revenko/idp/internal/domain/model"
	"github.com/al-revenko/idp/internal/lib/meta"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ClientService interface {
	GetClientById(ctx context.Context, id string) (model.Client, error)
	CompareSecret(ctx context.Context, secret string, secretHash string) (bool, error)
}

const bearerPrefix = "Bearer "

func AuthInterceptor(clientService ClientService, publicRPCs map[string]struct{}) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := publicRPCs[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		authHeaders := md.Get("authorization")
		if len(authHeaders) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		auth := authHeaders[0]
		if !strings.HasPrefix(auth, bearerPrefix) {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header format")
		}

		credsBase64 := auth[len(bearerPrefix):]
		decodedCreds, err := base64.StdEncoding.DecodeString(credsBase64)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid bearer token format")
		}

		parts := strings.SplitN(string(decodedCreds), ":", 2)
		if len(parts) != 2 {
			return nil, status.Error(codes.Unauthenticated, "invalid bearer token format")
		}

		clientID := parts[0]
		secret := parts[1]

		client, err := clientService.GetClientById(ctx, clientID)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}

		match, err := clientService.CompareSecret(ctx, secret, client.SecretHash)
		if err != nil || !match {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}

		ctx = context.WithValue(ctx, meta.CtxClientIDKey{}, clientID)

		return handler(ctx, req)
	}
}
