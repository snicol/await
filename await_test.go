package await_test

import (
	"context"
	"log"
	"testing"

	"github.com/snicol/await"
)

func TestExample(t *testing.T) {
	ctx := context.Background()

	// Construct a group.
	g := await.Group(
		ctx,
		await.WithLimit(10), // Max 10 concurrenct routines.
	)

	// Input slice, can be []T where T = any.
	ids := []string{
		"id_a", "id_b", "id_c",
	}

	// Add a call to be performed on the group.
	// This will iterate over the slice items and call func on each one.
	await.Slice(g, ids, func(_ context.Context, v string) error {
		log.Println(v)

		return nil
	})

	type Test struct {
		foo string
	}

	v := &Test{}

	// Add another call to the group.
	// This executes a single function
	await.Any(g, v, func(_ context.Context, v *Test) error {
		v.foo = "test"
		return nil
	})

	// Execute all calls in the group and wait for completion.
	if err := g.Wait(); err != nil {
		t.Log(err)
	}

	// Instead of creating groups, you can use the Context suite of functions.
	// These immediately execute a group, returning the error. You lose the ability
	// to chain multiple calls onto a single group, but it improves readability for
	// single calls.
	err := await.SliceReplaceContext(ctx, ids, func(_ context.Context, id string) (string, error) {
		return "replaced", nil
	})
	if err != nil {
		t.Log(err)
	}

	log.Println(ids) // [replaced replaced replaced]

	// Call a function over every item, the Response value returned is collected into responses.
	responses, err := await.SliceReturnContext(ctx, ids, func(ctx context.Context, id string) (Response, error) {
		return exampleClientCall(ctx, id), nil
	})
	if err != nil {
		t.Log(err)
	}

	t.Logf("%+v", responses) // [{ID:replaced Status:true} {ID:replaced Status:true} {ID:replaced Status:true}]
}

type Response struct {
	ID     string `json:"id"`
	Status bool   `json:"status"`
}

func exampleClientCall(_ context.Context, id string) Response {
	return Response{
		ID:     id,
		Status: true,
	}
}
