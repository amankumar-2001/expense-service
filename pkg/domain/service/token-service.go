// Package service holds the expense-service domain logic. Each service is an
// interface (for DI/testability) plus an unexported implementation.
package service

import (
	"crypto/rsa"
	"fmt"
	"os"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
	"github.com/kharchibook/expense-service/config"
	apperrors "github.com/kharchibook/expense-service/errors"
	"github.com/kharchibook/expense-service/pkg/domain/dto/entity"
)

// ITokenService verifies access JWTs (RS256) issued by auth-service. This service
// never mints tokens — it only holds the RSA *public* key and validates
// signature, expiry, and issuer.
type ITokenService interface {
	// ParseAccessToken verifies the signature + expiry and returns the claims.
	ParseAccessToken(token string) (*entity.TokenClaims, error)
	// PublicKeyPEM returns the PEM-encoded verification key (for diagnostics).
	PublicKeyPEM() string
}

// authClaims mirrors the JWT payload shape minted by auth-service.
type authClaims struct {
	Roles    []string `json:"roles"`
	Verified bool     `json:"verified"`
	SID      string   `json:"sid"`
	jwt.RegisteredClaims
}

type tokenService struct {
	cfg       config.Token
	publicKey *rsa.PublicKey
	publicPEM string
}

// NewTokenService loads the RSA public key from configured PEM (inline or path).
// Unlike auth-service it has no private key and cannot sign tokens.
func NewTokenService(cfg config.Token) (ITokenService, error) {
	pemBytes, err := loadPublicKeyPEM(cfg)
	if err != nil {
		return nil, err
	}

	pub, err := jwt.ParseRSAPublicKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	return &tokenService{cfg: cfg, publicKey: pub, publicPEM: string(pemBytes)}, nil
}

func loadPublicKeyPEM(cfg config.Token) ([]byte, error) {
	switch {
	case cfg.PublicKeyPEM != "":
		return []byte(cfg.PublicKeyPEM), nil
	case cfg.PublicKeyPath != "":
		b, err := os.ReadFile(cfg.PublicKeyPath)
		if err != nil {
			return nil, fmt.Errorf("read public key %q: %w", cfg.PublicKeyPath, err)
		}
		return b, nil
	default:
		return nil, fmt.Errorf("no JWT public key configured (set token.publicKeyPath or token.publicKeyPEM)")
	}
}

func (s *tokenService) ParseAccessToken(token string) (*entity.TokenClaims, error) {
	var c authClaims
	parsed, err := jwt.ParseWithClaims(token, &c, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.publicKey, nil
	}, jwt.WithIssuer(s.cfg.Issuer))
	if err != nil || !parsed.Valid {
		return nil, apperrors.UnauthorizedError("invalid or expired token")
	}

	userID, _ := strconv.ParseInt(c.Subject, 10, 64)
	sid, _ := strconv.ParseInt(c.SID, 10, 64)
	out := &entity.TokenClaims{
		UserID:    userID,
		SessionID: sid,
		Roles:     c.Roles,
		Verified:  c.Verified,
	}
	if c.IssuedAt != nil {
		out.IssuedAt = c.IssuedAt.Time
	}
	if c.ExpiresAt != nil {
		out.ExpiresAt = c.ExpiresAt.Time
	}
	return out, nil
}

func (s *tokenService) PublicKeyPEM() string { return s.publicPEM }
