# golazy

golazy is a small, dependency-free Go library that provides a context-aware,
generic lazy-loading abstraction. It lets you declare values that are loaded on
first use (or preloaded) and optionally cached with a TTL. The implementation
is safe for concurrent use and keeps the public API intentionally small.

## Quick summary
- `LazyFunc[T]` — loader function type: `func(ctx context.Context, args ...any) (T, error)`
- `Lazy[T]` — main interface: `Value(ctxs ...context.Context) (T, error)` and `Clear()`
- Constructors: `WithLoader`, `WithLoaderTTL`, `Preloaded`, `PreloadedTTL`, `Static`

## Installation

Requires Go 1.19+. Install with:

```bash
go get github.com/duhnnie/golazy
```

## API reference & examples

### Loader signature

```go
type LazyFunc[T any] func(ctx context.Context, args ...any) (T, error)
```

The loader receives a `context.Context` (for cancellation/deadlines/values)
and zero or more `args` passed when the `Lazy` value was constructed.

### Constructors

- `WithLoader[T](loader LazyFunc[T], args ...any) Lazy[T]` — create a lazy value that calls `loader` on first use.
- `WithLoaderTTL[T](loader LazyFunc[T], ttl time.Duration, args ...any) Lazy[T]` — like `WithLoader` but caches the last successful value for `ttl`.
- `Preloaded[T](value T, loader LazyFunc[T], args ...any) Lazy[T]` — construct a `Lazy` already populated with `value`.
- `PreloadedTTL[T](value T, loader LazyFunc[T], ttl time.Duration, args ...any) Lazy[T]` — like `Preloaded` but preloads the value and enables TTL behavior.
- `Static[T](value T) Lazy[T]` — returns a `Lazy` that always yields `value` and never calls a loader.

### Basic usage

```go
loader := func(ctx context.Context, args ...any) (string, error) {
    // args[0] contains a key passed during construction
    key := args[0].(string)
    return "value-for-" + key, nil
}

lazy := golazy.WithLoader[string](loader, "alpha")
v, err := lazy.Value() // uses context.Background()
_ = v

// optionally pass a context:
v2, err := lazy.Value(context.WithValue(context.Background(), "k", "v"))
_ = v2

// clear cached value so next Value() triggers loader again
lazy.Clear()
```

### Using TTL

```go
lazyTTL := golazy.WithLoaderTTL[int](loaderFunc, 5*time.Second, 123)
```

### Preloaded & Static

```go
pre := golazy.Preloaded[string]("initial", loader)
v, _ := pre.Value()

st := golazy.Static(42)
v2, _ := st.Value()
```

### Clearing cache

Call `Clear()` on the `Lazy` value to mark it as unloaded. The next call to
`Value()` will run the loader again (or return the preloaded/static value).

```go
lazy.Clear()
```

## Notes & gotchas

- `Value` accepts zero or one `context.Context`. When omitted `context.Background()` is used.
- Constructors accept `args ...any` which are forwarded to the loader on every invocation. This lets you configure the loader with static parameters.
 - `PreloadedTTL` is implemented with the signature `PreloadedTTL[T](value T, loader LazyFunc[T], ttl time.Duration, args ...any)` and forwards to the internal constructor.

## Contributing

Contributions welcome — please follow the [guidelines](CONTRIBUTING.md).

## License

[MIT License](LICENSE)
