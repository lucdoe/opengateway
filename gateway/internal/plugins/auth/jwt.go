package pluginJWT

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

type TokenService interface {
	Validate(tokenStr string) (jwt.Claims, error)
}

type JWTConfig struct {
	SecretKey     []byte
	SigningMethod jwt.SigningMethod
	Issuer        string
	Audience      string
	Scope         string
}

type JWTService struct {
	config JWTConfig
}

func NewJWTService(config JWTConfig) TokenService {
	return &JWTService{
		config: config,
	}
}

func (j *JWTService) Validate(tokenStr string) (jwt.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.config.SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok && !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
