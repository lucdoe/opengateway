// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// jwksRefresh is how often the key set is fetched again, so keys an
	// identity provider retires stop being accepted.
	jwksRefresh = time.Hour
	// jwksRetry is how often a fetch is tried again while no keys are known.
	jwksRetry = 30 * time.Second
	// jwksUnknownKeyRefetch is the shortest time between fetches caused by a
	// token whose kid is not known yet, a sign of a newly rotated key. It
	// keeps tokens with made-up kids from hammering the identity provider.
	jwksUnknownKeyRefetch = time.Minute
	jwksTimeout           = 5 * time.Second
	jwksMaxBytes          = 1 << 20
)

// JWKS holds the public keys an identity provider publishes as a JSON Web
// Key Set (RFC 7517), such as https://example.com/.well-known/jwks.json. It
// fetches them when created, then every hour, and at once when a token names
// a key it doesn't know, at most once a minute. Until a fetch succeeds,
// every token is rejected.
type JWKS struct {
	url    string
	client *http.Client

	unknownKeyRefetch time.Duration

	mu        sync.RWMutex
	keys      []jwk
	fetchMu   sync.Mutex // one fetch at a time
	fetchedAt time.Time

	stop chan struct{}
	done chan struct{}
}

// jwk is one usable key from the set.
type jwk struct {
	id  string
	alg string // empty: any algorithm its type allows
	key crypto.PublicKey
}

// NewJWKS fetches the key set at url and keeps it fresh until Close. A
// failed first fetch is logged, not returned, so the gateway starts while
// the identity provider is down; it is tried again every 30 seconds.
func NewJWKS(url string) *JWKS {
	j := newJWKS(url, &http.Client{Timeout: jwksTimeout})
	if err := j.fetch(); err != nil {
		log.Printf("auth: fetching keys from %s failed, tokens are rejected until it works: %v", url, err)
	}
	go j.refreshLoop(jwksRefresh, jwksRetry)
	return j
}

func newJWKS(url string, client *http.Client) *JWKS {
	return &JWKS{
		url:               url,
		client:            client,
		unknownKeyRefetch: jwksUnknownKeyRefetch,
		stop:              make(chan struct{}),
		done:              make(chan struct{}),
	}
}

// Close stops the refreshes.
func (j *JWKS) Close() error {
	close(j.stop)
	<-j.done
	return nil
}

func (j *JWKS) refreshLoop(every, retry time.Duration) {
	defer close(j.done)
	for {
		wait := every
		if !j.hasKeys() {
			wait = retry
		}
		select {
		case <-j.stop:
			return
		case <-time.After(wait):
		}
		if err := j.fetch(); err != nil {
			log.Printf("auth: refreshing keys from %s failed, keeping the known ones: %v", j.url, err)
		}
	}
}

func (j *JWKS) hasKeys() bool {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return len(j.keys) > 0
}

// Keyfunc returns the keys that may have signed token: those with its kid,
// or every key when it has none, that suit its algorithm.
func (j *JWKS) Keyfunc(token *jwt.Token) (any, error) {
	kid, _ := token.Header["kid"].(string)
	alg := token.Method.Alg()

	keys := j.matching(kid, alg)
	if len(keys) == 0 && kid != "" && j.refetchForUnknownKey() {
		keys = j.matching(kid, alg)
	}
	switch len(keys) {
	case 0:
		return nil, fmt.Errorf("no key with kid %q for %s", kid, alg)
	case 1:
		return keys[0], nil
	}
	set := jwt.VerificationKeySet{}
	for _, key := range keys {
		set.Keys = append(set.Keys, key)
	}
	return set, nil
}

func (j *JWKS) matching(kid, alg string) []crypto.PublicKey {
	j.mu.RLock()
	defer j.mu.RUnlock()
	var keys []crypto.PublicKey
	for _, k := range j.keys {
		if (kid == "" || k.id == kid) && (k.alg == "" || k.alg == alg) && suits(k.key, alg) {
			keys = append(keys, k.key)
		}
	}
	return keys
}

// refetchForUnknownKey fetches the set again unless that happened less than
// unknownKeyRefetch ago, and reports whether it did.
// Requests that wait for a fetch in progress then find it recent and use its
// keys.
func (j *JWKS) refetchForUnknownKey() bool {
	j.fetchMu.Lock()
	defer j.fetchMu.Unlock()
	if time.Since(j.fetchedAt) < j.unknownKeyRefetch {
		return false
	}
	if err := j.fetchLocked(); err != nil {
		log.Printf("auth: fetching keys from %s for an unknown kid failed: %v", j.url, err)
		return false
	}
	return true
}

