package repository

import (
	"context"

	"github.com/goodfoodcesi/auth-api/infrastructure/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	Create(ctx context.Context, user *db.User) (*db.User, error)
	GetByEmail(ctx context.Context, email string) (*db.User, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*db.User, error)
	Update(ctx context.Context, user *db.User) error
}
