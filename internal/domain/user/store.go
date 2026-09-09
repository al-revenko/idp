package user

import (
	"context"
	"errors"

	"github.com/al-revenko/idp/db/sql/gen/dbstore"
	"github.com/al-revenko/idp/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Store struct {
	db *dbstore.Queries
}

func NewStore(db *dbstore.Queries) *Store {
	return &Store{db: db}
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string) (userId string, err error) {
	op := pkg.Op("Store.CreateUser")

	userId, err = s.db.CreateUser(ctx, dbstore.CreateUserParams{
		Username:     username,
		PasswordHash: passwordHash,
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
			return "", op.Err(domain.Error(domain.CodeConflict, "username already taken", err))
		}
		return "", op.Err(domain.Error(domain.CodeInternal, err.Error(), err))
	}
	return userId, nil
}

func (s *Store) GetUserByUsername(ctx context.Context, username string) (domain.User, error) {
	op := pkg.Op("Store.GetUserByUsername")

	u, err := s.db.GetUserByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, op.Err(domain.Error(domain.CodeNotFound, "user not found", err))
		}

		return domain.User{}, op.Err(domain.Error(domain.CodeInternal, err.Error(), err))
	}

	return s.toDomainUser(u), nil
}

func (s *Store) toDomainUser(u dbstore.User) domain.User {
	return domain.User{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
	}
}
