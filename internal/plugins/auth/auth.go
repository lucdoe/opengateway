// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

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
	// Keys returns the key, or keys, that may have signed a token: Secret
	// for a shared secret, JWKS.Keyfunc for an identity provider's keys.
	Keys jwt.Keyfunc
	// Algorithms are the accepted alg values. A token signed with any other
	// is rejected before Keys is asked.
	Algorithms []string
	Issuer     string
	Audience   string
	Scope      string
}

// Secret returns a Keys function for tokens signed with a shared secret.
func Secret(secret []byte) jwt.Keyfunc {
	return func(*jwt.Token) (any, error) { return secret, nil }
}

type Auth struct {
	keys   jwt.Keyfunc
	parser *jwt.Parser
	scopes []string
}

func NewAuthService(config JWTConfig) (*Auth, error) {
	if config.Keys == nil {
		return nil, errors.New("keys must not be nil")
	}
	if len(config.Algorithms) == 0 {
		return nil, errors.New("at least one algorithm is required")
	}
	options := []jwt.ParserOption{
		jwt.WithValidMethods(config.Algorithms),
		// A token without exp would stay valid forever once leaked.
		jwt.WithExpirationRequired(),
	}
	if config.Issuer != "" {
		options = append(options, jwt.WithIssuer(config.Issuer))
	}
	if config.Audience != "" {
		options = append(options, jwt.WithAudience(config.Audience))
	}
	return &Auth{
		keys:   config.Keys,
		parser: jwt.NewParser(options...),
		scopes: strings.Fields(config.Scope),
	}, nil
}

// Validate checks the token's signature, required expiry, issuer and audience, then
// that it carries the configured scopes, which fails with ErrMissingScope.
func (a *Auth) Validate(token string) (*Claims, error) {
	claims := &Claims{}
	if _, err := a.parser.ParseWithClaims(token, claims, a.keys); err != nil {
		return nil, err
	}
	if !claims.HasScopes(a.scopes) {
		return nil, ErrMissingScope
	}
	return claims, nil
}
