package await

import "context"

// SliceContext is a convenience method which calls Slice, but constructs the
// Group and Waits on it for you. This is useful if you do not require chaining
// other calls onto the same group: it runs fn over every item and returns the
// first error (or nil).
func SliceContext[T any](ctx context.Context, vs []T, fn func(context.Context, T) error) error {
	g := Group(ctx)

	Slice(g, vs, fn)

	return g.Wait()
}

// SliceReplaceContext is a convenience method which calls SliceReplace, but
// constructs the Group and Waits on it for you. This is useful if you do not
// require chaining other calls onto the same group. The input slice is mutated
// in place once Wait returns without error.
func SliceReplaceContext[T any](ctx context.Context, vs []T, fn func(context.Context, T) (T, error)) error {
	g := Group(ctx)

	SliceReplace(g, vs, fn)

	return g.Wait()
}

// SliceReturnContext is a convenience method which calls SliceReturn, but
// constructs the Group and Waits on it for you. It returns the collected
// results (in input order) or the first error encountered. This is useful if
// you do not require chaining other calls onto the same group.
func SliceReturnContext[T, O any](ctx context.Context, vs []T, fn func(context.Context, T) (O, error)) ([]O, error) {
	g := Group(ctx)

	o := SliceReturn(g, vs, fn)

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return o, nil
}

// MapContext is a convenience method which calls Map, but constructs the Group
// and Waits on it for you. It runs fn over every K/V pair and returns the first
// error (or nil). This is useful if you do not require chaining other calls
// onto the same group.
func MapContext[K comparable, V any](ctx context.Context, vs map[K]V, fn func(context.Context, K, V) error) error {
	g := Group(ctx)

	Map(g, vs, fn)

	return g.Wait()
}
