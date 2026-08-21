package client

import (
	"github.com/al-revenko/idp/internal/lib/derr"
)

type ClientWithSecretHash struct {
	ID         string
	Name       string
	SecretHash string
}

var (
	ClientNotFoundError = derr.New(derr.CodeNotFound, "client not found")
	ConflictError       = derr.New(derr.CodeConflict, "conflict")
	InvalidTokenError   = derr.New(derr.CodeUnauthorized, "invalid token")
)
