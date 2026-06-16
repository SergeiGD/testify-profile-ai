package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

const tokenByteLen = 32

type TokenService interface {
	Generate() (rawToken string, tokenHash string, err error)
	Hash(rawToken string) string
}

type tokenService struct{}

func NewTokenService() TokenService {
	return &tokenService{}
}

func (t *tokenService) Generate() (string, string, error) {
	b := make([]byte, tokenByteLen)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	raw := hex.EncodeToString(b)
	return raw, t.Hash(raw), nil
}

func (t *tokenService) Hash(rawToken string) string {
	h := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(h[:])
}
