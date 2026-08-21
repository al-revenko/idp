package crypt

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/al-revenko/idp/internal/lib/pkgmark"
)

var pkg = pkgmark.New("crypt")

func GenerateRandomBytes(size uint32) ([]byte, error) {
	op := pkg.Op("GenerateRandomBytes")

	bytes := make([]byte, size)
	if _, err := rand.Read(bytes); err != nil {
		return nil, op.Err(err)
	}

	return bytes, nil
}

func GenerateOpagueToken() (string, error) {
	op := pkg.Op("GenerateOpagueToken")

	secret, err := GenerateRandomBytes(32)
	if err != nil {
		return "", op.Err(err)
	}

	return hex.EncodeToString(secret), nil
}
