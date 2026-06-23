// Package await provides small, generic helpers for fanning work out across
// goroutines on top of golang.org/x/sync/errgroup. It adds three things on top
// of a plain errgroup: Go-generics helpers for running a function over a single
// value, a slice, or a map; helpers that collect or replace results without you
// having to manage synchronization; and automatic recovery of panics in worker
// goroutines (surfaced as a PanicError instead of crashing the process).
//
// All helpers share the group's context, so the first non-nil error (or panic)
// cancels that context and Wait returns it.
package await

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

// ctxErrgroup is a wrapper around errgroup.Group. It carries the context
// returned by errgroup.WithContext alongside the group itself so callers do
// not accidentally use a different (uncancelled) context inside their workers,
// a common source of bugs when using errgroup directly.
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

// WithLimit sets the maximum number of goroutines that may run concurrently in
// the group at any point in time. A non-positive limit means no limit. See
// errgroup.Group.SetLimit for the underlying semantics.
func WithLimit(limit int) Option {
	return func(ceg ctxErrgroup) {
		ceg.g.SetLimit(limit)
	}
}

// PanicError wraps a value recovered from a panicking worker goroutine. When a
// function passed to Go panics, await recovers it and returns it as a
// PanicError from Wait rather than letting the panic crash the process. The
// original recovered value is available via the Panic field.
type PanicError struct {
	Panic any
}

var _ error = PanicError{}

// Error conforms to the error interface.
func (pe PanicError) Error() string {
	return fmt.Sprintf("goroutine panicked: %s", pe.Panic)
}

// Go adds a call to the group, passing the group's context into fn. Any error
// returned from fn cancels the group's context and stops execution for the
// entire group. If fn panics, the panic is recovered and returned from Wait as
// a PanicError so a single bad worker cannot take down the whole process.
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
