// Copyright 2024 lucdoe
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
