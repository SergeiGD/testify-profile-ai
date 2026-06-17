package jwt

import (
	"crypto/rsa"
	"time"

	"github.com/SergeiGD/testify-profile/internal/domain"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

type AccessClaims struct {
	jwt.RegisteredClaims
	TokenType string    `json:"token_type"`
	UserID    uuid.UUID `json:"user_id"`
	Username  string    `json:"username"`
	BirthDate string    `json:"birth_date"`
}

type RefreshClaims struct {
	jwt.RegisteredClaims
	TokenType string    `json:"token_type"`
	UserID    uuid.UUID `json:"user_id"`
}

type JWTService interface {
	GenerateAccessToken(userID uuid.UUID, username, birthDate string) (string, error)
	GenerateRefreshToken(userID uuid.UUID) (string, error)
	ParseRefreshToken(tokenStr string) (*RefreshClaims, error)
}

type jwtService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTService(privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, accessTTL, refreshTTL time.Duration) JWTService {
	return &jwtService{
		privateKey: privateKey,
		publicKey:  publicKey,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *jwtService) GenerateAccessToken(userID uuid.UUID, username, birthDate string) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessTTL)),
		},
		TokenType: tokenTypeAccess,
		UserID:    userID,
		Username:  username,
		BirthDate: birthDate,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *jwtService) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshTTL)),
		},
		TokenType: tokenTypeRefresh,
		UserID:    userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *jwtService) ParseRefreshToken(tokenStr string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, domain.ErrInvalidRefreshToken
		}
		return s.publicKey, nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || claims.TokenType != tokenTypeRefresh {
		return nil, domain.ErrInvalidRefreshToken
	}

	return claims, nil
}
