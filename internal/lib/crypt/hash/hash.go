package hash

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/al-revenko/idp/internal/lib/pkgmark"
	"golang.org/x/crypto/argon2"
)

var pkg = pkgmark.New("crypt/hash")

type DecodedArgon2 struct {
	Algorithm   string
	Key         []byte
	Salt        []byte
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
}

type Argon2 struct {
	Memory        uint32
	Iterations    uint32
	Parallelism   uint8
	KeyLength     uint32
	SaltGenerator func() ([]byte, error)
}

func (a *Argon2) Create(str string) (string, error) {
	op := pkg.Op("Argon2.Create")

	bsalt, err := a.SaltGenerator()
	if err != nil {
		return "", op.Err(fmt.Errorf("generate salt: %w", err))
	}

	bhash := argon2.IDKey(
		[]byte(str),
		bsalt,
		a.Iterations,
		a.Memory,
		a.Parallelism,
		a.KeyLength,
	)

	algStr := "argon2id"
	paramsStr := fmt.Sprintf("i=%d_m=%d_p=%d", a.Iterations, a.Memory, a.Parallelism)

	algBase64 := base64.RawStdEncoding.EncodeToString([]byte(algStr))
	paramsBase64 := base64.RawStdEncoding.EncodeToString([]byte(paramsStr))
	hashBase64 := base64.RawStdEncoding.EncodeToString(bhash)
	saltBase64 := base64.RawStdEncoding.EncodeToString(bsalt)

	return fmt.Sprintf("%s$%s$%s$%s", algBase64, paramsBase64, hashBase64, saltBase64), nil
}

func (a *Argon2) Compare(str, hash string) (bool, error) {
	op := pkg.Op("Argon2.CompareHash")

	decodedHash, err := a.Decode(hash)
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

func (a *Argon2) Decode(hash string) (DecodedArgon2, error) {
	op := pkg.Op("Argon2.Decode")

	parts := strings.Split(hash, "$")
	if len(parts) != 4 {
		return DecodedArgon2{}, op.Err(fmt.Errorf("invalid hash format"))
	}
	decoded := DecodedArgon2{}

	algB, err := base64.RawStdEncoding.Strict().DecodeString(parts[0])
	if err != nil {
		return DecodedArgon2{}, op.Err(fmt.Errorf("invalid algorithm: %w", err))
	}

	algStr := string(algB)

	if algStr != "argon2id" {
		return DecodedArgon2{}, op.Err(fmt.Errorf("invalid algorithm: %s != argon2_%d", algStr, argon2.Version))
	}
	decoded.Algorithm = algStr

	bparams, err := base64.RawStdEncoding.Strict().DecodeString(parts[1])
	if err != nil {
		return DecodedArgon2{}, op.Err(fmt.Errorf("invalid params format: %w", err))
	}

	_, err = fmt.Sscanf(string(bparams), "i=%d_m=%d_p=%d", &decoded.Iterations, &decoded.Memory, &decoded.Parallelism)
	if err != nil {
		return DecodedArgon2{}, op.Err(fmt.Errorf("invalid params format: %w", err))
	}

	bkey, err := base64.RawStdEncoding.Strict().DecodeString(parts[2])
	if err != nil {
		return DecodedArgon2{}, op.Err(fmt.Errorf("invalid key format: %w", err))
	}
	decoded.Key = bkey

	bsalt, err := base64.RawStdEncoding.Strict().DecodeString(parts[3])
	if err != nil {
		return DecodedArgon2{}, op.Err(fmt.Errorf("invalid salt format: %w", err))
	}
	decoded.Salt = bsalt

	return decoded, nil
}
