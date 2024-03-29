package await

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// ctxErrgroup is a wrapper around errgroup. It manages passing in context
// to avoid issues when not using the errgroup's returned context.
type ctxErrgroup struct {
	ctx context.Context
	g   *errgroup.Group
}

// Option defines a function which can be passed to Group.
// See the available options:
//   - WithLimit
type Option func(ctxErrgroup)

// Group returns a new wrapped errgroup.Group, with the given context.
func Group(ctx context.Context, options ...Option) ctxErrgroup {
	g, gctx := errgroup.WithContext(ctx)

	ceg := ctxErrgroup{
		gctx, g,
	}

	for _, opt := range options {
		opt(ceg)
	}

	return ceg
}

// WithLimit sets the maximum amount of goroutines running in the errgroup at
// any point in time.
func WithLimit(limit int) Option {
	return func(ceg ctxErrgroup) {
		ceg.g.SetLimit(limit)
	}
}

// PanicError wraps a recovered panic and is returned if a goroutine panics.
type PanicError struct {
	Panic any
}

var _ error = PanicError{}

// Error conforms to the error interface.
func (pe PanicError) Error() string {
	return fmt.Sprintf("goroutine panicked: %s", pe.Panic)
}

// Go adds a call to the errgroup, which passes the context provided to Group
// into fn. Any error returned from the function provided will stop execution
// for the entire group.
func (ceg ctxErrgroup) Go(fn func(ctx context.Context) error) {
	ceg.g.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = PanicError{Panic: r}
			}
		}()

		err = fn(ceg.ctx)
		return err
	})
}

// Wait starts the group, and runs all registered functions. If any fail, the
// error will be returned and all other functions will be canceled.
func (ceg ctxErrgroup) Wait() error {
	return ceg.g.Wait()
}
