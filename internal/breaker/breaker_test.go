// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package breaker_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/lucdoe/opengateway/internal/breaker"
	"github.com/lucdoe/opengateway/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// upstream answers with whatever status is stored in status and counts the
// requests that reach it.
type upstream struct {
	status atomic.Int32
	calls  atomic.Int32
}

func newUpstream(status int) *upstream {
	u := &upstream{}
	u.status.Store(int32(status))
	return u
}

func (u *upstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u.calls.Add(1)
	w.WriteHeader(int(u.status.Load()))
	w.Write([]byte("from upstream"))
}

func settings(failures int, openTimeout time.Duration) config.CircuitBreakerConfig {
	return config.CircuitBreakerConfig{ConsecutiveFailures: failures, OpenTimeout: openTimeout, HalfOpenRequests: 1}
}

func send(handler http.Handler) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "http://gateway.test/items", nil))
	return w
}

func TestWrapPassesResponsesThroughWhileClosed(t *testing.T) {
	up := newUpstream(http.StatusCreated)
	handler := breaker.Wrap("api", settings(3, time.Minute), up)

	w := send(handler)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "from upstream", w.Body.String())
}

func TestWrapOpensAfterConsecutiveFailures(t *testing.T) {
	for _, status := range []int{500, 502, 503, 504} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			up := newUpstream(status)
			handler := breaker.Wrap("api", settings(3, 90*time.Second), up)

			for i := 0; i < 3; i++ {
				assert.Equal(t, status, send(handler).Code, "request %d still reaches the upstream", i+1)
			}
			w := send(handler)

			assert.Equal(t, http.StatusServiceUnavailable, w.Code)
			assert.Equal(t, "90", w.Header().Get("Retry-After"))
			assert.JSONEq(t, `{"message": "Service unavailable"}`, w.Body.String())
			assert.Equal(t, int32(3), up.calls.Load(), "the open breaker does not call the upstream")
		})
	}
}

func TestWrapDoesNotCountOtherStatusesAsFailures(t *testing.T) {
	for _, status := range []int{200, 404, 429, 501} {
		t.Run(http.StatusText(status), func(t *testing.T) {
			up := newUpstream(status)
			handler := breaker.Wrap("api", settings(2, time.Minute), up)

			for i := 0; i < 5; i++ {
				assert.Equal(t, status, send(handler).Code)
			}
		})
	}
}

func TestWrapCountsAnImplicitOKAsSuccess(t *testing.T) {
	failing := true
	handler := breaker.Wrap("api", settings(2, time.Minute), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if failing {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		w.Write([]byte("ok")) // no WriteHeader: the status is 200
	}))

	send(handler)
	failing = false
	send(handler)
	failing = true
	send(handler)

	assert.Equal(t, http.StatusBadGateway, send(handler).Code, "the success in between reset the count, so the breaker is still closed")
}

func TestWrapRecordsTheFinalStatusAfterInformationalResponses(t *testing.T) {
	handler := breaker.Wrap("api", settings(1, time.Minute), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusEarlyHints)
		w.WriteHeader(http.StatusBadGateway)
	}))

	send(handler)

	assert.Equal(t, http.StatusServiceUnavailable, send(handler).Code, "the 502 after the 103 counted as a failure")
}

func TestWrapIgnoresRequestsWhoseClientWentAway(t *testing.T) {
	up := newUpstream(http.StatusBadGateway)
	handler := breaker.Wrap("api", settings(1, time.Minute), up)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "http://gateway.test/items", nil).WithContext(ctx))

	assert.Equal(t, http.StatusBadGateway, send(handler).Code, "the cancelled request did not open the breaker")
}

func TestWrapClosesAgainWhenTheTrialRequestSucceeds(t *testing.T) {
	up := newUpstream(http.StatusBadGateway)
	handler := breaker.Wrap("api", settings(1, 50*time.Millisecond), up)

	send(handler)
	require.Equal(t, http.StatusServiceUnavailable, send(handler).Code)

	up.status.Store(int32(http.StatusOK))
	time.Sleep(60 * time.Millisecond)

	assert.Equal(t, http.StatusOK, send(handler).Code, "the trial request reaches the upstream")
	assert.Equal(t, http.StatusOK, send(handler).Code, "the breaker is closed again")
}

func TestWrapOpensAgainWhenTheTrialRequestFails(t *testing.T) {
	up := newUpstream(http.StatusBadGateway)
	handler := breaker.Wrap("api", settings(1, 50*time.Millisecond), up)

	send(handler)
	time.Sleep(60 * time.Millisecond)

	assert.Equal(t, http.StatusBadGateway, send(handler).Code, "the trial request reaches the upstream")
	assert.Equal(t, http.StatusServiceUnavailable, send(handler).Code, "the failed trial opened the breaker again")
	assert.Equal(t, int32(2), up.calls.Load())
}

func TestWrapLimitsTrialRequestsWhileHalfOpen(t *testing.T) {
	release := make(chan struct{})
	started := make(chan struct{})
	var calls atomic.Int32
	handler := breaker.Wrap("api", settings(1, 50*time.Millisecond), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		close(started)
		<-release
	}))

	send(handler)
	time.Sleep(60 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		send(handler)
		close(done)
	}()
	<-started

	assert.Equal(t, http.StatusServiceUnavailable, send(handler).Code, "only one trial request runs at a time")
	close(release)
	<-done
}

func TestWrapKeepsTheResponseFlushable(t *testing.T) {
	handler := breaker.Wrap("api", settings(1, time.Minute), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("part"))
		assert.NoError(t, http.NewResponseController(w).Flush())
	}))

	w := send(handler)

	assert.True(t, w.Flushed)
}

func TestWrapTellsObserversEveryNewState(t *testing.T) {
	up := newUpstream(http.StatusBadGateway)
	var states []int64
	handler := breaker.Wrap("api", settings(1, time.Millisecond), up, func(state int64) { states = append(states, state) })

	send(handler) // fails: open
	time.Sleep(5 * time.Millisecond)
	up.status.Store(int32(http.StatusOK))
	send(handler) // the trial request: half-open, then closed

	assert.Equal(t, []int64{2, 1, 0}, states, "open, half-open, closed")
}
