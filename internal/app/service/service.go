package service

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"

	"github.com/al-revenko/idp/internal/lib/pkgmark"
)

var pkg = pkgmark.New("app/service")

type PubKeyProvider interface {
	PubKey() *rsa.PublicKey
}

type Service struct {
	keysProvider PubKeyProvider
}

func New(keyProvider PubKeyProvider) *Service {
	return &Service{
		keysProvider: keyProvider,
	}
}

func (s *Service) GetPubKeyPemBlock() ([]byte, error) {
	op := pkg.Op("Service.GetPubKeyPemBlock")

	rsaPubKey := s.keysProvider.PubKey()

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
