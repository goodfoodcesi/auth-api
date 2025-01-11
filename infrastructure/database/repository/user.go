package repository

import (
	"context"

	"github.com/goodfoodcesi/auth-api/domain/entity"
	sqlc "github.com/goodfoodcesi/auth-api/infrastructure/database/sqlc"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
	q  *sqlc.Queries
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
		q:  sqlc.New(db),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	_, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           pgtype.UUID{Bytes: user.ID, Valid: true},
		Firstname:    user.FirstName,
		Lastname:     user.LastName,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		Role:         sqlc.UserRole(user.Role),
	})
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	dbUser, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return mapDBUserToEntity(dbUser), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id pgtype.UUID) (*entity.User, error) {
	dbUser, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return mapDBUserToEntity(dbUser), nil
}

func (r *UserRepository) Update(ctx context.Context, user *entity.User) error {
	err := r.q.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:        pgtype.UUID{Bytes: user.ID, Valid: true},
		Firstname: user.FirstName,
		Lastname:  user.LastName,
	})
	if err != nil {
		return err
	}
	return nil
}

func mapDBUserToEntity(dbUser sqlc.User) *entity.User {
	return &entity.User{
		ID:           dbUser.ID.Bytes,
		FirstName:    dbUser.Firstname,
		LastName:     dbUser.Lastname,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		Role:         entity.Role(dbUser.Role),
		CreatedAt:    dbUser.CreatedAt.Time,
		UpdatedAt:    dbUser.UpdatedAt.Time,
	}
}