// suits reports whether key can verify a signature made with alg. It keeps
// a token from choosing how a key is used, such as an RSA key as an HMAC
// secret.
func suits(key crypto.PublicKey, alg string) bool {
	switch k := key.(type) {
	case *rsa.PublicKey:
		return strings.HasPrefix(alg, "RS") || strings.HasPrefix(alg, "PS")
	case *ecdsa.PublicKey:
		return alg == map[string]string{"P-256": "ES256", "P-384": "ES384", "P-521": "ES512"}[k.Curve.Params().Name]
	case ed25519.PublicKey:
		return alg == "EdDSA"
	}
	return false
}

// fetch replaces the keys with the set at url. On any error the keys stay
// as they were. A fetch, failed or not, counts for unknownKeyRefetch.
func (j *JWKS) fetch() error {
	j.fetchMu.Lock()
	defer j.fetchMu.Unlock()
	return j.fetchLocked()
}

func (j *JWKS) fetchLocked() error {
	j.fetchedAt = time.Now()

	res, err := j.client.Get(j.url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("status %s", res.Status)
	}
	body, err := io.ReadAll(io.LimitReader(res.Body, jwksMaxBytes+1))
	if err != nil {
		return err
	}
	if len(body) > jwksMaxBytes {
		return fmt.Errorf("key set is over %d bytes", jwksMaxBytes)
	}
	keys, err := parseKeySet(body)
	if err != nil {
		return err
	}

	j.mu.Lock()
	j.keys = keys
	j.mu.Unlock()
	return nil
}

// rawKey holds the JWK members the gateway reads.
type rawKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// parseKeySet returns the signature keys of a JWK set. Keys for encryption,
// of unknown types or that don't parse are skipped, so one odd key doesn't
// lock out the rest; a set without a single usable key is an error.
func parseKeySet(body []byte) ([]jwk, error) {
	var set struct {
		Keys []rawKey `json:"keys"`
	}
	if err := json.Unmarshal(body, &set); err != nil {
		return nil, fmt.Errorf("not a JWK set: %w", err)
	}
	var keys []jwk
	for _, raw := range set.Keys {
		if raw.Use != "" && raw.Use != "sig" {
			continue
		}
		key, err := raw.publicKey()
		if err != nil {
			continue
		}
		keys = append(keys, jwk{id: raw.Kid, alg: raw.Alg, key: key})
	}
	if len(keys) == 0 {
		return nil, errors.New("the key set has no usable signature key")
	}
	return keys, nil
}

func (k rawKey) publicKey() (crypto.PublicKey, error) {
	switch k.Kty {
	case "RSA":
		n, err := bigInt(k.N)
		if err != nil {
			return nil, err
		}
		e, err := bigInt(k.E)
		if err != nil || !e.IsInt64() || e.Int64() < 3 || e.Int64() > 1<<31 {
			return nil, errors.New("bad RSA exponent")
		}
		if n.BitLen() < 2048 {
			return nil, errors.New("RSA key shorter than 2048 bits")
		}
		return &rsa.PublicKey{N: n, E: int(e.Int64())}, nil
	case "EC":
		curve := map[string]elliptic.Curve{"P-256": elliptic.P256(), "P-384": elliptic.P384(), "P-521": elliptic.P521()}[k.Crv]
		if curve == nil {
			return nil, fmt.Errorf("unsupported curve %q", k.Crv)
		}
		x, errX := bigInt(k.X)
		y, errY := bigInt(k.Y)
		if errX != nil || errY != nil || !curve.IsOnCurve(x, y) {
			return nil, errors.New("bad EC point")
		}
		return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
	case "OKP":
		x, err := base64.RawURLEncoding.DecodeString(k.X)
		if k.Crv != "Ed25519" || err != nil || len(x) != ed25519.PublicKeySize {
			return nil, errors.New("bad Ed25519 key")
		}
		return ed25519.PublicKey(x), nil
	}
	return nil, fmt.Errorf("unsupported key type %q", k.Kty)
}

func bigInt(value string) (*big.Int, error) {
	b, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(b) == 0 {
		return nil, errors.New("bad base64url number")
	}
	return new(big.Int).SetBytes(b), nil
}
