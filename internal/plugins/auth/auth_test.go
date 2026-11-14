// Copyright 2024 lucdoe
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var secret = []byte("secret")

func newAuth(t *testing.T, config JWTConfig) *Auth {
	t.Helper()
	config.SecretKey = secret
	config.SigningMethod = jwt.SigningMethodHS256
	auth, err := NewAuthService(config)
	require.NoError(t, err)
	return auth
}

func sign(t *testing.T, method jwt.SigningMethod, claims Claims) string {
	t.Helper()
	if claims.ExpiresAt == nil {
		claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute))
	}
	token, err := jwt.NewWithClaims(method, claims).SignedString(secret)
	require.NoError(t, err)
	return token
}

func TestNewAuthServiceRejectsMissingSettings(t *testing.T) {
	_, err := NewAuthService(JWTConfig{SigningMethod: jwt.SigningMethodHS256})
	assert.Error(t, err, "empty secret")

	_, err = NewAuthService(JWTConfig{SecretKey: secret})
	assert.Error(t, err, "no signing method")
}

func TestValidate(t *testing.T) {
	strict := JWTConfig{Issuer: "issuer", Audience: "gateway", Scope: "read write"}
	claims := func(issuer string, audience []string, scope string) Claims {
		return Claims{
			RegisteredClaims: jwt.RegisteredClaims{Subject: "123", Issuer: issuer, Audience: audience},
			Scope:            scope,
		}
	}

	tests := []struct {
		name    string
		config  JWTConfig
		claims  Claims
		wantErr error // nil: accepted; errAny: rejected for another reason
	}{
		{"matching token", strict, claims("issuer", []string{"gateway"}, "read write"), nil},
		{"extra audiences and scopes", strict, claims("issuer", []string{"other", "gateway"}, "admin write read"), nil},
		{"wrong issuer", strict, claims("someone-else", []string{"gateway"}, "read write"), errAny},
		{"no issuer", strict, claims("", []string{"gateway"}, "read write"), errAny},
		{"wrong audience", strict, claims("issuer", []string{"other"}, "read write"), errAny},
		{"one scope missing", strict, claims("issuer", []string{"gateway"}, "read"), ErrMissingScope},
		{"no scope", strict, claims("issuer", []string{"gateway"}, ""), ErrMissingScope},
		{"nothing required", JWTConfig{}, claims("", nil, ""), nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := newAuth(t, tc.config).Validate(sign(t, jwt.SigningMethodHS256, tc.claims))

			switch tc.wantErr {
			case nil:
				require.NoError(t, err)
				assert.Equal(t, "123", got.Subject)
			case errAny:
				assert.Error(t, err)
				assert.NotErrorIs(t, err, ErrMissingScope)
			default:
				assert.ErrorIs(t, err, tc.wantErr)
			}
		})
	}
}

func TestValidateRejectsBadTokens(t *testing.T) {
	auth := newAuth(t, JWTConfig{})

	_, err := auth.Validate("not-a-token")
	assert.Error(t, err)

	_, err = auth.Validate(sign(t, jwt.SigningMethodHS384, Claims{}))
	assert.Error(t, err, "other signing method")

	expired := Claims{RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute))}}
	_, err = auth.Validate(sign(t, jwt.SigningMethodHS256, expired))
	assert.Error(t, err, "expired")
}

func TestHasScopes(t *testing.T) {
	claims := &Claims{Scope: "read  write"}

	assert.True(t, claims.HasScopes(nil))
	assert.True(t, claims.HasScopes([]string{"write", "read"}))
	assert.False(t, claims.HasScopes([]string{"read", "admin"}))
}
