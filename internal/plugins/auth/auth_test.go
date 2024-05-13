package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const (
	issuer   = "test_issuer"
	audience = "test_audience"
	scope    = "test_scope"
)

func TestNewAuthService(t *testing.T) {
	config := JWTConfig{
		SecretKey:     []byte("secret"),
		SigningMethod: jwt.SigningMethodHS256,
		Issuer:        issuer,
		Audience:      audience,
		Scope:         scope,
	}

	auth := NewAuthService(config)

	assert.NotNil(t, auth)
	assert.Equal(t, config.SecretKey, auth.config.SecretKey)
	assert.Equal(t, config.SigningMethod, auth.config.SigningMethod)
	assert.Equal(t, config.Issuer, auth.config.Issuer)
	assert.Equal(t, config.Audience, auth.config.Audience)
	assert.Equal(t, config.Scope, auth.config.Scope)
}

func TestNewAuthServiceInvalidSecretKey(t *testing.T) {
	config := JWTConfig{
		SecretKey:     []byte(""),
		SigningMethod: jwt.SigningMethodHS256,
		Issuer:        issuer,
		Audience:      audience,
		Scope:         scope,
	}

	assert.Panics(t, func() {
		NewAuthService(config)
	})
}

func TestNewAuthServiceInvalidSigningMethod(t *testing.T) {
	config := JWTConfig{
		SecretKey:     []byte("secret"),
		SigningMethod: nil,
		Issuer:        issuer,
		Audience:      audience,
		Scope:         scope,
	}

	assert.Panics(t, func() {
		NewAuthService(config)
	})
}

func TestAuthValidate(t *testing.T) {
	config := JWTConfig{
		SecretKey:     []byte("secret"),
		SigningMethod: jwt.SigningMethodHS256,
		Issuer:        issuer,
		Audience:      audience,
		Scope:         scope,
	}

	auth := NewAuthService(config)

	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    issuer,
		Audience:  []string{config.Audience},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		Subject:   "123",
	})
	validTokenStr, _ := validToken.SignedString(config.SecretKey)

	c, err := auth.Validate(validTokenStr)

	assert.NoError(t, err)
	assert.NotNil(t, c)
}

func TestAuthValidateInvalidToken(t *testing.T) {
	config := JWTConfig{
		SecretKey:     []byte("secret"),
		SigningMethod: jwt.SigningMethodHS256,
		Issuer:        issuer,
		Audience:      audience,
		Scope:         scope,
	}

	auth := NewAuthService(config)

	tokenStr := "invalid-token"

	_, err := auth.Validate(tokenStr)

	assert.Error(t, err)
}

func TestAuthParseToken(t *testing.T) {
	config := JWTConfig{
		SecretKey:     []byte("secret"),
		SigningMethod: jwt.SigningMethodHS256,
		Issuer:        issuer,
		Audience:      audience,
		Scope:         scope,
	}

	auth := NewAuthService(config)

	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:   issuer,
		Audience: []string{config.Audience},
		Subject:  "123",
	})
	validTokenStr, _ := validToken.SignedString(config.SecretKey)

	claims, err := auth.ParseToken(validTokenStr)

	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, "123", claims.Subject)
}

func TestAuthParseTokenInvalidToken(t *testing.T) {
	config := JWTConfig{
		SecretKey:     []byte("secret"),
		SigningMethod: jwt.SigningMethodHS256,
		Issuer:        issuer,
		Audience:      audience,
		Scope:         scope,
	}

	auth := NewAuthService(config)

	tokenStr := "invalid-token"

	_, err := auth.ParseToken(tokenStr)

	assert.Error(t, err)
}

func TestAuthParseTokenInvalidAudience(t *testing.T) {
	config := JWTConfig{
		SecretKey:     []byte("secret"),
		SigningMethod: jwt.SigningMethodHS256,
		Issuer:        issuer,
		Audience:      audience,
		Scope:         scope,
	}

	auth := NewAuthService(config)

	wrongAudienceToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Audience: []string{"another_audience"},
	})
	wrongAudienceTokenStr, _ := wrongAudienceToken.SignedString(config.SecretKey)

	_, err := auth.ParseToken(wrongAudienceTokenStr)

	assert.Error(t, err)
}
