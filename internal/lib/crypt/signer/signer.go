package signer

import (
	"crypto/rsa"

	"github.com/al-revenko/idp/internal/lib/pkgmark"
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
