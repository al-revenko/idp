package rsa

import (
	"crypto/rsa"
	"errors"

	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwt"
)

type Manager struct {
	privateKey *rsa.PrivateKey
}

func NewManager(privateKey *rsa.PrivateKey) *Manager {
	return &Manager{
		privateKey: privateKey,
	}
}

func (m *Manager) PubKey() *rsa.PublicKey {
	return &m.privateKey.PublicKey
}

func (m *Manager) SignJWT(token *jwt.Token) ([]byte, error) {
	op := pkg.Op("Manager.SignJWT")

	if token == nil {
		return nil, op.Err(errors.New("token is nil"))
	}

	signedToken, err := jwt.Sign(*token, jwt.WithKey(jwa.RS256(), m.privateKey))
	if err != nil {
		return nil, op.Err(err)
	}
	return signedToken, nil
}
