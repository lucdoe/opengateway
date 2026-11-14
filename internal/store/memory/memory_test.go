// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

package memory

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/lucdoe/opengateway/internal/store"
	"github.com/lucdoe/opengateway/internal/store/storetest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// clock is a time source tests move forward by hand.
type clock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func newTestStore() (*Store, *clock) {
	c := &clock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	return newWithClock(c.Now), c
}

func TestStoreConformance(t *testing.T) {
	storetest.Run(t, func(t *testing.T) storetest.Backend {
		s, c := newTestStore()
		return storetest.Backend{Store: s, Advance: c.Advance}
	})
}

func TestNewStartsEmpty(t *testing.T) {
	_, err := New().Get("k")
	assert.ErrorIs(t, err, store.ErrNotFound)
}

func TestIncrementOverAValueStartsACounter(t *testing.T) {
	s, _ := newTestStore()
	require.NoError(t, s.Set("k", "cached response", time.Minute))

	count, _, err := s.Increment("k", time.Minute)

	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestSetDropsNewKeysWhenFull(t *testing.T) {
	s, _ := newTestStore()
	for i := 0; i < MaxEntries*2; i++ {
		require.NoError(t, s.Set("k"+strconv.Itoa(i), "v", time.Hour))
	}

	held := 0
	for i := range s.shards {
		held += len(s.shards[i].entries)
	}
	assert.LessOrEqual(t, held, MaxEntries)

	_, err := s.Get("k0")
	assert.NoError(t, err, "keys stored before it was full are kept")

	count, _, err := s.Increment("counter", time.Minute)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "counters are kept even when full")
}

// One client per IPv6 address could otherwise grow the store without limit.
func TestIncrementRefusesNewCountersWhenFull(t *testing.T) {
	s, _ := newTestStore()
	refused := 0
	for i := 0; i < MaxCounters*2; i++ {
		if _, _, err := s.Increment("c"+strconv.Itoa(i), time.Minute); err != nil {
			assert.ErrorIs(t, err, ErrFull)
			refused++
		}
	}

	held := 0
	for i := range s.shards {
		held += len(s.shards[i].entries)
	}
	assert.LessOrEqual(t, held, MaxCounters)
	assert.Positive(t, refused)

	count, _, err := s.Increment("c0", time.Minute)
	require.NoError(t, err, "existing counters keep counting")
	assert.Equal(t, int64(2), count)

	require.NoError(t, s.Set("value", "v", time.Hour))
	_, err = s.Get("value")
	assert.NoError(t, err, "counters do not use up the room for values")
}

func TestExpiredCountersFreeTheirRoom(t *testing.T) {
	s, clock := newTestStore()
	for i := 0; i < MaxCounters; i++ {
		s.Increment("c"+strconv.Itoa(i), time.Minute)
	}

	clock.Advance(2 * time.Minute)
	_, _, err := s.Increment("new", time.Minute)

	assert.NoError(t, err)
}

func TestSetReplacesAKeyEvenWhenFull(t *testing.T) {
	s, _ := newTestStore()
	for i := 0; i < MaxEntries*2; i++ {
		require.NoError(t, s.Set("k"+strconv.Itoa(i), "v", time.Hour))
	}

	require.NoError(t, s.Set("k0", "updated", time.Hour))

	value, err := s.Get("k0")
	require.NoError(t, err)
	assert.Equal(t, "updated", value)
}

func TestSweepRemovesExpiredKeys(t *testing.T) {
	s, c := newTestStore()
	for i := 0; i < 1000; i++ {
		require.NoError(t, s.Set("k"+strconv.Itoa(i), "v", time.Second))
	}

	c.Advance(sweepInterval)
	for i := range s.shards {
		// any write runs the shard's sweep once the interval has passed
		s.shards[i].mu.Lock()
		s.shards[i].sweep(c.Now())
		s.shards[i].mu.Unlock()
	}

	held := 0
	for i := range s.shards {
		held += len(s.shards[i].entries)
	}
	assert.Zero(t, held)
}

func TestSweepWaitsForTheInterval(t *testing.T) {
	s, c := newTestStore()
	require.NoError(t, s.Set("k", "v", time.Second))
	c.Advance(2 * time.Second)

	sh := s.shard("k")
	sh.mu.Lock()
	sh.sweep(c.Now())
	held := len(sh.entries)
	sh.mu.Unlock()

	assert.Equal(t, 1, held, "the expired key stays until the next sweep or read")
}
