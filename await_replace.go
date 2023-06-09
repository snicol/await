package await

import (
	"context"
	"sync"
)

// SliceReplace takes a slice of any input, and asyncronously calls `fn` over
// every item, replacing the value in the slice with the one returned by `fn`.
// If any error is encoutered, the context is cancelled and execution stops.
// This function is useful if you wish to modify the values of the slice as a
// result of an asynchronous call, without changing the underlying pointer
// value which could result in data races.
func SliceReplace[T any](g ctxErrgroup, vs []T, fn func(context.Context, T) (T, error)) {
	for i, v := range vs {
		i, v := i, v

		g.Go(func(ctx context.Context) error {
			res, err := fn(ctx, v)
			if err != nil {
				return err
			}

			vs[i] = res
			return nil
		})
	}
}

// MapReplace takes a map of any input, and asyncronously calls `fn` over
// every item, replacing the value in the map with the one returned by `fn`.
// If any error is encoutered, the context is cancelled and execution stops.
// This function is useful if you wish to modify the values of the map as a
// result of an asynchronous call, as all operations are wrapped in a mutex for
// safety.
func MapReplace[K comparable, V any](g ctxErrgroup, vs map[K]V, fn func(context.Context, K, V) (V, error)) {
	m := &sync.Mutex{}

	for k, v := range vs {
		k, v := k, v

		g.Go(func(ctx context.Context) error {
			res, err := fn(ctx, k, v)
			if err != nil {
				return err
			}
			defer m.Unlock()
			m.Lock()

			vs[k] = res
			return nil
		})
	}
}
