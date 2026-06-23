package await

import (
	"context"
	"sync"
)

// SliceReturn takes a slice of any input, and asynchronously calls `fn` over
// every item, collecting the return values of each execution into a new slice.
// The returned slice has the same length and ordering as the input: result i
// corresponds to input i.
// If any error is encountered, the group's context is canceled and execution
// stops.
// The returned slice should not be accessed until the group's Wait has
// returned.
func SliceReturn[T, O any](g ctxErrgroup, vs []T, fn func(context.Context, T) (O, error)) []O {
	out := make([]O, len(vs))

	for i, v := range vs {
		g.Go(func(ctx context.Context) error {
			res, err := fn(ctx, v)
			if err != nil {
				return err
			}

			// Safe without a lock: every goroutine owns a unique index i.
			out[i] = res
			return nil
		})
	}

	return out
}

// MapReturn takes a map of any input, and asynchronously calls `fn` over every
// item. `fn` returns a (possibly new) key, a value, and an error; the returned
// key/value pair is collected into a new output map. This lets you transform
// both the keys and the values of a map concurrently.
// If any error is encountered, the group's context is canceled and execution
// stops.
// Writes back to the output map are guarded by a mutex because Go maps are not
// safe for concurrent writes. If `fn` returns the same key for two different
// inputs, the last write wins.
// The returned map should not be accessed until the group's Wait has returned.
func MapReturn[K comparable, V, VO any](g ctxErrgroup, vs map[K]V, fn func(context.Context, K, V) (K, VO, error)) map[K]VO {
	m := &sync.Mutex{}
	out := make(map[K]VO, len(vs))

	for k, v := range vs {
		g.Go(func(ctx context.Context) error {
			newKey, res, err := fn(ctx, k, v)
			if err != nil {
				return err
			}

			// Guard the map write: concurrent writes to a Go map panic.
			defer m.Unlock()
			m.Lock()

			out[newKey] = res
			return nil
		})
	}

	return out
}
