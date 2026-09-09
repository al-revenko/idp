package auth

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"time"
	"uuid"

	"github.com/al-revenko/idp/internal/domain"
	"github.com/al-revenko/idp/internal/domain/model"
	"github.com/al-revenko/idp/internal/lib/crypt"
	"github.com/lestrrat-go/jwx/v4/jwt"
)

var ErrRefreshTokenInvalid = domain.Error(domain.CodeUnauthenticated, "token invalid", nil)

var ErrRefreshTokenNotFound = domain.Error(domain.CodeUnauthenticated, "token not found", ErrRefreshTokenInvalid)
var ErrRefreshTokenRevoked = domain.Error(domain.CodeUnauthenticated, "token revoked", ErrRefreshTokenInvalid)
var ErrRefreshTokenUsed = domain.Error(domain.CodeUnauthenticated, "token used", ErrRefreshTokenInvalid)
var ErrRefreshTokenExpired = domain.Error(domain.CodeUnauthenticated, "token expired", ErrRefreshTokenInvalid)

type UserProvider interface {
	GetUserByUsername(ctx context.Context, username string) (model.User, error)
	GetUserById(ctx context.Context, id string) (model.User, error)
	CreateUser(ctx context.Context, username, password string) (string, error)
	ComparePassword(ctx context.Context, password, passwordHash string) (bool, error)
}

type ClientProvider interface {
	GetClientById(ctx context.Context, id string) (model.Client, error)
	CreateClient(ctx context.Context, name string) (string, string, error)
}

type RsaSigner interface {
	SignJWT(token *jwt.Token) ([]byte, error)
	PubKey() *rsa.PublicKey
}

type Hasher interface {
	Hash(secret string, salt []byte) (string, error)
	Compare(secret, hash string) (bool, error)
}

type TokenParams struct {
	Signer                      RsaSigner
	Hash                        Hasher
	Issuer                      string
	AccessTokenTTL              time.Duration
	RefreshTokenTTL             time.Duration
	RefreshTokenRevokedStoreTTL time.Duration
	RefreshTokenGracePeriod     time.Duration
}

type Service struct {
	store  *Store
	user   UserProvider
	client ClientProvider
	token  *TokenParams
}

func NewService(store *Store, client ClientProvider, user UserProvider, token *TokenParams) *Service {
	return &Service{store: store, client: client, user: user, token: token}
}

func (s *Service) GetPubKeyPemBlock() ([]byte, error) {
	op := pkg.Op("Service.GetPubKeyPemBlock")

	rsaPubKey := s.token.Signer.PubKey()

	pubKeyBytes, err := x509.MarshalPKIXPublicKey(rsaPubKey)
	if err != nil {
		return nil, op.Err(err)
	}

	pemBlock := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return pemBlock, nil
}

func (s *Service) Register(ctx context.Context, username, password string) (userId string, err error) {
	op := pkg.Op("Service.Register")

	userId, err = s.user.CreateUser(ctx, username, password)
	if err != nil {
		return "", op.Err(err)
	}

	return userId, nil
}

func (s *Service) Login(ctx context.Context, clientId string, username, password string) (accessToken, refreshToken string, err error) {
	op := pkg.Op("Service.Login")

	invalidCredsMsg := "invalid credentials"

	user, err := s.user.GetUserByUsername(ctx, username)
	if err != nil {
		if domain.IsErrCode(err, domain.CodeNotFound) {
			return "", "", op.Err(domain.Error(domain.CodeUnauthenticated, invalidCredsMsg, err))
		}

		return "", "", op.Err(err)
	}

	match, err := s.user.ComparePassword(ctx, password, user.PasswordHash)
	if err == nil && !match {
		return "", "", op.Err(domain.Error(domain.CodeUnauthenticated, invalidCredsMsg, nil))
	}
	if err != nil {
		return "", "", op.Err(err)
	}

	issuedAt := time.Now()

	refreshToken, err = s.generateRefreshToken()
	if err != nil {
		return "", "", op.Err(err)
	}
	refreshTokenHash, err := s.hashRefreshToken(refreshToken)
	if err != nil {
		return "", "", op.Err(err)
	}
	refreshTokenData := model.RefreshToken{
		FamilyId:  uuid.NewV7().String(),
		UserId:    user.ID,
		ClientId:  clientId,
		Revoked:   false,
		Used:      false,
		IssuedAt:  issuedAt,
		ExpiresAt: issuedAt.Add(s.token.RefreshTokenTTL),
	}

	unsignedAccessToken, err := s.MakeAccessToken(clientId, user, issuedAt)
	if err != nil {
		return "", "", op.Err(err)
	}
	accessTokenB, err := s.token.Signer.SignJWT(unsignedAccessToken)
	if err != nil {
		return "", "", op.Err(err)
	}

	err = s.store.SetRefreshToken(ctx, refreshTokenHash, refreshTokenData, &s.token.RefreshTokenTTL)

	if err != nil {
		return "", "", op.Err(err)
	}

	return string(accessTokenB), refreshToken, nil
}

