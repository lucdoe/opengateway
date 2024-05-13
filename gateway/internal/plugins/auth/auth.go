package auth

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type AuthInterface interface {
	Validate(token string) (jwt.Claims, error)
	ParseToken(tokenStr string) (*jwt.RegisteredClaims, error)
}

type Auth struct {
	config JWTConfig
}

type JWTConfig struct {
	SecretKey     []byte
	SigningMethod jwt.SigningMethod
	Issuer        string
	Audience      string
	Scope         string
}

func NewAuthService(config JWTConfig) *Auth {
	if len(config.SecretKey) == 0 {
		panic("secret key must not be empty")
	}
	if config.SigningMethod == nil {
		panic("signing method must not be nil")
	}
	return &Auth{config: config}
}

func (j *Auth) Validate(tokenStr string) (jwt.Claims, error) {
	claims, err := j.ParseToken(tokenStr)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

func (j *Auth) ParseToken(tokenStr string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != j.config.SigningMethod {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.config.SecretKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	audienceStr := strings.Join(claims.Audience, ", ")
	audienceMatch := audienceStr == j.config.Audience
	if !audienceMatch {
		return nil, fmt.Errorf("invalid audience")
	}

	return claims, nil
}
