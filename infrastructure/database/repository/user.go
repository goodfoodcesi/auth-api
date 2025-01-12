package repository

import (
	"context"

	"github.com/goodfoodcesi/auth-api/infrastructure/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	dbPool *pgxpool.Pool
	q      *db.Queries
}

func NewUserRepository(dbPool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		dbPool: dbPool,
		q:      db.New(dbPool),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *db.User) (*db.User, error) {
	dbUser, err := r.q.CreateUser(ctx, db.CreateUserParams{
		Firstname:    user.Firstname,
		Lastname:     user.Lastname,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         user.Role,
	})

	if err != nil {
		return nil, err
	}

	return mapDBUserToEntity(dbUser), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	dbUser, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return mapDBUserToEntity(dbUser), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id pgtype.UUID) (*db.User, error) {
	dbUser, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapDBUserToEntity(dbUser), nil
}

func (r *UserRepository) Update(ctx context.Context, user *db.User) error {
	err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:        user.ID,
		Firstname: user.Firstname,
		Lastname:  user.Lastname,
	})
	if err != nil {
		return err
	}
	return nil
}

func mapDBUserToEntity(dbUser db.User) *db.User {
	return &db.User{
		ID:           dbUser.ID,
		Firstname:    dbUser.Firstname,
		Lastname:     dbUser.Lastname,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		Role:         dbUser.Role,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
	}
}
