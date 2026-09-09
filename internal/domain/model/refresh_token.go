package model

import (
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

type RefreshToken struct {
	FamilyId  string
	UserId    string
	ClientId  string
	Revoked   bool
	Used      bool
	UsedAt    time.Time
	IssuedAt  time.Time
	ExpiresAt time.Time
}

func (r RefreshToken) MarshalBinary() (data []byte, err error) {
	type Alias RefreshToken
	return msgpack.Marshal(Alias(r))
}

func (r *RefreshToken) UnmarshalBinary(data []byte) error {
	type Alias RefreshToken
	return msgpack.Unmarshal(data, (*Alias)(r))
}
