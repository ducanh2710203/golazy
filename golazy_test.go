package golazy

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestWithLoader tests the WithLoader constructor
func TestWithLoader_CallsLoader(t *testing.T) {
	t.Parallel()
	callCount := 0
	loader := func(ctx context.Context, args ...any) (string, error) {
		callCount++
		return "loaded", nil
	}

	lazy := WithLoader(loader)
	v, err := lazy.Value()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "loaded" {
		t.Fatalf("expected 'loaded', got %q", v)
	}
	if callCount != 1 {
		t.Fatalf("loader should be called once, was called %d times", callCount)
	}
}

func TestWithLoader_CachesValue(t *testing.T) {
	t.Parallel()
	callCount := 0
	loader := func(ctx context.Context, args ...any) (string, error) {
		callCount++
		return "value", nil
	}

	lazy := WithLoader(loader)
	v1, _ := lazy.Value()
	v2, _ := lazy.Value()
	v3, _ := lazy.Value()

	if v1 != v2 || v2 != v3 {
		t.Fatalf("all values should be identical")
	}
	if callCount != 1 {
		t.Fatalf("loader should be called once, was called %d times", callCount)
	}
}

func TestWithLoader_ForwardsArgs(t *testing.T) {
	t.Parallel()
	loader := func(ctx context.Context, args ...any) (string, error) {
		if len(args) < 2 {
			return "", errors.New("expected at least 2 args")
		}
		prefix := args[0].(string)
		suffix := args[1].(string)
		return prefix + suffix, nil
	}

	lazy := WithLoader(loader, "hello", "world")
	v, err := lazy.Value()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "helloworld" {
		t.Fatalf("expected 'helloworld', got %q", v)
	}
}

func TestWithLoader_ClearsCache(t *testing.T) {
	t.Parallel()
	callCount := 0
	loader := func(ctx context.Context, args ...any) (int, error) {
		callCount++
		return callCount, nil
	}

	lazy := WithLoader(loader)
	v1, _ := lazy.Value()
	if v1 != 1 {
		t.Fatalf("first call should return 1, got %d", v1)
	}

	lazy.Clear()
	v2, _ := lazy.Value()
	if v2 != 2 {
		t.Fatalf("after Clear, should call loader again and return 2, got %d", v2)
	}
}

func TestWithLoader_AcceptsContext(t *testing.T) {
	t.Parallel()
	ctxReceived := false
	loader := func(ctx context.Context, args ...any) (bool, error) {
		ctxReceived = ctx != nil
		return ctxReceived, nil
	}

	lazy := WithLoader(loader)
	customCtx := context.WithValue(context.Background(), "key", "value")
	v, err := lazy.Value(customCtx)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !v {
		t.Fatalf("context should have been passed to loader")
	}
}

func TestWithLoader_UsesBackgroundContextWhenOmitted(t *testing.T) {
	t.Parallel()
	ctxReceived := false
	loader := func(ctx context.Context, args ...any) (bool, error) {
		ctxReceived = ctx != nil
		return ctxReceived, nil
	}

	lazy := WithLoader(loader)
	v, err := lazy.Value() // no context provided

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !v {
		t.Fatalf("context.Background() should have been used")
	}
}

func TestWithLoader_PropagatesErrors(t *testing.T) {
	t.Parallel()
	loader := func(ctx context.Context, args ...any) (string, error) {
		return "", errors.New("loader failed")
	}

	lazy := WithLoader(loader)
	_, err := lazy.Value()

	if err == nil {
		t.Fatal("expected error from loader, got nil")
	}
	if err.Error() != "loader failed" {
		t.Fatalf("expected 'loader failed', got %q", err.Error())
	}
}

// TestWithLoaderTTL tests the WithLoaderTTL constructor
func TestWithLoaderTTL_CachesWithTTL(t *testing.T) {
	t.Parallel()
	callCount := 0
	loader := func(ctx context.Context, args ...any) (string, error) {
		callCount++
		return "value", nil
	}

	lazy := WithLoaderTTL(loader, 100*time.Millisecond)
	v1, _ := lazy.Value()
	v2, _ := lazy.Value()

	if v1 != v2 {
		t.Fatalf("values should be identical before TTL expires")
	}
	if callCount != 1 {
		t.Fatalf("loader should be called once, was called %d times", callCount)
	}
}

