// Package golazy is a small, dependency-free Go library that provides a
// context-aware, generic lazy-loading abstraction. It lets you declare values
// that are loaded on first use (or preloaded) and optionally cached with a TTL.
// The implementation is safe for concurrent use and keeps the public API
// intentionally small.
package golazy

import (
	"context"
	"time"
)

// LazyFunc is the type of loader functions passed to WithLoader, WithLoaderTTL,
// Preloaded, and PreloadedTTL. The loader receives a context.Context (for
// cancellation/deadlines/values) and optional args provided during construction,
// and must return a value of type T and an error.
type LazyFunc[T any] func(ctx context.Context, args ...any) (T, error)

// Lazy represents a lazy-loaded value of type T. Calls to Value() will
// invoke the configured loader when needed and cache the result. Clear()
// allows clearing cached entries.
type Lazy[T any] interface {
	// Value returns the value, calling the loader if the value is not yet present
	// (or if TTL expired for TTL-enabled instances). While it accepts a variadic
	// context.Context parameter it only uses the first one (if supplied), so you
	// don't have to provide anything if you don't need it; if none is provided,
	// context.Background() is used.
	Value(ctxs ...context.Context) (T, error)

	// Clear marks the cached value as unloaded so the next Value() call will
	// invoke the loader again.
	Clear()
}

// WithLoader creates a Lazy[T] that will call the provided loader when
// Value() is invoked for the first time. Subsequent calls return the cached
// value unless Clear() is called. The args are forwarded to the loader on
// each invocation.
func WithLoader[T any](loader LazyFunc[T], args ...any) Lazy[T] {
	return newWithLoader(loader, false, 0, args...)
}

// WithLoaderTTL is like WithLoader but also enables TTL caching. Cached values
// are invalidated after ttl expires, then the provided loader will be invoked
// again to load a fresh value. The args are forwarded to the loader on each
// invocation.
func WithLoaderTTL[T any](loader LazyFunc[T], ttl time.Duration, args ...any) Lazy[T] {
	return newWithLoader(loader, true, ttl, args...)
}

// Preloaded returns a Lazy[T] pre-populated with value. The loader is still
// kept and may be used for other contexts or after Clear() is called. This is
// useful when you already have an initial value or a default to return. The args
// are forwarded to the loader when it runs.
func Preloaded[T any](value T, loader LazyFunc[T], args ...any) Lazy[T] {
	return newWithLoaderPreloaded(value, loader, false, 0, args...)
}

// PreloadedTTL is like Preloaded but also enables TTL caching. The initial value
// is returned immediately, and the loader may be invoked after the TTL expires
// to load a fresh value. The args are forwarded to the loader on each invocation.
func PreloadedTTL[T any](value T, loader LazyFunc[T], ttl time.Duration, args ...any) Lazy[T] {
	return newWithLoaderPreloaded(value, loader, true, ttl, args...)
}

// Static returns a Lazy[T] that always returns the provided value and never
// invokes a loader. This is a convenience for tests or fixed values. Clear()
// is a no-op for Static.
func Static[T any](value T) Lazy[T] {
	return &static[T]{
		value: value,
	}
}
