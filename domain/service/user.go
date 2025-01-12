package service

import (
	"context"
	"errors"
	"github.com/goodfoodcesi/api-utils-go/pkg/event"
	"go.uber.org/zap"
	"time"

	"github.com/goodfoodcesi/auth-api/crypto"
	"github.com/goodfoodcesi/auth-api/domain/repository"
	"github.com/goodfoodcesi/auth-api/infrastructure/database/sqlc"
	"github.com/goodfoodcesi/auth-api/infrastructure/jwt"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserService struct {
	repo             repository.UserRepository
	tokenMgr         *jwt.TokenManager
	pwdManager       *crypto.PasswordManager
	messagingService *MessagingService
	logger           *zap.Logger
}

type RegisterUserInput struct {
	FirstName string      `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string      `json:"last_name" validate:"required,min=2,max=100"`
	Email     string      `json:"email" validate:"required,email"`
	Password  string      `json:"password" validate:"required,min=8"`
	Role      db.UserRole `json:"role" validate:"required,oneof=client manager driver"`
}

type UpdateUserInput struct {
	FirstName string `json:"first_name" validate:"omitempty,min=2,max=100"`
	LastName  string `json:"last_name" validate:"omitempty,min=2,max=100"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func NewUserService(
	repo repository.UserRepository,
	tokenMgr *jwt.TokenManager,
	pwdManager *crypto.PasswordManager,
	messagingService *MessagingService,
	logger *zap.Logger,
) *UserService {
	return &UserService{
		repo:             repo,
		tokenMgr:         tokenMgr,
		pwdManager:       pwdManager,
		messagingService: messagingService,
		logger:           logger,
	}
}

func (s *UserService) Update(ctx context.Context, id pgtype.UUID, input UpdateUserInput) (*db.User, error) {
	existingUser, _ := s.repo.GetByID(ctx, id)
	if existingUser == nil {
		return nil, errors.New("user not found")
	}

	user := &db.User{
		ID:        existingUser.ID,
		Firstname: input.FirstName,
		Lastname:  input.LastName,
		Email:     existingUser.Email,
		Role:      existingUser.Role,
		CreatedAt: existingUser.CreatedAt,
		UpdatedAt: pgtype.Timestamptz{Time: time.Now()},
	}

	if err := s.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) Register(ctx context.Context, input RegisterUserInput) (*db.User, error) {
	existingUser, _ := s.repo.GetByEmail(ctx, input.Email)
	if existingUser != nil {
		return nil, errors.New("email already exists")
	}

	hashedPassword, err := s.pwdManager.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &db.User{
		Firstname:    input.FirstName,
		Lastname:     input.LastName,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Role:         input.Role,
		CreatedAt:    pgtype.Timestamptz{Time: time.Now()},
		UpdatedAt:    pgtype.Timestamptz{Time: time.Now()},
	}

	user, err = s.repo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	userCreatedEvent := event.UserCreatedEvent{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.Firstname,
		LastName:  user.Lastname,
		Role:      string(user.Role),
		CreatedAt: time.Now(),
	}

	if err := s.messagingService.PublishUserCreated(ctx, userCreatedEvent); err != nil {
		s.logger.Error("failed to publish user created event", zap.Error(err))
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

func (s *UserService) GetByID(ctx context.Context, id pgtype.UUID) (*db.User, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}