func TestWithLoaderTTL_ReloadsAfterTTLExpires(t *testing.T) {
	callCount := 0
	loader := func(ctx context.Context, args ...any) (int, error) {
		callCount++
		return callCount, nil
	}

	lazy := WithLoaderTTL(loader, 100*time.Millisecond)
	v1, _ := lazy.Value()
	if v1 != 1 {
		t.Fatalf("first call should return 1, got %d", v1)
	}

	time.Sleep(150 * time.Millisecond)
	v2, _ := lazy.Value()
	if v2 != 2 {
		t.Fatalf("after TTL expires, should reload and return 2, got %d", v2)
	}
}

func TestWithLoaderTTL_ForwardsArgs(t *testing.T) {
	t.Parallel()
	loader := func(ctx context.Context, args ...any) (int, error) {
		if len(args) == 0 {
			return 0, errors.New("expected args")
		}
		return args[0].(int) * 2, nil
	}

	lazy := WithLoaderTTL(loader, 1*time.Second, 21)
	v, err := lazy.Value()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 42 {
		t.Fatalf("expected 42, got %d", v)
	}
}

// TestPreloaded tests the Preloaded constructor
func TestPreloaded_ReturnsPreloadedValue(t *testing.T) {
	t.Parallel()
	preloadedValue := "preloaded"
	loaderCalled := false
	loader := func(ctx context.Context, args ...any) (string, error) {
		loaderCalled = true
		return "from-loader", nil
	}

	lazy := Preloaded(preloadedValue, loader)
	v, err := lazy.Value()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != preloadedValue {
		t.Fatalf("expected %q, got %q", preloadedValue, v)
	}
	if loaderCalled {
		t.Fatal("loader should not be called for preloaded value")
	}
}

func TestPreloaded_CallsLoaderAfterClear(t *testing.T) {
	t.Parallel()
	preloadedValue := "preloaded"
	loaderValue := "from-loader"
	loaderCallCount := 0
	loader := func(ctx context.Context, args ...any) (string, error) {
		loaderCallCount++
		return loaderValue, nil
	}

	lazy := Preloaded(preloadedValue, loader)
	v1, _ := lazy.Value()
	if v1 != preloadedValue {
		t.Fatalf("first call should return preloaded value")
	}

	lazy.Clear()
	v2, _ := lazy.Value()
	if v2 != loaderValue {
		t.Fatalf("after Clear, should call loader and return %q, got %q", loaderValue, v2)
	}
	if loaderCallCount != 1 {
		t.Fatalf("loader should be called once after Clear, was called %d times", loaderCallCount)
	}
}

func TestPreloaded_ForwardsArgs(t *testing.T) {
	t.Parallel()
	preloadedValue := "preloaded"
	loader := func(ctx context.Context, args ...any) (string, error) {
		if len(args) < 1 {
			return "", errors.New("expected args")
		}
		return args[0].(string), nil
	}

	lazy := Preloaded(preloadedValue, loader, "arg-value")
	v, _ := lazy.Value()

	// First call returns preloaded value
	if v != preloadedValue {
		t.Fatalf("expected preloaded %q, got %q", preloadedValue, v)
	}

	// After Clear, args should be forwarded
	lazy.Clear()
	v2, _ := lazy.Value()
	if v2 != "arg-value" {
		t.Fatalf("after Clear, expected 'arg-value', got %q", v2)
	}
}

// TestPreloadedTTL tests the PreloadedTTL constructor
func TestPreloadedTTL_ReturnsPreloadedValue(t *testing.T) {
	t.Parallel()
	preloadedValue := "preloaded"
	loader := func(ctx context.Context, args ...any) (string, error) {
		return "from-loader", nil
	}

	lazy := PreloadedTTL(preloadedValue, loader, 1*time.Second)
	v, err := lazy.Value()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != preloadedValue {
		t.Fatalf("expected %q, got %q", preloadedValue, v)
	}
}

func TestPreloadedTTL_ReloadsAfterTTLExpires(t *testing.T) {
	preloadedValue := "preloaded"
	callCount := 0
	loader := func(ctx context.Context, args ...any) (string, error) {
		callCount++
		return "fresh", nil
	}

	lazy := PreloadedTTL(preloadedValue, loader, 100*time.Millisecond)
	v1, _ := lazy.Value()
	if v1 != preloadedValue {
		t.Fatalf("first call should return preloaded value")
	}

	time.Sleep(150 * time.Millisecond)
	v2, _ := lazy.Value()
	if v2 != "fresh" {
		t.Fatalf("after TTL expires, should reload and return 'fresh', got %q", v2)
	}
	if callCount != 1 {
		t.Fatalf("loader should be called once after TTL, was called %d times", callCount)
	}
}

