package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/auth"
	"github.com/stretchr/testify/assert"
)

func TestTokenValidation(t *testing.T) {
	config := auth.JWTConfig{
		SecretKey:     []byte("test_secret"),
		SigningMethod: jwt.SigningMethodHS256,
		Audience:      "test_audience",
	}
	service, err := auth.NewAuthService(config)
	assert.NoError(t, err, "Failed to create AuthService")

	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "test_issuer",
		Audience:  []string{config.Audience},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		Subject:   "123",
	})
	validTokenStr, _ := validToken.SignedString(config.SecretKey)

	invalidSignatureToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Issuer:    "test_issuer",
		Audience:  []string{"test_audience"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		Subject:   "123",
	})
	invalidSignatureTokenStr, _ := invalidSignatureToken.SignedString(config.SecretKey)

	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(-10 * time.Minute)),
	})
	expiredTokenStr, _ := expiredToken.SignedString(config.SecretKey)

	wrongAudienceToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Audience: []string{"another_audience"},
	})
	wrongAudienceTokenStr, _ := wrongAudienceToken.SignedString(config.SecretKey)

	testCases := []struct {
		name       string
		tokenStr   string
		wantErr    bool
		errMessage string
	}{
		{
			name:     "Valid token",
			tokenStr: validTokenStr,
			wantErr:  false,
		},
		{
			name:       "Invalid signature method",
			tokenStr:   invalidSignatureTokenStr,
			wantErr:    true,
			errMessage: "unexpected signing method",
		},
		{
			name:       "Expired token",
			tokenStr:   expiredTokenStr,
			wantErr:    true,
			errMessage: "token is expired",
		},
		{
			name:       "Invalid audience",
			tokenStr:   wrongAudienceTokenStr,
			wantErr:    true,
			errMessage: "invalid audience",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := service.ParseToken(tc.tokenStr)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMessage != "" {
					assert.Contains(t, err.Error(), tc.errMessage, "Error message should match")
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims, "Claims should not be nil for valid tokens")
			}
		})
	}
}

func TestNewAuthService(t *testing.T) {
	testCases := []struct {
		name        string
		config      auth.JWTConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Empty Secret Key",
			config: auth.JWTConfig{
				SecretKey:     []byte(""),
				SigningMethod: jwt.SigningMethodHS256,
			},
			expectError: true,
			errorMsg:    "secret key must not be empty",
		},
		{
			name: "Nil Signing Method",
			config: auth.JWTConfig{
				SecretKey:     []byte("valid_secret"),
				SigningMethod: nil,
			},
			expectError: true,
			errorMsg:    "signing method must not be nil",
		},
		{
			name: "Valid Configuration",
			config: auth.JWTConfig{
				SecretKey:     []byte("valid_secret"),
				SigningMethod: jwt.SigningMethodHS256,
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, err := auth.NewAuthService(tc.config)
			if tc.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg, "Error message should match")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, service, "Service should not be nil when created with valid configuration")
			}
		})
	}
}

func TestValidate(t *testing.T) {
	config := auth.JWTConfig{
		SecretKey:     []byte("valid_secret"),
		SigningMethod: jwt.SigningMethodHS256,
		Audience:      "test_audience",
	}
	service, _ := auth.NewAuthService(config)

	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "test_issuer",
		Audience:  []string{config.Audience},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 5)),
	})
	validTokenStr, _ := validToken.SignedString(config.SecretKey)

	invalidTokenStr := "invalid.token.string"

	testCases := []struct {
		name        string
		tokenStr    string
		expectError bool
	}{
		{
			name:        "Valid Token",
			tokenStr:    validTokenStr,
			expectError: false,
		},
		{
			name:        "Invalid Token Format",
			tokenStr:    invalidTokenStr,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := service.Validate(tc.tokenStr)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}
