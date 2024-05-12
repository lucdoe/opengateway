package auth

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type AuthInterface interface {
	Validate(token string) (jwt.Claims, error)
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

func NewAuthService(config JWTConfig) (*Auth, error) {
	if len(config.SecretKey) == 0 {
		return nil, fmt.Errorf("secret key must not be empty")
	}
	if config.SigningMethod == nil {
		return nil, fmt.Errorf("signing method must not be nil")
	}
	return &Auth{config: config}, nil
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
		if token.Method != jwt.SigningMethodHS256 {
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
	return claims, nil
}
