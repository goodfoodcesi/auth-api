package service

import (
	"context"
	"errors"
	"time"

	"github.com/goodfoodcesi/auth-api/crypto"
	"github.com/goodfoodcesi/auth-api/domain/entity"
	"github.com/goodfoodcesi/auth-api/domain/repository"
	"github.com/goodfoodcesi/auth-api/infrastructure/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserService struct {
	repo       repository.UserRepository
	tokenMgr   *jwt.TokenManager
	pwdManager *crypto.PasswordManager
}

type RegisterUserInput struct {
	FirstName string      `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string      `json:"last_name" validate:"required,min=2,max=100"`
	Email     string      `json:"email" validate:"required,email"`
	Password  string      `json:"password" validate:"required,min=8"`
	Role      entity.Role `json:"role" validate:"required,oneof=client manager driver"`
}

type UpdateUserInput struct {
	FirstName string `json:"first_name" validate:"omitempty,min=2,max=100"`
	LastName  string `json:"last_name" validate:"omitempty,min=2,max=100"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func NewUserService(repo repository.UserRepository, tokenMgr *jwt.TokenManager, pwdManager *crypto.PasswordManager) *UserService {
	return &UserService{
		repo:       repo,
		tokenMgr:   tokenMgr,
		pwdManager: pwdManager,
	}
}

func (s *UserService) Update(ctx context.Context, id pgtype.UUID, input UpdateUserInput) (*entity.User, error) {
	existingUser, _ := s.repo.GetByID(ctx, id)
	if existingUser == nil {
		return nil, errors.New("user not found")
	}

	user := &entity.User{
		ID:        existingUser.ID,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Email:     existingUser.Email,
		Role:      existingUser.Role,
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: time.Now(),
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Register(ctx context.Context, input RegisterUserInput) (*entity.User, error) {
	existingUser, _ := s.repo.GetByEmail(ctx, input.Email)
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := s.pwdManager.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		ID:           uuid.New(),
		FirstName:    input.FirstName,
		LastName:     input.LastName,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Role:         input.Role,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Login(ctx context.Context, input LoginInput) (*jwt.TokenPair, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, err
	}

	if !s.pwdManager.ComparePassword(user.PasswordHash, input.Password) {
		return nil, errors.New("invalid credentials")
	}

	tokens, err := s.tokenMgr.GenerateTokenPair(user.ID.String(), user.Role)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}

func (s *UserService) GetByID(ctx context.Context, id pgtype.UUID) (*entity.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}