func TestPreloadedTTL_ForwardsArgs(t *testing.T) {
	t.Parallel()
	preloadedValue := "preloaded"
	loader := func(ctx context.Context, args ...any) (string, error) {
		if len(args) < 1 {
			return "", errors.New("expected args")
		}
		return args[0].(string), nil
	}

	lazy := PreloadedTTL(preloadedValue, loader, 1*time.Second, "arg-value")
	v1, _ := lazy.Value()

	// First call returns preloaded value
	if v1 != preloadedValue {
		t.Fatalf("expected preloaded %q, got %q", preloadedValue, v1)
	}

	// After Clear, args should be forwarded
	lazy.Clear()
	v2, _ := lazy.Value()
	if v2 != "arg-value" {
		t.Fatalf("after Clear, expected 'arg-value', got %q", v2)
	}
}

// TestStatic tests the Static constructor
func TestStatic_AlwaysReturnsValue(t *testing.T) {
	t.Parallel()
	value := "constant"
	lazy := Static(value)

	v1, err1 := lazy.Value()
	v2, err2 := lazy.Value()
	v3, err3 := lazy.Value()

	if err1 != nil || err2 != nil || err3 != nil {
		t.Fatal("Static should never return an error")
	}
	if v1 != value || v2 != value || v3 != value {
		t.Fatalf("all calls should return %q", value)
	}
}

func TestStatic_ClearIsNoOp(t *testing.T) {
	t.Parallel()
	value := 42
	lazy := Static(value)

	v1, _ := lazy.Value()
	lazy.Clear()
	v2, _ := lazy.Value()

	if v1 != v2 || v2 != value {
		t.Fatalf("Clear should be a no-op, expected %d for both calls", value)
	}
}

func TestStatic_IgnoresContext(t *testing.T) {
	t.Parallel()
	value := "constant"
	lazy := Static(value)

	contexts := []context.Context{
		context.Background(),
		context.WithValue(context.Background(), "key", "value"),
	}

	for _, ctx := range contexts {
		v, err := lazy.Value(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if v != value {
			t.Fatalf("expected %q, got %q", value, v)
		}
	}
}

// TestLazyFuncFor tests the LazyFuncFor helper
func TestLazyFuncFor_ReturnsValue(t *testing.T) {
	t.Parallel()
	value := "test-value"
	lazyFunc := LazyFuncFor(value)

	result, err := lazyFunc(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != value {
		t.Fatalf("expected %q, got %q", value, result)
	}
}

func TestLazyFuncFor_IgnoresContext(t *testing.T) {
	t.Parallel()
	value := 42
	lazyFunc := LazyFuncFor(value)

	contexts := []context.Context{
		context.Background(),
		context.WithValue(context.Background(), "key", "value"),
	}

	for _, ctx := range contexts {
		result, err := lazyFunc(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != value {
			t.Fatalf("expected %d, got %d", value, result)
		}
	}
}

func TestLazyFuncFor_IgnoresArgs(t *testing.T) {
	t.Parallel()
	value := []string{"a", "b", "c"}
	lazyFunc := LazyFuncFor(value)

	result, err := lazyFunc(context.Background(), "arg1", "arg2", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != len(value) {
		t.Fatalf("expected length %d, got %d", len(value), len(result))
	}
}

func TestLazyFuncFor_WithWithLoader(t *testing.T) {
	t.Parallel()
	value := "constant"
	lazyFunc := LazyFuncFor(value)
	lazy := WithLoader(lazyFunc)

	v, err := lazy.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != value {
		t.Fatalf("expected %q, got %q", value, v)
	}
}

func TestLazyFuncFor_WithPreloaded(t *testing.T) {
	t.Parallel()
	preloadedValue := "preloaded"
	loaderValue := "from-loader"
	lazyFunc := LazyFuncFor(loaderValue)
	lazy := Preloaded(preloadedValue, lazyFunc)

	// First call should return preloaded value
	v, err := lazy.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != preloadedValue {
		t.Fatalf("expected %q, got %q", preloadedValue, v)
	}

	// After Clear, should call the loader
	lazy.Clear()
	v2, err := lazy.Value()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v2 != loaderValue {
		t.Fatalf("expected %q after Clear, got %q", loaderValue, v2)
	}
}
