// Copyright lucdoe
// SPDX-License-Identifier: Apache-2.0

// Package memory is a store that keeps everything in the gateway process.
// It needs no other service, but its state is per gateway instance: several
// instances each count their own rate limits and keep their own cache.
package memory

import (
	"errors"
	"hash/maphash"
	"strconv"
	"sync"
	"time"

	"github.com/lucdoe/opengateway/internal/store"
)

const (
	shardCount = 32
	// MaxEntries bounds how many values the store holds, so keys made from
	// request paths and queries cannot grow it without limit. Once it is
	// full, Set drops new keys.
	MaxEntries = 100_000
	// MaxCounters bounds the counters separately, so a full cache never
	// stops rate limiting and many clients cannot grow the store either.
	// Once it is full, Increment refuses new counters with ErrFull.
	MaxCounters = 100_000
	// sweepInterval is how often a shard removes its expired keys.
	sweepInterval = time.Minute
)

type entry struct {
	value   string
	count   int64
	expires time.Time
}

// ErrFull is returned by Increment for a new counter once MaxCounters
// counters exist.
var ErrFull = errors.New("memory store: too many counters")

type shard struct {
	mu        sync.Mutex
	entries   map[string]entry
	values    int // entries set with Set
	counters  int // entries made by Increment
	lastSweep time.Time
}

// put stores e at key and keeps the values and counters counts right.
func (sh *shard) put(key string, e entry) {
	sh.remove(key)
	sh.entries[key] = e
	sh.tally(e, 1)
}

func (sh *shard) remove(key string) {
	if old, ok := sh.entries[key]; ok {
		sh.tally(old, -1)
		delete(sh.entries, key)
	}
}

func (sh *shard) tally(e entry, delta int) {
	if e.count > 0 {
		sh.counters += delta
	} else {
		sh.values += delta
	}
}

// Store is an in-process store.Store. Expired keys are dropped on read and by
// a per-shard sweep.
type Store struct {
	seed   maphash.Seed
	shards [shardCount]shard
	now    func() time.Time
}

var _ store.Store = (*Store)(nil)

// New returns an empty store.
func New() *Store {
	return newWithClock(time.Now)
}

func newWithClock(now func() time.Time) *Store {
	s := &Store{seed: maphash.MakeSeed(), now: now}
	for i := range s.shards {
		s.shards[i].entries = map[string]entry{}
		s.shards[i].lastSweep = now()
	}
	return s
}

func (s *Store) shard(key string) *shard {
	return &s.shards[maphash.String(s.seed, key)%shardCount]
}

// Get returns the value at key, or store.ErrNotFound.
func (s *Store) Get(key string) (string, error) {
	sh := s.shard(key)
	now := s.now()
	sh.mu.Lock()
	defer sh.mu.Unlock()
	e, ok := sh.entries[key]
	if !ok || !now.Before(e.expires) {
		sh.remove(key)
		return "", store.ErrNotFound
	}
	return e.value, nil
}

// Set stores value at key for ttl. When the store is full it keeps the
// existing keys and drops this one; a cache then simply misses again.
func (s *Store) Set(key, value string, ttl time.Duration) error {
	sh := s.shard(key)
	now := s.now()
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.sweep(now)
	if old, exists := sh.entries[key]; (!exists || old.count > 0) && sh.values >= MaxEntries/shardCount {
		return nil
	}
	sh.put(key, entry{value: value, expires: now.Add(ttl)})
	return nil
}

// Increment adds one to the counter at key. A counter that does not exist,
// has expired, or holds a value set with Set starts over at 1. A new counter
// beyond MaxCounters is refused with ErrFull.
func (s *Store) Increment(key string, window time.Duration) (int64, time.Duration, error) {
	sh := s.shard(key)
	now := s.now()
	sh.mu.Lock()
	defer sh.mu.Unlock()
	sh.sweep(now)
	e, ok := sh.entries[key]
	if !ok || !now.Before(e.expires) || e.count == 0 {
		if (!ok || e.count == 0) && sh.counters >= MaxCounters/shardCount {
			return 0, 0, ErrFull
		}
		e = entry{expires: now.Add(window)}
	}
	e.count++
	e.value = strconv.FormatInt(e.count, 10)
	sh.put(key, e)
	return e.count, e.expires.Sub(now), nil
}

// sweep removes expired keys if the last sweep was a while ago. The caller
// holds sh.mu.
func (sh *shard) sweep(now time.Time) {
	if now.Sub(sh.lastSweep) < sweepInterval {
		return
	}
	sh.lastSweep = now
	for key, e := range sh.entries {
		if !now.Before(e.expires) {
			sh.remove(key)
		}
	}
}
