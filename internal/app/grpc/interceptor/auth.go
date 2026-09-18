package interceptor

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/al-revenko/idp/internal/domain/model"
	"github.com/al-revenko/idp/internal/lib/meta"
	"github.com/jwx-go/jwkfetch/v4"
	"github.com/lestrrat-go/jwx/v4/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ClientService interface {
	GetClientById(ctx context.Context, id string) (model.Client, error)
}

type Validator interface {
	UUID(s any) error
}

const bearerPrefix = "Bearer "

type AuthInterceptorParams struct {
	ClientService ClientService
	Validator     Validator
	ExpectedAud   string
	PublicRPCs    map[string]struct{}
}

func AuthInterceptor(params AuthInterceptorParams) grpc.UnaryServerInterceptor {
	jwksClient := jwkfetch.NewClient()

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if _, ok := params.PublicRPCs[info.FullMethod]; ok {
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
		jwtStr := parts[1]
		if err := params.Validator.UUID(clientID); err != nil {
			return nil, status.Error(codes.Unauthenticated, "clientId must be UUID")
		}

		client, err := params.ClientService.GetClientById(ctx, clientID)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}

		jwks, err := jwksClient.Fetch(ctx, client.PubKeyUrl)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "cannot fetch JWKS")
		}

		_, err = jwt.Parse(
			[]byte(jwtStr),
			jwt.WithKeySet(jwks),
			jwt.WithIssuer(client.ID),
			jwt.WithSubject(client.ID),
			jwt.WithAudience(params.ExpectedAud),
		)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid JWT")
		}

		ctx = context.WithValue(ctx, meta.CtxClientIDKey{}, client.ID)

		return handler(ctx, req)
	}
}
