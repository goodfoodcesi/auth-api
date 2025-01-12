package jwt

import (
	"github.com/golang-jwt/jwt/v5"
	db "github.com/goodfoodcesi/auth-api/infrastructure/database/sqlc"
	"time"
)

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	Role   db.UserRole `json:"role"`
	UserID string      `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	accessTokenSecret  []byte
	refreshTokenSecret []byte
	accessTokenTTL     time.Duration
	refreshTokenTTL    time.Duration
}

func NewTokenManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		accessTokenSecret:  []byte(accessSecret),
		refreshTokenSecret: []byte(refreshSecret),
		accessTokenTTL:     accessTTL,
		refreshTokenTTL:    refreshTTL,
	}
}

func (tm *TokenManager) GenerateTokenPair(userID string, role db.UserRole) (*TokenPair, error) {
	claims := Claims{
		role,
		userID,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tm.accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "Goodfood",
			Audience:  []string{userID},
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessTokenString, err := accessToken.SignedString(tm.accessTokenSecret)
	if err != nil {
		return nil, err
	}

	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(tm.refreshTokenTTL))
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	refreshTokenString, err := refreshToken.SignedString(tm.refreshTokenSecret)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
	}, nil
}

func (tm *TokenManager) ValidateToken(tokenString string, isRefresh bool) (*Claims, error) {
	secret := tm.accessTokenSecret
	if isRefresh {
		secret = tm.refreshTokenSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}
