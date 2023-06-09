package await

import (
	"context"
)

// Any takes any input, and asyncronously calls `fn`.
// If any error is encoutered, the context is canceled and execution stops.
// n.b. this method does not provide any thread-safe guarantees on T.
func Any[T any](g ctxErrgroup, v T, fn func(context.Context, T) error) {
	g.Go(func(ctx context.Context) error {
		return fn(ctx, v)
	})
}

// Slice takes a slice of any input, and asyncronously calls `fn` over
// every item. If any error is encoutered, the context is cancelled and
// execution stops.
func Slice[T any](g ctxErrgroup, vs []T, fn func(context.Context, T) error) {
	for _, v := range vs {
		v := v

		Any(g, v, fn)
	}
}

// Map takes a map of any input, and asyncronously calls `fn` over each K/V
// pair. If any error is encoutered, the context is cancelled and execution
// stops.
func Map[K comparable, V any](g ctxErrgroup, vs map[K]V, fn func(context.Context, K, V) error) {
	for k, v := range vs {
		k, v := k, v

		g.Go(func(ctx context.Context) error {
			return fn(ctx, k, v)
		})
	}
}
