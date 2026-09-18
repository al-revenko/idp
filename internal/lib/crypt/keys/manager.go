package keys

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"errors"
	"time"
	"uuid"

	"github.com/al-revenko/idp/internal/lib/sign"
	"github.com/lestrrat-go/jwx/v4/jwa"
	"github.com/lestrrat-go/jwx/v4/jwk"
	"github.com/lestrrat-go/jwx/v4/jwt"
)

var pkg = sign.Pkg("crypt/keys")

type Key struct {
	Value     *ecdsa.PrivateKey
	Active    bool
	KeyId     string
	CreatedAt time.Time
}

type KeyStore interface {
	SetKeys(keys []Key) error
	GetKeys() ([]Key, error)
	DeleteKeys(keys []Key) error
}

type Manager struct {
	store       KeyStore
	gracePeriod time.Duration
}

func NewManager(store KeyStore, gracePeriod time.Duration) *Manager {
	return &Manager{
		store:       store,
		gracePeriod: gracePeriod,
	}
}

func (m *Manager) RotateKeys() error {
	op := pkg.Op("Manager.RotateKeys")

	now := time.Now()

	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return op.Err(err)
	}
	keys, err := m.store.GetKeys()
	if err != nil {
		return op.Err(err)
	}

	keysToDelete := make([]Key, 0, len(keys))
	keysToKeep := make([]Key, 0, len(keys)+1)

	for _, key := range keys {
		key.Active = false

		if key.CreatedAt.Add(m.gracePeriod).Before(now) {
			keysToDelete = append(keysToDelete, key)
		} else {
			keysToKeep = append(keysToKeep, key)
		}
	}

	keysToKeep = append(keysToKeep, Key{
		Value:     ecdsaKey,
		Active:    true,
		KeyId:     uuid.NewV7().String(),
		CreatedAt: now,
	})

	if err := m.store.SetKeys(keysToKeep); err != nil {
		return op.Err(err)
	}

	if err := m.store.DeleteKeys(keysToDelete); err != nil {
		return op.Err(err)
	}

	return nil
}

func (m *Manager) JWS(token *jwt.Token) ([]byte, error) {
	op := pkg.Op("Manager.JWS")

	if token == nil {
		return nil, op.Err(errors.New("token is nil"))
	}

	keys, err := m.store.GetKeys()
	if err != nil {
		return nil, op.Err(err)
	}
	privateKey := getActiveKey(keys)
	if privateKey == nil {
		return nil, op.Err(errors.New("no active key"))
	}

	signedToken, err := jwt.Sign(*token, jwt.WithKey(jwa.ES256(), privateKey))
	if err != nil {
		return nil, op.Err(err)
	}
	return signedToken, nil
}

func (m *Manager) JWKS() (jwk.Set, error) {
	op := pkg.Op("Manager.JWKS")

	privateKeys, err := m.store.GetKeys()
	if err != nil {
		return nil, op.Err(err)
	}

	jwks := jwk.NewSet()
	for _, k := range privateKeys {
		pubJwk, err := jwk.PublicKeyOf(k.Value)
		if err != nil {
			return nil, op.Err(err)
		}

		pubJwk.Set(jwk.AlgorithmKey, jwa.ES256())
		pubJwk.Set(jwk.KeyIDKey, k.KeyId)
		pubJwk.Set(jwk.KeyUsageKey, "sig")

		if err := jwks.AddKey(pubJwk); err != nil {
			return nil, err
		}
	}

	return jwks, nil
}

func getActiveKey(keys []Key) *ecdsa.PrivateKey {
	for _, k := range keys {
		if k.Active {
			return k.Value
		}
	}

	return nil
}
