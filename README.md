# await

[![Test](https://github.com/snicol/await/actions/workflows/test.yml/badge.svg)](https://github.com/snicol/await/actions/workflows/test.yml)
[![Lint](https://github.com/snicol/await/actions/workflows/lint.yml/badge.svg)](https://github.com/snicol/await/actions/workflows/lint.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/snicol/await.svg)](https://pkg.go.dev/github.com/snicol/await)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`await` is a tiny, generic Go library for running work concurrently over a
single value, a slice, or a map. It is a thin, type-safe layer on top of
[`golang.org/x/sync/errgroup`](https://pkg.go.dev/golang.org/x/sync/errgroup)
that removes the boilerplate of spawning goroutines, wiring up a shared context,
and collecting results — while keeping the same fail-fast error semantics you
already expect from `errgroup`.

## Features

- **Generics-first.** Works over `[]T`, `map[K]V`, or any single `T` with no
  `interface{}` and no type assertions.
- **Fail-fast.** The first non-nil error cancels the shared context and is
  returned from `Wait`, just like `errgroup`.
- **Panic-safe.** A panic in a worker goroutine is recovered and returned as a
  `PanicError` instead of crashing your process.
- **Result collection without the bookkeeping.** Replace values in place
  (`SliceReplace` / `MapReplace`) or collect new results (`SliceReturn` /
  `MapReturn`) without writing your own mutexes or index juggling.
- **Concurrency limiting.** Cap the number of in-flight goroutines with
  `WithLimit`.
- **Two ergonomics.** Build a `Group` and chain several calls onto it, or use
  the one-shot `*Context` helpers for single calls.

## Installation

```sh
go get github.com/snicol/await
```

Requires Go 1.25 or newer (per [go.mod](go.mod)).

```go
import "github.com/snicol/await"
```

## Quick start

```go
ctx := context.Background()

ids := []string{"id_a", "id_b", "id_c"}

// Run fn over every item, concurrently. Returns the first error (or nil).
err := await.SliceContext(ctx, ids, func(ctx context.Context, id string) error {
    return process(ctx, id)
})
if err != nil {
    log.Fatal(err)
}
```

## Concepts

### Groups vs. `*Context` helpers

There are two ways to use the library:

1. **Build a `Group` and register work on it.** This lets you fan out several
   different operations onto a single group and wait for all of them together.

   ```go
   g := await.Group(ctx, await.WithLimit(10)) // optional concurrency cap

   await.Slice(g, ids, fetchOne)              // fan out over a slice
   await.Any(g, &report, buildReport)         // plus a single one-off call

   if err := g.Wait(); err != nil {           // run everything, wait, return
       return err
   }
   ```

2. **Use the one-shot `*Context` helpers.** These construct a group, register
   the work, and call `Wait` for you. You lose the ability to chain multiple
   calls onto one group, but they read better for single operations.

   ```go
   err := await.SliceContext(ctx, ids, fetchOne)
   ```

Nothing runs until `Wait` is called (directly, or by a `*Context` helper).
`Wait` blocks until every registered function has returned, then returns the
first error encountered — or nil if all succeeded.

### Error handling and cancellation

All work registered on a group shares a single context derived from the one you
pass to `Group`. The moment any function returns a non-nil error, that context
is cancelled, signalling the other in-flight functions to stop early (provided
they respect `ctx.Done()`). `Wait` returns that first error.

### Panic recovery

If a worker function panics, `await` recovers the panic and returns it from
`Wait` as a `PanicError`, rather than letting it unwind and crash the process:

```go
err := await.SliceContext(ctx, ids, mightPanic)

var pe await.PanicError
if errors.As(err, &pe) {
    log.Printf("a worker panicked: %v", pe.Panic)
}
```

## API overview

### Constructing a group

| Function | Description |
|----------|-------------|
| `Group(ctx, ...Option) ctxErrgroup` | Create a group bound to `ctx`. |
| `WithLimit(n) Option` | Cap concurrent goroutines at `n` (non-positive = unlimited). |

### Registering work on a group

| Function | Input | Behaviour |
|----------|-------|-----------|
| `Any(g, v, fn)` | a single `T` | Run `fn(ctx, v)` once. |
| `Slice(g, vs, fn)` | `[]T` | Run `fn(ctx, v)` for each element. |
| `Map(g, vs, fn)` | `map[K]V` | Run `fn(ctx, k, v)` for each pair. |
| `SliceReplace(g, vs, fn)` | `[]T` | Replace each element in place with `fn`'s result. |
| `MapReplace(g, vs, fn)` | `map[K]V` | Replace each value in place with `fn`'s result. |
| `SliceReturn(g, vs, fn) []O` | `[]T` | Collect `fn`'s results into a new, order-preserving slice. |
| `MapReturn(g, vs, fn) map[K]VO` | `map[K]V` | Collect `fn`'s (key, value) results into a new map. |

For `*Return` helpers, do not read the returned slice/map until `Wait` has
returned. Ordering of `SliceReturn` matches the input: result `i` corresponds to
input `i`.

### One-shot `*Context` helpers

Each constructs a group, registers the work, and calls `Wait` for you:

| Function | Equivalent to |
|----------|---------------|
| `SliceContext(ctx, vs, fn) error` | `Group` + `Slice` + `Wait` |
| `MapContext(ctx, vs, fn) error` | `Group` + `Map` + `Wait` |
| `SliceReplaceContext(ctx, vs, fn) error` | `Group` + `SliceReplace` + `Wait` |
| `SliceReturnContext(ctx, vs, fn) ([]O, error)` | `Group` + `SliceReturn` + `Wait` |

## Examples

### Collecting results from a slice

```go
ids := []string{"id_a", "id_b", "id_c"}

responses, err := await.SliceReturnContext(ctx, ids,
    func(ctx context.Context, id string) (Response, error) {
        return client.Get(ctx, id)
    },
)
if err != nil {
    return err
}
// responses[i] corresponds to ids[i].
```

### Replacing slice values in place

```go
// Concurrently transform every element, mutating the slice in place.
err := await.SliceReplaceContext(ctx, ids,
    func(ctx context.Context, id string) (string, error) {
        return normalize(ctx, id)
    },
)
```

### Transforming a map

```go
g := await.Group(ctx)

// fn returns a (possibly new) key, a value, and an error.
out := await.MapReturn(g, in,
    func(ctx context.Context, k string, v int) (string, string, error) {
        return strings.ToUpper(k), strconv.Itoa(v), nil
    },
)

if err := g.Wait(); err != nil {
    return err
}
// out is safe to read now.
```

A fuller, runnable example lives in [await_test.go](await_test.go).

## Thread safety notes

- `Slice`, `SliceReplace`, and `SliceReturn` each have every goroutine write to
  a distinct slice index, so no locking is required.
- `MapReplace` and `MapReturn` guard their map writes with a mutex, because Go
  maps are not safe for concurrent writes.
- `Any` and `Slice` make **no** guarantees about safe concurrent access to your
  values. If your `fn` mutates shared state, you are responsible for
  synchronizing it. Use the `*Replace` / `*Return` helpers when you need results
  written back safely.

## Development

```sh
go test ./...     # run tests
go vet ./...      # vet
golangci-lint run # lint (config in .golangci.yml)
```

CI runs tests and linting on every push via the workflows in
[.github/workflows](.github/workflows).

## License

Released under the [MIT License](LICENSE).
