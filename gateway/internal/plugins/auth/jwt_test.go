package pluginJWT_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	pluginJWT "github.com/lucdoe/open-gateway/gateway/internal/plugins/auth"
	"github.com/stretchr/testify/assert"
)

func TestJWTServiceValidate(t *testing.T) {
	config := pluginJWT.JWTConfig{
		SecretKey:     []byte("test_secret"),
		SigningMethod: jwt.SigningMethodHS256,
		Issuer:        "test_issuer",
		Audience:      "test_audience",
		Scope:         "test_scope",
	}
	mockJWTService := pluginJWT.NewJWTService(config)

	validExpectedClaims := &jwt.RegisteredClaims{
		Issuer:    "test_issuer",
		Audience:  jwt.ClaimStrings{"test_audience"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * time.Minute)),
		Subject:   "123",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, validExpectedClaims)
	tokenStr, err := token.SignedString(config.SecretKey)
	assert.NoError(t, err)

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
			tokenStr:   "eyJhbGciOiJQUzUxMiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.J5W09-rNx0pt5_HBiydR-vOluS6oD-RpYNa8PVWwMcBDQSXiw6-EPW8iSsalXPspGj3ouQjAnOP_4-zrlUUlvUIt2T79XyNeiKuooyIFvka3Y5NnGiOUBHWvWcWp4RcQFMBrZkHtJM23sB5D7Wxjx0-HFeNk-Y3UJgeJVhg5NaWXypLkC4y0ADrUBfGAxhvGdRdULZivfvzuVtv6AzW6NRuEE6DM9xpoWX_4here-yvLS2YPiBTZ8xbB3axdM99LhES-n52lVkiX5AWg2JJkEROZzLMpaacA_xlbUz_zbIaOaoqk8gB5oO7kI6sZej3QAdGigQy-hXiRnW_L98d4GQ",
			wantErr:    true,
			errMessage: "unexpected signing method",
		},
		{
			name:       "Expired token",
			tokenStr:   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpc3MiOiJ0ZXN0X2lzc3VlciIsInN1YiI6IjEyMyIsImF1ZCI6WyJ0ZXN0X2F1ZGllbmNlIl0sImV4cCI6MTcxNTI1NTUyNH0.IruMf2V-cWArbQsIduGg_hKfgCXVPmAvAfUr2EKyrjc",
			wantErr:    true,
			errMessage: "token has invalid claims: token is expired",
		},
		{
			name:       "Invalid signing method",
			tokenStr:   "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.e30.HMhSiB95MEDIHjjV0a4DZ_qhHafRwXQ8D_lrcRrZO0eVHd7I9X3VZUz3oZ_Hoge2HaHQ3Yp6qjJODH3X8LFrIw",
			wantErr:    true,
			errMessage: "signature is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			claims, err := mockJWTService.Validate(tc.tokenStr)
			if tc.wantErr {
				assert.Error(t, err)
				if tc.errMessage != "" {
					assert.Contains(t, err.Error(), tc.errMessage)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, validExpectedClaims.Issuer, claims.(*jwt.RegisteredClaims).Issuer)
				assert.Equal(t, validExpectedClaims.Audience, claims.(*jwt.RegisteredClaims).Audience)
				assert.Equal(t, validExpectedClaims.ExpiresAt, claims.(*jwt.RegisteredClaims).ExpiresAt)
				assert.Equal(t, validExpectedClaims.Subject, claims.(*jwt.RegisteredClaims).Subject)
			}
		})
	}
}
