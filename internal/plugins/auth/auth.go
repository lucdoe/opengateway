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
	"errors"
	"slices"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// ErrMissingScope means the token is valid but lacks a required scope.
var ErrMissingScope = errors.New("token lacks a required scope")

// Claims are the token claims the gateway reads. Scope is space-separated,
// as in OAuth 2.0.
type Claims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope"`
}

// HasScopes reports whether the token carries every scope in required.
func (c *Claims) HasScopes(required []string) bool {
	granted := strings.Fields(c.Scope)
	for _, scope := range required {
		if !slices.Contains(granted, scope) {
			return false
		}
	}
	return true
}

// JWTConfig says which tokens are accepted. Issuer, Audience and Scope are
// only checked when set; Scope lists space-separated scopes every token needs.
type JWTConfig struct {
	SecretKey     []byte
	SigningMethod jwt.SigningMethod
	Issuer        string
	Audience      string
	Scope         string
}

type Auth struct {
	secret []byte
	parser *jwt.Parser
	scopes []string
}

func NewAuthService(config JWTConfig) (*Auth, error) {
	if len(config.SecretKey) == 0 {
		return nil, errors.New("secret key must not be empty")
	}
	if config.SigningMethod == nil {
		return nil, errors.New("signing method must not be nil")
	}
	options := []jwt.ParserOption{jwt.WithValidMethods([]string{config.SigningMethod.Alg()})}
	if config.Issuer != "" {
		options = append(options, jwt.WithIssuer(config.Issuer))
	}
	if config.Audience != "" {
		options = append(options, jwt.WithAudience(config.Audience))
	}
	return &Auth{
		secret: config.SecretKey,
		parser: jwt.NewParser(options...),
		scopes: strings.Fields(config.Scope),
	}, nil
}

// Validate checks the token's signature, expiry, issuer and audience, then
// that it carries the configured scopes, which fails with ErrMissingScope.
func (a *Auth) Validate(token string) (*Claims, error) {
	claims := &Claims{}
	if _, err := a.parser.ParseWithClaims(token, claims, func(*jwt.Token) (any, error) {
		return a.secret, nil
	}); err != nil {
		return nil, err
	}
	if !claims.HasScopes(a.scopes) {
		return nil, ErrMissingScope
	}
	return claims, nil
}
