// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// asymmetric are the algorithms the gateway accepts with a JWKS.
var asymmetric = []string{"RS256", "RS384", "RS512", "PS256", "PS384", "PS512", "ES256", "ES384", "ES512", "EdDSA"}

// signer is a private key with the kid and JWK under which its public key
// is published.
type signer struct {
	kid string
	key crypto.Signer
	jwk map[string]string
}

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func rsaSigner(t *testing.T, kid string) signer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	return signer{kid, key, map[string]string{"kty": "RSA", "kid": kid, "use": "sig",
		"n": b64(key.N.Bytes()), "e": b64(big.NewInt(int64(key.E)).Bytes())}}
}

func ecSigner(t *testing.T, kid string) signer {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	return signer{kid, key, map[string]string{"kty": "EC", "kid": kid, "crv": "P-256",
		"x": b64(key.X.FillBytes(make([]byte, 32))), "y": b64(key.Y.FillBytes(make([]byte, 32)))}}
}

func edSigner(t *testing.T, kid string) signer {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	return signer{kid, private, map[string]string{"kty": "OKP", "kid": kid, "crv": "Ed25519", "x": b64(public)}}
}

// token signs claims valid for a minute with s, under its kid unless kid is
// false.
func (s signer) token(t *testing.T, method jwt.SigningMethod, withKid bool) string {
	t.Helper()
	token := jwt.NewWithClaims(method, Claims{RegisteredClaims: jwt.RegisteredClaims{
		Subject: "alice", ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}})
	if withKid {
		token.Header["kid"] = s.kid
	}
	signed, err := token.SignedString(s.key)
	require.NoError(t, err)
	return signed
}

// provider is an identity provider serving a key set that tests can swap.
type provider struct {
	*httptest.Server
	mu      sync.Mutex
	body    string
	status  int
	fetches atomic.Int32
}

func newProvider(t *testing.T, keys ...map[string]string) *provider {
	t.Helper()
	p := &provider{status: http.StatusOK}
	p.publish(t, keys...)
	p.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.fetches.Add(1)
		p.mu.Lock()
		defer p.mu.Unlock()
		w.WriteHeader(p.status)
		w.Write([]byte(p.body))
	}))
	t.Cleanup(p.Close)
	return p
}

func (p *provider) publish(t *testing.T, keys ...map[string]string) {
	t.Helper()
	if keys == nil {
		keys = []map[string]string{}
	}
	body, err := json.Marshal(map[string]any{"keys": keys})
	require.NoError(t, err)
	p.mu.Lock()
	p.body = string(body)
	p.mu.Unlock()
}

// jwksAuth is an Auth on p's keys that accepts every asymmetric algorithm.
func jwksAuth(t *testing.T, p *provider) (*Auth, *JWKS) {
	t.Helper()
	jwks := NewJWKS(p.URL)
	t.Cleanup(func() { jwks.Close() })
	a, err := NewAuthService(JWTConfig{Keys: jwks.Keyfunc, Algorithms: asymmetric})
	require.NoError(t, err)
	return a, jwks
}

func TestJWKSAcceptsTokensFromEveryKeyType(t *testing.T) {
	rsaKey, ecKey, edKey := rsaSigner(t, "rsa"), ecSigner(t, "ec"), edSigner(t, "ed")
	a, _ := jwksAuth(t, newProvider(t, rsaKey.jwk, ecKey.jwk, edKey.jwk))

	for name, token := range map[string]string{
		"RS256": rsaKey.token(t, jwt.SigningMethodRS256, true),
		"PS256": rsaKey.token(t, jwt.SigningMethodPS256, true),
		"ES256": ecKey.token(t, jwt.SigningMethodES256, true),
		"EdDSA": edKey.token(t, jwt.SigningMethodEdDSA, true),
		"no kid, the right key is found by trying": ecKey.token(t, jwt.SigningMethodES256, false),
	} {
		t.Run(name, func(t *testing.T) {
			claims, err := a.Validate(token)
			require.NoError(t, err)
			assert.Equal(t, "alice", claims.Subject)
		})
	}
}

func TestJWKSRejectsTokensNoPublishedKeySigned(t *testing.T) {
	published := rsaSigner(t, "a")
	a, _ := jwksAuth(t, newProvider(t, published.jwk))

	other := rsaSigner(t, "a") // same kid, different key
	_, err := a.Validate(other.token(t, jwt.SigningMethodRS256, true))
	assert.Error(t, err)
}

