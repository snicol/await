package await

import (
	"context"
	"sync"
)

// SliceReplace takes a slice of any input, and asynchronously calls `fn` over
// every item, replacing the value in the slice with the one returned by `fn`.
// If any error is encountered, the group's context is cancelled and execution
// stops.
// This function is useful if you wish to modify the values of the slice as a
// result of an asynchronous call. Each goroutine writes to a distinct index, so
// no locking is required and the original slice is mutated in place.
// The slice should not be read until the group's Wait has returned.
func SliceReplace[T any](g ctxErrgroup, vs []T, fn func(context.Context, T) (T, error)) {
	for i, v := range vs {
		g.Go(func(ctx context.Context) error {
			res, err := fn(ctx, v)
			if err != nil {
				return err
			}

			// Safe without a lock: every goroutine owns a unique index i.
			vs[i] = res
			return nil
		})
	}
}

// MapReplace takes a map of any input, and asynchronously calls `fn` over
// every item, replacing the value in the map with the one returned by `fn`.
// If any error is encountered, the group's context is cancelled and execution
// stops.
// This function is useful if you wish to modify the values of the map as a
// result of an asynchronous call. Writes back to the map are guarded by a mutex
// because Go maps are not safe for concurrent writes.
// The map should not be read until the group's Wait has returned.
func MapReplace[K comparable, V any](g ctxErrgroup, vs map[K]V, fn func(context.Context, K, V) (V, error)) {
	m := &sync.Mutex{}

	for k, v := range vs {
		g.Go(func(ctx context.Context) error {
			res, err := fn(ctx, k, v)
			if err != nil {
				return err
			}

			// Guard the map write: concurrent writes to a Go map panic.
			defer m.Unlock()
			m.Lock()

			vs[k] = res
			return nil
		})
	}
}
