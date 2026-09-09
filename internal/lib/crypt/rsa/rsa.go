package rsa

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/al-revenko/idp/internal/lib/sign"
)

var pkg = sign.Pkg("crypt/rsa")

func GenerateKey() (*rsa.PrivateKey, error) {
	op := pkg.Op("GenerateRSAKeyPair")

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, op.Err(fmt.Errorf("Failed to generate RSA private key: %w", err))
	}

	return privateKey, nil
}

func DecodePrivatePemFile(privateKeyPath string) (*rsa.PrivateKey, error) {
	op := pkg.Op("DecodePrivatePemFile")

	pemData, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, op.Err(fmt.Errorf("Failed to read private key file: %w", err))
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, op.Err(errors.New("failed to parse PEM block"))
	}

	switch block.Type {
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(block.Bytes)

	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, op.Err(fmt.Errorf("failed to parse PKCS#8 private key: %w", err))
		}

		rsaKey, ok := key.(*rsa.PrivateKey)
		if !ok {
			return nil, op.Err(errors.New("not an RSA private key"))
		}
		return rsaKey, nil

	default:
		return nil, op.Err(fmt.Errorf("unsupported key type: %s", block.Type))
	}
}
