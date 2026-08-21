package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"github.com/al-revenko/idp/internal/lib/pkgmark"
)

var pkg = pkgmark.New("crypt/rsa")

func GenerateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	op := pkg.Op("GenerateRSAKeyPair")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, op.Err(fmt.Errorf("Failed to generate RSA private key: %w", err))
	}

	publicKey := &privateKey.PublicKey

	return privateKey, publicKey, nil
}

func DecodeBase64Keys(privateKeyStr, publicKeyStr string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	op := pkg.Op("DecodeBase64RSAKeys")

	privateKeyBytes, err := base64.StdEncoding.DecodeString(privateKeyStr)
	if err != nil {
		return nil, nil, op.Err(fmt.Errorf("decode private key: %w", err))
	}

	publicKeyBytes, err := base64.StdEncoding.DecodeString(publicKeyStr)
	if err != nil {
		return nil, nil, op.Err(fmt.Errorf("decode public key: %w", err))
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, nil, op.Err(fmt.Errorf("parse private key: %w", err))
	}

	publicKey, err := x509.ParsePKCS1PublicKey(publicKeyBytes)
	if err != nil {
		return nil, nil, op.Err(fmt.Errorf("parse public key: %w", err))
	}

	return privateKey, publicKey, nil
}
