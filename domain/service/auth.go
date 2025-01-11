package service

import (
	"context"
	"errors"

	"github.com/goodfoodcesi/auth-api/infrastructure/jwt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AuthService struct {
	userService *UserService
	tokenMgr    *jwt.TokenManager
}

func NewAuthService(userService *UserService, tokenMgr *jwt.TokenManager) *AuthService {
	return &AuthService{
		userService: userService,
		tokenMgr:    tokenMgr,
	}
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	claims, err := s.tokenMgr.ValidateToken(refreshToken, true)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	user, err := s.userService.GetByID(ctx, pgtype.UUID{Bytes: uuid.MustParse(claims.UserID), Valid: true})
	if err != nil {
		return nil, err
	}

	return s.tokenMgr.GenerateTokenPair(user.ID.String(), user.Role)
}
