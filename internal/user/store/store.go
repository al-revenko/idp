package store

import (
	"context"
	"errors"

	"github.com/al-revenko/idp/db/sql/gen/dbstore"
	"github.com/al-revenko/idp/internal/lib/apperr"
	"github.com/al-revenko/idp/internal/lib/pkgmark"
	"github.com/al-revenko/idp/internal/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var pkg = pkgmark.New("user/store")

type Store struct {
	queries *dbstore.Queries
}

func New(db dbstore.DBTX) *Store {
	return &Store{queries: dbstore.New(db)}
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string) (userId string, err error) {
	op := pkg.Op("Store.CreateUser")

	userId, err = s.queries.CreateUser(ctx, dbstore.CreateUserParams{
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return "", op.Err(apperr.From(apperr.CodeConflict, "username already taken", err))
		}
		return "", op.Err(apperr.From(apperr.CodeInternal, err.Error(), err))
	}
	return userId, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (user.UserWithPasswordHash, error) {
	op := pkg.Op("Store.GetUserByUsername")

	u, err := s.queries.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.UserWithPasswordHash{}, op.Err(apperr.From(apperr.CodeNotFound, "user not found", err))
		}

		return user.UserWithPasswordHash{}, op.Err(apperr.From(apperr.CodeInternal, err.Error(), err))
	}

	return toDomainUser(u), nil
}

func toDomainUser(u dbstore.User) user.UserWithPasswordHash {
	return user.UserWithPasswordHash{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
	}
}
