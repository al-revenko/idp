package blake3

import (
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/al-revenko/idp/internal/lib/crypt/hash"
	"github.com/al-revenko/idp/internal/lib/sign"
	"github.com/zeebo/blake3"
)

var pkg = sign.Pkg("hash/blake3")

const Algorithm = "blake3"

type Blake3 struct {
	hash.Hash
}

type Hasher struct {
	SaltGenerator func() ([]byte, error)
}

func (h *Hasher) Hash(data string, salt []byte) (string, error) {
	op := pkg.Op("Hasher.Hash")

	saltValue := salt
	if saltValue == nil {
		var err error
		saltValue, err = h.SaltGenerator()
		if err != nil {
			return "", op.Err(err)
		}
	}

	hasher := blake3.New()
	hasher.WriteString(data)
	hasher.Write(saltValue)
	hash := hasher.Sum(nil)

	var builder strings.Builder
	builder.WriteString(Algorithm)
	builder.WriteString("$")
	builder.WriteString(base64.RawStdEncoding.EncodeToString(saltValue))
	builder.WriteString("$")
	builder.WriteString(base64.RawStdEncoding.EncodeToString(hash))

	return builder.String(), nil
}

func (h *Hasher) Compare(str, hashStr string) (bool, error) {
	op := pkg.Op("Hasher.Compare")

	decoded, err := h.Decode(hashStr)
	if err != nil {
		return false, op.Err(err)
	}

	hasher := blake3.New()
	hasher.WriteString(str)
	hasher.Write(decoded.Salt)
	computedDigest := hasher.Sum(nil)

	keyLen := int32(len(decoded.Key))
	otherKeyLen := int32(len(computedDigest))

	if subtle.ConstantTimeEq(keyLen, otherKeyLen) == 0 {
		subtle.ConstantTimeCompare(decoded.Key, decoded.Key)
		return false, nil
	}

	isEqual := subtle.ConstantTimeCompare(decoded.Key, computedDigest) == 1
	return isEqual, nil
}

func (h *Hasher) Decode(hashStr string) (Blake3, error) {
	op := pkg.Op("Hasher.Decode")

	parts := strings.Split(hashStr, "$")
	if len(parts) != 3 {
		return Blake3{}, op.Err(fmt.Errorf("invalid hash format"))
	}

	alg := parts[0]
	salt := parts[1]
	key := parts[2]

	if alg != Algorithm {
		return Blake3{}, op.Err(fmt.Errorf("invalid algorithm: %s != %s", alg, Algorithm))
	}

	decoded := Blake3{}
	decoded.Algorithm = alg

	bsalt, err := base64.RawStdEncoding.Strict().DecodeString(salt)
	if err != nil {
		return Blake3{}, op.Err(fmt.Errorf("invalid salt format: %w", err))
	}
	decoded.Salt = bsalt

	bkey, err := base64.RawStdEncoding.Strict().DecodeString(key)
	if err != nil {
		return Blake3{}, op.Err(fmt.Errorf("invalid key format: %w", err))
	}
	decoded.Key = bkey

	return decoded, nil
}
