package await

import (
	"context"
	"sync"
)

// SliceReturn takes a slice of any input, and asyncronously calls `fn` over
// every item, collecting the return values of each execution to be returned as
// a new slice when finished.
// If any error is encoutered, the context is cancelled and execution stops.
// The returned slice should not be accessed until the the group's Wait has
// finished.
func SliceReturn[T, O any](g ctxErrgroup, vs []T, fn func(context.Context, T) (O, error)) []O {
	out := make([]O, len(vs))

	for i, v := range vs {
		i, v := i, v

		g.Go(func(ctx context.Context) error {
			res, err := fn(ctx, v)
			if err != nil {
				return err
			}

			out[i] = res
			return nil
		})
	}

	return out
}

// MapReturn takes a map of any input, and asyncronously calls `fn` over
// every item, returning a key, value pair or error. If
// If any error is encoutered, the context is cancelled and execution stops.
// This function is useful if you wish to modify the values of the slice as a
// result of an asynchronous call, without changing the underlying pointer
// value which could result in data races.
// The returned map should not be accessed until the the group's Wait has
// finished.
func MapReturn[K comparable, V, VO any](g ctxErrgroup, vs map[K]V, fn func(context.Context, K, V) (K, VO, error)) map[K]VO {
	m := &sync.Mutex{}
	out := make(map[K]VO, len(vs))

	for k, v := range vs {
		k, v := k, v

		g.Go(func(ctx context.Context) error {
			newKey, res, err := fn(ctx, k, v)
			if err != nil {
				return err
			}
			defer m.Unlock()
			m.Lock()

			out[newKey] = res
			return nil
		})
	}

	return out
}
