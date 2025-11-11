package golazy

import (
	"time"
)

// Package golazy provides a small, generic lazy-loading helper.
// It exposes a Lazy[T] abstraction that can load values on demand using a
// user-provided loader function. The implementation supports optional TTL-based
// caching and preloaded values keyed by an arbitrary context value.

// LazyFunc is the type of loader functions passed to WithLoader/Preloaded.
// The loader receives a context (of any type) and should return a value of type
// T and an error. The context can be used as a key for per-context caching.
type LazyFunc[T any] func(ctx any) (T, error)

// Lazy represents a lazy-loaded value of type T. Calls to Value(ctx) will
// invoke the configured loader when needed and cache the result. Clear/ClearAll
// allow clearing cached entries.
type Lazy[T any] interface {
	// Value returns the value for the given context, calling the loader if the
	// value is not yet present (or if TTL expired for TTL-enabled instances).
	Value(ctx any) (T, error)

	// Clear deletes the cached value for the provided context.
	Clear(ctx any)

	// ClearAll clears the entire cache maintained by the Lazy instance.
	ClearAll()
}

// WithLoader creates a Lazy[T] that will call the provided loader when
// using a context for the first time at calling the Value() method.
// Next times using that context, value will be returned from cache
// unless cache is cleared.
func WithLoader[T any](loader LazyFunc[T]) Lazy[T] {
	return newWithLoader(loader, false, 0)
}

// WithLoaderTTL id like WithLoader but also enables TTL caching per
// context. Cached value will be invalidated after TTL expires, then
// provided loader will be invoked again to load value again.
func WithLoaderTTL[T any](loader LazyFunc[T], ttl time.Duration) Lazy[T] {
	return newWithLoader(loader, true, ttl)
}

// Preloaded returns a Lazy[T] pre-populated with value for the given ctx. The
// loader is still kept and may be used for other contexts or after cache is
// cleared for current/all contexts.
func Preloaded[T any](loader LazyFunc[T], value T, ctx any) Lazy[T] {
	return newWithLoaderPreloaded(loader, value, ctx, false, 0)
}

// PreloadedTTL is like Preloaded but also enables TTL caching for the instance.
// Cached value will be invalidated after TTL expires, then provided loader will
// be invoked again to load value again.
func PreloadedTTL[T any](loader LazyFunc[T], value T, ctx any, ttl time.Duration) *withLoader[T] {
	return newWithLoaderPreloaded(loader, value, ctx, true, ttl)
}

// Static returns a Lazy[T] that always returns the provided value
// (even after clearing cache) and never invokes a loader. This is a convenience
// for tests or fixed values.
func Static[T any](value T) *static[T] {
	return &static[T]{
		value: value,
	}
}
