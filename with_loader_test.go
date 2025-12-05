package golazy

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// Internal tests for newWithLoader and withLoader implementation
func TestNewWithLoader(t *testing.T) {
	t.Parallel()
	loader := func(ctx context.Context, args ...any) (string, error) {
		return "value", nil
	}
	wl := newWithLoader(loader, false, 0)

	if wl.loader == nil {
		t.Error("loader should not be nil")
	}
	if wl.withTTL != false {
		t.Error("withTTL should be false")
	}
}

func TestNewWithLoaderPreloaded(t *testing.T) {
	t.Parallel()
	loader := func(ctx context.Context, args ...any) (string, error) {
		return "new_value", nil
	}
	wl := newWithLoaderPreloaded("preloaded", loader, false, 0)

	if !wl.loaded && wl.value != "preloaded" {
		t.Error("preloaded value not set correctly")
	}
}

func TestValue_CachedValue(t *testing.T) {
	t.Parallel()
	callCount := 0
	loader := func(ctx context.Context, args ...any) (string, error) {
		callCount++
		return "value", nil
	}
	wl := newWithLoader(loader, false, 0)

	v1, err := wl.Value()
	if err != nil || v1 != "value" || callCount != 1 {
		t.Error("first call should invoke loader")
	}

	v2, err := wl.Value()
	if err != nil || v2 != "value" || callCount != 1 {
		t.Error("second call should use cache")
	}
}

func TestValue_LoaderError(t *testing.T) {
	t.Parallel()
	loader := func(ctx context.Context, args ...any) (string, error) {
		return "", errors.New("loader error")
	}
	wl := newWithLoader(loader, false, 0)

	_, err := wl.Value()
	if err == nil || err.Error() != "loader error" {
		t.Error("error should be propagated")
	}
}

func TestValue_TTLExpiration(t *testing.T) {
	callCount := 0
	loader := func(ctx context.Context, args ...any) (string, error) {
		callCount++
		return "value", nil
	}
	wl := newWithLoader(loader, true, 100*time.Millisecond)

	wl.Value()
	if callCount != 1 {
		t.Error("loader should be called once")
	}

	wl.Value()
	if callCount != 1 {
		t.Error("cached value should be used")
	}

	time.Sleep(150 * time.Millisecond)
	wl.Value()
	if callCount != 2 {
		t.Error("expired cache should trigger reload")
	}
}

func TestClear(t *testing.T) {
	t.Parallel()
	loader := func(ctx context.Context, args ...any) (string, error) {
		return "value", nil
	}
	wl := newWithLoader(loader, false, 0)
	wl.Value()

	if !wl.loaded {
		t.Error("value should be loaded")
	}

	wl.Clear()
	if wl.loaded {
		t.Error("value should be cleared")
	}
}

func TestConcurrency(t *testing.T) {
	callCount := 0
	loader := func(ctx context.Context, args ...any) (string, error) {
		callCount++
		time.Sleep(10 * time.Millisecond)
		return "value", nil
	}
	wl := newWithLoader(loader, false, 0)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wl.Value()
		}()
	}
	wg.Wait()

	if callCount != 1 {
		t.Errorf("loader should be called once, but was called %d times", callCount)
	}
}
