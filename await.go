package await

import (
	"context"
)

// Any takes any input, and asynchronously calls `fn` with it on the group.
// If any error is encountered, the group's context is canceled and execution
// stops for every other call registered on the group.
// n.b. this method does not provide any thread-safe guarantees on T; if fn
// mutates T concurrently with other goroutines you must synchronize access
// yourself (see SliceReplace/MapReplace for safe mutation helpers).
func Any[T any](g ctxErrgroup, v T, fn func(context.Context, T) error) {
	g.Go(func(ctx context.Context) error {
		return fn(ctx, v)
	})
}

// Slice takes a slice of any input, and asynchronously calls `fn` once for
// every item, each in its own goroutine on the group. If any error is
// encountered, the group's context is canceled and execution stops.
func Slice[T any](g ctxErrgroup, vs []T, fn func(context.Context, T) error) {
	for _, v := range vs {
		Any(g, v, fn)
	}
}

// Map takes a map of any input, and asynchronously calls `fn` once for each
// K/V pair, each in its own goroutine on the group. If any error is
// encountered, the group's context is canceled and execution stops.
func Map[K comparable, V any](g ctxErrgroup, vs map[K]V, fn func(context.Context, K, V) error) {
	for k, v := range vs {
		g.Go(func(ctx context.Context) error {
			return fn(ctx, k, v)
		})
	}
}
