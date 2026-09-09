package argon2

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/al-revenko/idp/internal/lib/crypt/hash"
	"github.com/al-revenko/idp/internal/lib/sign"
	"golang.org/x/crypto/argon2"
)

var pkg = sign.Pkg("hash/argon2")

const Algorithm = "argon2id"

type Argon2 struct {
	hash.Hash
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
}

type Hasher struct {
	Memory        uint32
	Iterations    uint32
	Parallelism   uint8
	KeyLength     uint32
	SaltGenerator func() ([]byte, error)
}

func (h *Hasher) Hash(data string, salt []byte) (string, error) {
	op := pkg.Op("Hasher.Hash")

	var saltValue []byte = salt
	if salt == nil {
		genSalt, err := h.SaltGenerator()
		if err != nil {
			return "", op.Err(err)
		}
		saltValue = genSalt
	}

	hash := argon2.IDKey(
		[]byte(data),
		saltValue,
		h.Iterations,
		h.Memory,
		h.Parallelism,
		h.KeyLength,
	)

	strBuilder := strings.Builder{}
	strBuilder.WriteString(Algorithm)
	strBuilder.WriteString("$")
	strBuilder.WriteString(fmt.Sprintf("i=%d_m=%d_p=%d", h.Iterations, h.Memory, h.Parallelism))
	strBuilder.WriteString("$")
	strBuilder.WriteString(base64.RawStdEncoding.EncodeToString(saltValue))
	strBuilder.WriteString("$")
	strBuilder.WriteString(base64.RawStdEncoding.EncodeToString(hash))

	return strBuilder.String(), nil
}

func (h *Hasher) Compare(str, hash string) (bool, error) {
	op := pkg.Op("Hasher.Compare")

	decodedHash, err := h.Decode(hash)
	if err != nil {
		return false, op.Err(err)
	}

	otherKey := argon2.IDKey(
		[]byte(str),
		decodedHash.Salt,
		decodedHash.Iterations,
		decodedHash.Memory,
		decodedHash.Parallelism,
		uint32(len(decodedHash.Key)),
	)

	keyLen := int32(len(decodedHash.Key))
	otherKeyLen := int32(len(otherKey))

	if subtle.ConstantTimeEq(keyLen, otherKeyLen) == 0 {
		subtle.ConstantTimeCompare(decodedHash.Key, decodedHash.Key)
		return false, nil
	}

	isEqual := subtle.ConstantTimeCompare(decodedHash.Key, otherKey) == 1
	return isEqual, nil
}

func (h *Hasher) Decode(hash string) (Argon2, error) {
	op := pkg.Op("Hasher.Decode")

	parts := strings.Split(hash, "$")
	if len(parts) != 4 {
		return Argon2{}, op.Err(fmt.Errorf("invalid hash format"))
	}
	decoded := Argon2{}

	alg := parts[0]
	params := parts[1]
	salt := parts[2]
	key := parts[3]

	if alg != Algorithm {
		return Argon2{}, op.Err(fmt.Errorf("invalid algorithm: %s != %s", alg, Algorithm))
	}
	decoded.Algorithm = alg

	_, err := fmt.Sscanf(params, "i=%d_m=%d_p=%d", &decoded.Iterations, &decoded.Memory, &decoded.Parallelism)
	if err != nil {
		return Argon2{}, op.Err(fmt.Errorf("invalid params format: %w", err))
	}

	bkey, err := base64.RawStdEncoding.Strict().DecodeString(key)
	if err != nil {
		return Argon2{}, op.Err(fmt.Errorf("invalid key format: %w", err))
	}
	decoded.Key = bkey

	bsalt, err := base64.RawStdEncoding.Strict().DecodeString(salt)
	if err != nil {
		return Argon2{}, op.Err(fmt.Errorf("invalid salt format: %w", err))
	}
	decoded.Salt = bsalt

	return decoded, nil
}
