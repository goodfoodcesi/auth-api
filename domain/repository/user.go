package repository

import (
	"context"

	"github.com/goodfoodcesi/auth-api/domain/entity"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
}