// The classic JWKS attack: sign an HS256 token with the public key as the
// secret, so a verifier that lets the token pick the algorithm accepts it.
func TestJWKSRejectsAPublicKeyUsedAsAnHMACSecret(t *testing.T) {
	published := rsaSigner(t, "a")
	a, jwks := jwksAuth(t, newProvider(t, published.jwk))

	public, err := x509.MarshalPKIXPublicKey(published.key.Public())
	require.NoError(t, err)
	forged := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{RegisteredClaims: jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
	}})
	forged.Header["kid"] = "a"
	signed, err := forged.SignedString(public)
	require.NoError(t, err)

	_, err = a.Validate(signed)
	assert.Error(t, err)

	// Even with HS256 allowed, no JWKS key serves as an HMAC secret.
	lenient, err := NewAuthService(JWTConfig{Keys: jwks.Keyfunc, Algorithms: append([]string{"HS256"}, asymmetric...)})
	require.NoError(t, err)
	_, err = lenient.Validate(signed)
	assert.Error(t, err)
}

func TestJWKSHonoursTheKeysAlgorithm(t *testing.T) {
	key := rsaSigner(t, "a")
	key.jwk["alg"] = "RS256"
	a, _ := jwksAuth(t, newProvider(t, key.jwk))

	_, err := a.Validate(key.token(t, jwt.SigningMethodRS256, true))
	assert.NoError(t, err)
	_, err = a.Validate(key.token(t, jwt.SigningMethodPS256, true))
	assert.Error(t, err, "the key says RS256 only")
}

func TestJWKSFetchesAgainForANewlyRotatedKey(t *testing.T) {
	old, rotated := ecSigner(t, "2025"), ecSigner(t, "2026")
	p := newProvider(t, old.jwk)
	a, jwks := jwksAuth(t, p)
	jwks.unknownKeyRefetch = 0
	require.Equal(t, int32(1), p.fetches.Load())

	p.publish(t, old.jwk, rotated.jwk)
	_, err := a.Validate(rotated.token(t, jwt.SigningMethodES256, true))

	assert.NoError(t, err, "the unknown kid made it fetch the new set")
	assert.Equal(t, int32(2), p.fetches.Load())
}

func TestJWKSLimitsFetchesForUnknownKeys(t *testing.T) {
	known, stranger := ecSigner(t, "known"), ecSigner(t, "made-up")
	p := newProvider(t, known.jwk)
	a, _ := jwksAuth(t, p)

	for i := 0; i < 5; i++ {
		_, err := a.Validate(stranger.token(t, jwt.SigningMethodES256, true))
		assert.Error(t, err)
	}

	assert.Equal(t, int32(1), p.fetches.Load(), "within a minute of the last fetch, unknown kids fetch nothing")
}

func TestJWKSKeepsItsKeysWhenAFetchFails(t *testing.T) {
	key := edSigner(t, "a")
	p := newProvider(t, key.jwk)
	a, jwks := jwksAuth(t, p)

	p.mu.Lock()
	p.status = http.StatusInternalServerError
	p.mu.Unlock()
	assert.Error(t, jwks.fetch())

	_, err := a.Validate(key.token(t, jwt.SigningMethodEdDSA, true))
	assert.NoError(t, err)
}

func TestJWKSStartsWithoutTheProviderAndPicksUpKeysLater(t *testing.T) {
	key := edSigner(t, "a")
	p := newProvider(t)
	p.status = http.StatusServiceUnavailable

	jwks := newJWKS(p.URL, p.Client())
	go jwks.refreshLoop(time.Hour, 10*time.Millisecond)
	t.Cleanup(func() { jwks.Close() })
	a, err := NewAuthService(JWTConfig{Keys: jwks.Keyfunc, Algorithms: asymmetric})
	require.NoError(t, err)

	_, err = a.Validate(key.token(t, jwt.SigningMethodEdDSA, false))
	assert.Error(t, err, "no keys yet")

	p.mu.Lock()
	p.status = http.StatusOK
	p.mu.Unlock()
	p.publish(t, key.jwk)

	assert.Eventually(t, func() bool {
		_, err := a.Validate(key.token(t, jwt.SigningMethodEdDSA, false))
		return err == nil
	}, 2*time.Second, 10*time.Millisecond)
}

func TestParseKeySetSkipsKeysItCannotUse(t *testing.T) {
	good := ecSigner(t, "good")
	short, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	body, err := json.Marshal(map[string]any{"keys": []map[string]string{
		good.jwk,
		{"kty": "RSA", "kid": "enc", "use": "enc", "n": good.jwk["x"], "e": "AQAB"},
		{"kty": "RSA", "kid": "short", "n": b64(short.N.Bytes()), "e": "AQAB"},
		{"kty": "EC", "kid": "off-curve", "crv": "P-256", "x": good.jwk["x"], "y": good.jwk["x"]},
		{"kty": "EC", "kid": "curve", "crv": "secp256k1", "x": good.jwk["x"], "y": good.jwk["y"]},
		{"kty": "OKP", "kid": "x25519", "crv": "X25519", "x": b64(make([]byte, 32))},
		{"kty": "oct", "kid": "secret", "k": "c2VjcmV0"},
	}})
	require.NoError(t, err)

	keys, err := parseKeySet(body)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, "good", keys[0].id)
}

