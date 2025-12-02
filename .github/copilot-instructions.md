# Copilot instructions for `golazy`

Short, actionable guidance for AI coding agents working on this repository.

Overview
- **Purpose**: `golazy` provides a context-aware generic `Lazy[T]` helper to lazily load values and optionally cache them with a TTL.
- **Core files**: `golazy.go` (public API), `with_loader.go` (implementation & TTL), `static.go` (static value implementation).

Important API notes
- **Loader signature**: `type LazyFunc[T any] func(ctx context.Context, args ...any) (T, error)` — the loader receives a `context.Context` and any constructor `args`.
- **Lazy interface**: `Value(ctxs ...context.Context) (T, error)` accepts zero or one `context.Context` (uses `context.Background()` when none supplied). `Clear()` marks the value as unloaded.
- **Constructors**:
  - `WithLoader[T](loader LazyFunc[T], args ...any) Lazy[T]`
  - `WithLoaderTTL[T](loader LazyFunc[T], ttl time.Duration, args ...any) Lazy[T]`
  - `Preloaded[T](value T, loader LazyFunc[T], args ...any) Lazy[T]`
  - `PreloadedTTL` currently has an inconsistent signature in the source (see README note)
  - `Static[T](value T) Lazy[T]`

Code-generation constraints & patterns
- Preserve the context-aware loader behavior: always pass a `context.Context` as first loader argument.
- Constructors forward `args` to the loader; keep that forwarding behavior when modifying constructors or internal code.
- Concurrency: `with_loader.go` uses a `sync.Mutex` to serialize loads and to protect internal state. Preserve mutex-based semantics or update tests when changing.
- TTL: `WithLoaderTTL`/`PreloadedTTL` set `withTTL` and `ttl` fields; expiration uses `lastLoad` timestamp.

Developer workflows (commands)
- Run tests: `make test` (runs `go test ./... -v`)
- Format: `make fmt` (`go fmt ./...`)
- Lint: `make lint` (`go vet` and `staticcheck` if available)

Where to look for examples
- `with_loader.go` demonstrates loader invocation and TTL handling.
- `static.go` shows the no-op `Clear` / constant `Value` behaviour.
- `README.md` contains usage examples reflecting the current API.

Notes for maintainers
- The repository currently contains an odd `PreloadedTTL` signature (`loader, value, ctx, ttl, args`) that doesn't match the internal `newWithLoaderPreloaded` ordering. If you change the public API, update callers and tests accordingly.
- If you add new features that change loader invocation (e.g., per-key caching or parallel loads), add tests that validate concurrency (`sync` usage) and TTL behaviour.

If anything here is unclear, tell me which API or example you want expanded and I'll update this guidance.