func (s *Service) RefreshToken(ctx context.Context, oldRefreshToken string) (accessToken, refreshToken string, err error) {
	op := pkg.Op("Service.RefreshToken")

	now := time.Now()

	oldRefreshTokenHash, oldRefreshTokenData, err := s.ValidateRefreshToken(ctx, now, oldRefreshToken)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenUsed) {
			oldRefreshTokenData.Revoked = true
			if err := s.store.SetRefreshToken(ctx, oldRefreshTokenHash, oldRefreshTokenData, nil); err != nil {
				return "", "", op.Err(err)
			}

			return "", "", op.Err(domain.Error(domain.CodeUnauthenticated, "token reuse detected; please re-authenticate", err))
		}

		if errors.Is(err, ErrRefreshTokenInvalid) {
			return "", "", op.Err(domain.Error(domain.CodeUnauthenticated, "invalid token", err))
		}

		return "", "", op.Err(err)
	}

	user, err := s.user.GetUserById(ctx, oldRefreshTokenData.UserId)
	if err != nil {
		return "", "", op.Err(domain.Error(domain.CodeUnauthenticated, "invalid token", err))
	}

	unsignedAccessToken, err := s.MakeAccessToken(oldRefreshTokenData.ClientId, user, now)
	if err != nil {
		return "", "", op.Err(err)
	}
	accessTokenB, err := s.token.Signer.SignJWT(unsignedAccessToken)
	if err != nil {
		return "", "", op.Err(err)
	}

	newRefreshToken, err := s.generateRefreshToken()
	if err != nil {
		return "", "", op.Err(err)
	}
	newRefreshTokenHash, err := s.hashRefreshToken(newRefreshToken)
	if err != nil {
		return "", "", op.Err(err)
	}
	newRefreshTokenData := model.RefreshToken{
		FamilyId:  oldRefreshTokenData.FamilyId,
		UserId:    oldRefreshTokenData.UserId,
		ClientId:  oldRefreshTokenData.ClientId,
		Revoked:   false,
		Used:      false,
		UsedAt:    time.Time{},
		IssuedAt:  now,
		ExpiresAt: oldRefreshTokenData.ExpiresAt,
	}

	err = s.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		oldRefreshTokenTTL := s.token.RefreshTokenRevokedStoreTTL
		oldRefreshTokenData.Used = true
		oldRefreshTokenData.UsedAt = now
		oldRefreshTokenData.ExpiresAt = now.Add(oldRefreshTokenTTL)
		if err := s.store.SetRefreshToken(txCtx, oldRefreshTokenHash, oldRefreshTokenData, &oldRefreshTokenTTL); err != nil {
			return err
		}

		if err := s.store.SetRefreshToken(txCtx, newRefreshTokenHash, newRefreshTokenData, nil); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return "", "", op.Err(err)
	}

	return string(accessTokenB), newRefreshToken, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	op := pkg.Op("Service.Logout")

	now := time.Now()

	refreshTokenHash, refreshTokenData, err := s.ValidateRefreshToken(ctx, now, refreshToken)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenInvalid) {
			return nil
		}

		return op.Err(err)
	}

	refreshTokenData.Revoked = true
	refreshTokenData.Used = true

	if err := s.store.SetRefreshToken(ctx, refreshTokenHash, refreshTokenData, nil); err != nil {
		return op.Err(err)
	}

	return nil
}

func (s *Service) ValidateRefreshToken(ctx context.Context, now time.Time, refreshToken string) (hash string, data model.RefreshToken, err error) {
	op := pkg.Op("Service.ValidateRefreshToken")

	refreshTokenHash, err := s.hashRefreshToken(refreshToken)
	if err != nil {
		return "", model.RefreshToken{}, op.Err(err)
	}
	refreshTokenData, err := s.store.GetRefreshToken(ctx, refreshTokenHash)
	if err != nil {
		if domain.IsErrCode(err, domain.CodeNotFound) {
			return "", model.RefreshToken{}, op.Err(ErrRefreshTokenNotFound)
		}

		return "", model.RefreshToken{}, op.Err(err)
	}

	if refreshTokenData.ExpiresAt.Before(now) {
		return refreshTokenHash, refreshTokenData, op.Err(ErrRefreshTokenExpired)
	}

	if refreshTokenData.Revoked {
		return refreshTokenHash, refreshTokenData, op.Err(ErrRefreshTokenRevoked)
	}

	if refreshTokenData.Used && refreshTokenData.UsedAt.Add(s.token.RefreshTokenGracePeriod).Before(now) {
		return refreshTokenHash, refreshTokenData, op.Err(ErrRefreshTokenUsed)
	}

	return refreshTokenHash, refreshTokenData, nil
}

func (s *Service) MakeAccessToken(clientID string, user model.User, issuedAt time.Time) (*jwt.Token, error) {
	op := pkg.Op("Service.MakeAccessToken")

	tokenBuilder := jwt.NewBuilder()
	tokenBuilder.Issuer(s.token.Issuer)
	tokenBuilder.Subject(user.ID)
	tokenBuilder.Audience([]string{clientID})
	tokenBuilder.JwtID(uuid.NewV7().String())
	tokenBuilder.IssuedAt(issuedAt)
	tokenBuilder.Expiration(issuedAt.Add(s.token.AccessTokenTTL))
	tokenBuilder.Claim("username", user.Username)

	token, err := tokenBuilder.Build()
	if err != nil {
		return nil, op.Err(err)
	}

	return &token, nil
}

func (s *Service) generateRefreshToken() (string, error) {
	op := pkg.Op("Service.generateRefreshToken")

	refreshToken, err := crypt.GenerateOpagueToken(32)
	if err != nil {
		return "", op.Err(err)
	}

	return refreshToken, nil
}

func (s *Service) hashRefreshToken(refreshToken string) (string, error) {
	op := pkg.Op("Service.hashRefreshToken")

	refreshTokenHash, err := s.token.Hash.Hash(refreshToken, []byte("-"))
	if err != nil {
		return "", op.Err(err)
	}

	return refreshTokenHash, nil
}
