package golazy

import "context"

// static is a simple Lazy[T] implementation that always returns a fixed value.
// It satisfies the Lazy[T] interface but does not perform any loading or
// caching behavior (Clear is a no-op).
type static[T any] struct {
	// value is the constant value returned by Value()
	value T
}

// Value returns the static value and a nil error. The context argument is ignored.
func (l *static[T]) Value(ctxs ...context.Context) (T, error) {
	return l.value, nil
}

// Clear is a no-op for static values since there is nothing to clear.
func (l *static[T]) Clear() {
	// nothing
}
