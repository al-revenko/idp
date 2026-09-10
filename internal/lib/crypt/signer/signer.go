package signer

import (
	"crypto/rsa"
	"errors"

	"github.com/al-revenko/idp/internal/lib/pkgmark"
	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwt"
)

var pkg = pkgmark.New("crypt/signer")

type Signer struct {
	rsaPrivateKey *rsa.PrivateKey
	rsaPublicKey  *rsa.PublicKey
}

func New(rsaPrivateKey *rsa.PrivateKey, rsaPublicKey *rsa.PublicKey) *Signer {
	return &Signer{
		rsaPrivateKey: rsaPrivateKey,
		rsaPublicKey:  rsaPublicKey,
	}
}

func (s *Signer) PubKey() *rsa.PublicKey {
	return s.rsaPublicKey
}

func (s *Signer) SignJWT(token *jwt.Token) ([]byte, error) {
	op := pkg.Op("Signer.SignJWT")

	if token == nil {
		return nil, op.Err(errors.New("token is nil"))
	}

	signedToken, err := jwt.Sign(*token, jwt.WithKey(jwa.RS256(), s.rsaPrivateKey))
	if err != nil {
		return nil, op.Err(err)
	}
	return signedToken, nil
}
