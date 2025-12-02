package golazy

import (
	"context"
	"sync"
	"time"
)

// withLoader is the internal implementation of Lazy[T] that supports a loader
// function, optional TTL-based caching, and preloaded values.
type withLoader[T any] struct {
	// value holds the cached result of the loader
	value T
	// args are constructor arguments forwarded to the loader on each invocation
	args []any
	// loader is the LazyFunc that loads the value
	loader LazyFunc[T]
	// loaded tracks whether value is valid (loader succeeded)
	loaded bool
	// withTTL enables TTL-based cache expiration
	withTTL bool
	// ttl is the time-to-live duration for cached values
	ttl time.Duration
	// lastLoad records the timestamp of the most recent successful load
	lastLoad *time.Time
	// mu serializes access to the cache and loader invocation
	mu *sync.Mutex
}

// newWithLoader constructs a withLoader that will call loader on demand. If
// withTTL is true, returned values are cached for the provided ttl duration.
// The args are forwarded to the loader on each invocation.
func newWithLoader[T any](loader LazyFunc[T], withTTL bool, ttl time.Duration, args ...any) *withLoader[T] {
	return &withLoader[T]{
		args:    args,
		loader:  loader,
		withTTL: withTTL,
		ttl:     ttl,
		mu:      &sync.Mutex{},
	}
}

// newWithLoaderPreloaded constructs a withLoader pre-populated with a value
// for the provided context. This is useful when you already have a value and want
// to expose it through the Lazy API with TTL support.
func newWithLoaderPreloaded[T any](value T, loader LazyFunc[T], withTTL bool, ttl time.Duration, args ...any) *withLoader[T] {
	lastLoad := time.Now()

	return &withLoader[T]{
		value:    value,
		args:     args,
		loader:   loader,
		withTTL:  withTTL,
		ttl:      ttl,
		lastLoad: &lastLoad,
		mu:       &sync.Mutex{},
	}
}

// Value returns the cached value if present and not expired. Otherwise
// it calls the loader to obtain the value, caches it, and returns it.
// The method is safe for concurrent use because it synchronizes using a mutex.
// If one context.Context is provided, it is passed to the loader; otherwise
// context.Background() is used.
func (l *withLoader[T]) Value(ctxs ...context.Context) (T, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	var err error
	needsRefresh := l.withTTL && l.lastLoad != nil && time.Since(*l.lastLoad) > l.ttl

	if !l.loaded || needsRefresh {
		var ctx context.Context

		if len(ctxs) > 0 {
			ctx = ctxs[0]
		} else {
			ctx = context.Background()
		}

		l.value, err = l.loader(ctx, l.args...)
		l.loaded = err == nil
		t := time.Now()
		l.lastLoad = &t
	}

	return l.value, err
}

// Clear marks the cached value as unloaded so the next Value() call will
// invoke the loader again.
func (l *withLoader[T]) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.loaded = false
	l.lastLoad = nil
}