func TestParseKeySetRejectsSetsWithoutAUsableKey(t *testing.T) {
	for name, body := range map[string]string{
		"not JSON":     "<html>",
		"empty set":    `{"keys": []}`,
		"only secrets": `{"keys": [{"kty": "oct", "k": "c2VjcmV0"}]}`,
	} {
		t.Run(name, func(t *testing.T) {
			_, err := parseKeySet([]byte(body))
			assert.Error(t, err)
		})
	}
}

func TestJWKSRejectsAnOversizedSet(t *testing.T) {
	p := newProvider(t)
	p.body = `{"keys": [` + strings.Repeat(" ", jwksMaxBytes) + `]}`
	jwks := newJWKS(p.URL, p.Client())

	assert.ErrorContains(t, jwks.fetch(), "over")
}

func TestJWKSTriesEveryFittingKeyForATokenWithoutKid(t *testing.T) {
	first, second := ecSigner(t, "first"), ecSigner(t, "second")
	a, _ := jwksAuth(t, newProvider(t, first.jwk, second.jwk))

	_, err := a.Validate(second.token(t, jwt.SigningMethodES256, false))

	assert.NoError(t, err, "both keys suit ES256, so both are tried")
}

func TestNewJWKSStartsWhenTheProviderCannotBeReached(t *testing.T) {
	p := newProvider(t)
	p.Close()

	jwks := NewJWKS(p.URL)
	defer jwks.Close()

	assert.False(t, jwks.hasKeys())
}

func TestJWKSRefreshLoopKeepsTryingAfterFailures(t *testing.T) {
	p := newProvider(t)
	p.status = http.StatusServiceUnavailable
	jwks := newJWKS(p.URL, p.Client())
	go jwks.refreshLoop(time.Hour, time.Millisecond)

	assert.Eventually(t, func() bool { return p.fetches.Load() >= 3 }, 2*time.Second, time.Millisecond)
	jwks.Close()
	assert.False(t, jwks.hasKeys())
}

func TestJWKSRejectsAnUnknownKidWhenTheRefetchFails(t *testing.T) {
	known, rotated := ecSigner(t, "known"), ecSigner(t, "rotated")
	p := newProvider(t, known.jwk)
	a, jwks := jwksAuth(t, p)
	jwks.unknownKeyRefetch = 0

	p.mu.Lock()
	p.status = http.StatusBadGateway
	p.mu.Unlock()
	_, err := a.Validate(rotated.token(t, jwt.SigningMethodES256, true))

	assert.Error(t, err)
	assert.Equal(t, int32(2), p.fetches.Load(), "it did try")
}

func TestJWKSFetchErrors(t *testing.T) {
	truncated := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "100")
		w.Write([]byte(`{"keys": [`))
	}))
	t.Cleanup(truncated.Close)
	noKeys := newProvider(t)

	for name, tc := range map[string]struct {
		url, want string
	}{
		"body cut off":  {truncated.URL, "unexpected EOF"},
		"no usable key": {noKeys.URL, "no usable signature key"},
		"not a URL":     {"://", "missing protocol scheme"},
	} {
		t.Run(name, func(t *testing.T) {
			assert.ErrorContains(t, newJWKS(tc.url, http.DefaultClient).fetch(), tc.want)
		})
	}
}

func TestParseKeySetSkipsMalformedRSAKeys(t *testing.T) {
	good := rsaSigner(t, "good")
	for name, key := range map[string]map[string]string{
		"modulus not base64url": {"kty": "RSA", "n": "!!", "e": "AQAB"},
		"empty modulus":         {"kty": "RSA", "n": "", "e": "AQAB"},
		"exponent not base64":   {"kty": "RSA", "n": good.jwk["n"], "e": "!!"},
		"exponent of 1":         {"kty": "RSA", "n": good.jwk["n"], "e": "AQ"},
		"exponent over 2^31":    {"kty": "RSA", "n": good.jwk["n"], "e": b64(new(big.Int).Lsh(big.NewInt(1), 40).Bytes())},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := rawKey{Kty: key["kty"], N: key["n"], E: key["e"]}.publicKey()
			assert.Error(t, err)
		})
	}
}

func TestSuitsRejectsKeysOfOtherTypes(t *testing.T) {
	assert.False(t, suits([]byte("secret"), "HS256"))
	assert.False(t, suits(rsaSigner(t, "a").key.Public(), "ES256"))
}
