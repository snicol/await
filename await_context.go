package await

import "context"

// SliceContext is a convenience method which calls Slice, but constructs the
// Group for you. This is useful if you do not require chaining the group with
// any other types.
func SliceContext[T any](ctx context.Context, vs []T, fn func(context.Context, T) error) error {
	g := Group(ctx)

	Slice(g, vs, fn)

	return g.Wait()
}

// SliceReplaceContext is a convenience method which calls SliceReplace,
// but constructs the Group for you. This is useful if you do not require
// chaining the group with any other types.
func SliceReplaceContext[T any](ctx context.Context, vs []T, fn func(context.Context, T) (T, error)) error {
	g := Group(ctx)

	SliceReplace(g, vs, fn)

	return g.Wait()
}

// SliceReplaceContext is a convenience method which calls SliceReplace,
// but constructs the Group for you. This is useful if you do not require
// chaining the group with any other types.
func SliceReturnContext[T, O any](ctx context.Context, vs []T, fn func(context.Context, T) (O, error)) ([]O, error) {
	g := Group(ctx)

	o := SliceReturn(g, vs, fn)

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return o, nil
}

// MapContext is a convenience method which calls Map, but constructs the
// Group for you. This is useful if you do not require chaining the group with
// any other types.
func MapContext[K comparable, V any](ctx context.Context, vs map[K]V, fn func(context.Context, K, V) error) error {
	g := Group(ctx)

	Map(g, vs, fn)

	return g.Wait()
}
