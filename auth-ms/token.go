package main

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenIssuer = "auth-ms"

// Claims is the JWT payload issued on login.
type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// TokenManager signs and verifies HS256 JWTs with a shared secret.
type TokenManager struct {
	secret []byte
	ttl    time.Duration
}

// NewTokenManager builds a manager. The same secret must be configured on any
// service that verifies these tokens (e.g. the gopdfsuit backend).
func NewTokenManager(secret string, ttl time.Duration) *TokenManager {
	return &TokenManager{secret: []byte(secret), ttl: ttl}
}

// Issue returns a signed JWT for the given user.
func (m *TokenManager) Issue(u User) (string, error) {
	now := time.Now()
	claims := Claims{
		Email: u.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(u.ID, 10),
			Issuer:    tokenIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

// Verify parses and validates a token string, returning its claims.
func (m *TokenManager) Verify(tokenString string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	}, jwt.WithIssuer(tokenIssuer), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	return claims, nil
}
