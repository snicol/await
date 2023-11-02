package await_test

import (
	"context"
	"testing"

	"github.com/snicol/await"
)

func TestCtxErrorGroupPanic(t *testing.T) {
	ctx := context.Background()

	g := await.Group(ctx)

	g.Go(func(ctx context.Context) error {
		panic("something went wrong")

		return nil
	})

	err := g.Wait()
	pe, ok := err.(await.PanicError)
	if !ok {
		t.Error("expected a panic error")
		return
	}

	if pe.Error() != "goroutine panicked: something went wrong" {
		t.Error("unexpected error found")
	}

	if pe.Panic != "something went wrong" {
		t.Error("expected underlying panic to be set")
	}
}
