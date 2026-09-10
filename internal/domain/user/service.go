package user

import (
	"context"
	"time"
	"uuid"

	"github.com/al-revenko/idp/internal/domain"
	"github.com/al-revenko/idp/internal/lib/crypt"
	"github.com/lestrrat-go/jwx/v4/jwt"
)

type ClientProvider interface {
	GetClientById(ctx context.Context, clientId string) (domain.Client, error)
}

type HashProvider interface {
	Create(str string) (string, error)
	Compare(str, hash string) (bool, error)
}

type JWTSigner interface {
	SignJWT(token *jwt.Token) ([]byte, error)
}

type JWTParams struct {
	Signer          JWTSigner
	Issuer          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

type Service struct {
	store  *Store
	client ClientProvider
	hash   HashProvider
	jwt    *JWTParams
}

func NewService(jwtParams *JWTParams, store *Store, client ClientProvider, hash HashProvider) *Service {
	return &Service{store: store, client: client, hash: hash, jwt: jwtParams}
}

func (s *Service) RegisterUser(ctx context.Context, username, password string) (userId string, err error) {
	op := pkg.Op("Service.RegisterUser")

	passwordHash, err := s.hash.Create(password)
	if err != nil {
		return "", op.Err(err)
	}

	userId, err = s.store.CreateUser(ctx, username, passwordHash)
	if err != nil {
		return "", op.Err(err)
	}

	return userId, nil
}

func (s *Service) LoginUser(ctx context.Context, username, password, clientID string) (accessToken, refreshToken string, err error) {
	op := pkg.Op("Service.LoginUser")

	_, err = s.client.GetClientById(ctx, clientID)
	if err != nil {
		return "", "", op.Err(err)
	}

	invalidCredsMsg := "invalid credentials"

	user, err := s.store.GetUserByUsername(ctx, username)
	if err != nil {
		return "", "", op.Err(domain.Error(domain.CodeUnauthorized, invalidCredsMsg, err))
	}

	match, err := s.hash.Compare(password, user.PasswordHash)
	if err != nil {
		return "", "", op.Err(err)
	}
	if !match {
		return "", "", op.Err(domain.Error(domain.CodeUnauthorized, invalidCredsMsg, nil))
	}

	refreshToken, err = crypt.GenerateOpagueToken(32)
	if err != nil {
		return "", "", op.Err(err)
	}

	accessTokenB, _, err := s.MakeJWT(clientID, user)
	if err != nil {
		return "", "", op.Err(err)
	}

	return string(accessTokenB), refreshToken, nil
}

func (s *Service) MakeJWT(clientID string, user domain.User) (token []byte, jti string, err error) {
	op := pkg.Op("Service.MakeJWT")

	jti = uuid.NewV7().String()
	issuedAt := time.Now()

	tokenBuilder := jwt.NewBuilder()
	tokenBuilder.Issuer(s.jwt.Issuer)
	tokenBuilder.Subject(user.ID)
	tokenBuilder.Audience([]string{clientID})
	tokenBuilder.JwtID(jti)
	tokenBuilder.IssuedAt(issuedAt)
	tokenBuilder.Expiration(issuedAt.Add(s.jwt.AccessTokenTTL))
	tokenBuilder.Claim("username", user.Username)

	unsignedToken, err := tokenBuilder.Build()
	if err != nil {
		return nil, "", op.Err(err)
	}

	token, err = s.jwt.Signer.SignJWT(&unsignedToken)
	if err != nil {
		return nil, "", op.Err(err)
	}

	return token, jti, nil
}
