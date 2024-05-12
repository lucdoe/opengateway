package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lucdoe/open-gateway/gateway/internal/plugins/auth"
	"github.com/stretchr/testify/assert"
)

func TestAuthServiceParseToken(t *testing.T) {
	config := auth.JWTConfig{
		SecretKey:     []byte("test_secret"),
		SigningMethod: jwt.SigningMethodHS256,
	}
	service, err := auth.NewAuthService(config)
	assert.NoError(t, err, "Failed to create AuthService")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "test_issuer",
		Audience:  []string{"test_audience"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		Subject:   "123",
	})
	tokenStr, _ := token.SignedString(config.SecretKey)

	invalidToken := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.RegisteredClaims{
		Issuer:    "test_issuer",
		Audience:  []string{"test_audience"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		Subject:   "123",
	})
	invalidTokenStr, _ := invalidToken.SignedString(config.SecretKey)

	testCases := []struct {
		name       string
		tokenStr   string
		wantErr    bool
		errMessage string
	}{
		{
			name:     "Valid token",
			tokenStr: tokenStr,
			wantErr:  false,
		},
		{
			name:       "Invalid signature method",
			tokenStr:   invalidTokenStr,
			wantErr:    true,
			errMessage: "unexpected signing method",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := service.ParseToken(tc.tokenStr)
			if tc.wantErr {
				if err != nil {
					if tc.errMessage != "" {
						assert.Contains(t, err.Error(), tc.errMessage, "Error message should match")
					}
				} else {
					t.Error("Expected an error but got nil")
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims, "Claims should not be nil for valid tokens")
			}
		})
	}
}
