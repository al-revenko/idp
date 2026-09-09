package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/al-revenko/idp/internal/domain"
	"github.com/al-revenko/idp/internal/domain/model"
	"github.com/al-revenko/idp/internal/lib/meta"
	"github.com/redis/go-redis/v9"
)

type Store struct {
	redis redis.UniversalClient
}

func NewStore(rdb redis.UniversalClient) *Store {
	return &Store{redis: rdb}
}

func (s *Store) WithinTransaction(ctx context.Context, cb func(ctx context.Context) error) error {
	op := pkg.Op("Store.WithinTransaction")

	tx := s.redis.TxPipeline()
	defer tx.Discard()

	txCtx := injectTx(ctx, tx)
	err := cb(txCtx)
	if err != nil {
		return op.Err(err)
	}

	_, err = tx.Exec(txCtx)
	if err != nil {
		return op.Err(err)
	}

	return nil
}

func (s *Store) SetRefreshToken(ctx context.Context, tokenHash string, tokenData model.RefreshToken, ttl *time.Duration) error {
	op := pkg.Op("Store.SetRefreshToken")

	txQuery := func(pipe redis.Cmdable) error {
		rKey := rKeyRefreshToken(tokenHash)
		pipe.Set(ctx, rKey, tokenData, redis.KeepTTL)

		rKeyFamily := rKeyRefreshTokenFamilyRevoked(tokenData.FamilyId)
		pipe.Set(ctx, rKeyFamily, tokenData.Revoked, redis.KeepTTL)

		if ttl != nil {
			pipe.Expire(ctx, rKey, *ttl)
			pipe.Expire(ctx, rKeyFamily, *ttl)
		}

		return nil
	}

	if inTx(ctx) {
		err := txQuery(s.db(ctx))

		if err != nil {
			return op.Err(err)
		}

		return nil
	} else {
		_, err := s.db(ctx).TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			return txQuery(pipe)
		})

		if err != nil {
			return op.Err(err)
		}

		return nil
	}

}

func (s *Store) GetRefreshToken(ctx context.Context, tokenHash string) (model.RefreshToken, error) {
	op := pkg.Op("Store.GetRefreshToken")
	rKeyToken := rKeyRefreshToken(tokenHash)

	var token model.RefreshToken
	err := s.db(ctx).Get(ctx, rKeyToken).Scan(&token)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.RefreshToken{}, op.Err(domain.Error(domain.CodeNotFound, "refresh token doesn't exists", err))
		}

		return model.RefreshToken{}, op.Err(err)
	}

	rKeyFamily := rKeyRefreshTokenFamilyRevoked(token.FamilyId)
	familyRevoked, err := s.db(ctx).Get(ctx, rKeyFamily).Bool()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.RefreshToken{}, op.Err(domain.Error(domain.CodeInternal, "family doesn't exists", err))
		}

		return model.RefreshToken{}, op.Err(err)
	}

	token.Revoked = familyRevoked

	return token, nil
}

func (s *Store) db(ctx context.Context) redis.Cmdable {
	if tx := extractTx(ctx); tx != nil {
		return tx
	}

	return s.redis
}

func inTx(ctx context.Context) bool {
	return extractTx(ctx) != nil
}

func injectTx(ctx context.Context, tx redis.Pipeliner) context.Context {
	return context.WithValue(ctx, meta.CtxRedisTxKey{}, tx)
}

func extractTx(ctx context.Context) redis.Pipeliner {
	if tx, ok := ctx.Value(meta.CtxRedisTxKey{}).(redis.Pipeliner); ok {
		return tx
	}
	return nil
}

func rKeyRefreshToken(tokenHash string) string {
	return fmt.Sprintf("refresh_token:%s", tokenHash)
}

func rKeyRefreshTokenFamilyRevoked(familyId string) string {
	return fmt.Sprintf("refresh_token:family:%s:revoked", familyId)
}
