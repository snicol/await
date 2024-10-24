# await

`await` is a Go package that provides a simple and efficient way to handle asynchronous operations using context and error groups. It allows you to perform operations on slices and maps concurrently, with built-in error handling and context cancellation.

## Installation

To install the `await` package, run the following command:

```sh
go get github.com/snicol/await
```

## Usage

Here are some examples of how to use the `await` package:

### Example 1: Basic Usage

```go
package main

import (
	"context"
	"log"

	"github.com/snicol/await"
)

func main() {
	ctx := context.Background()

	// Construct a group.
	g := await.Group(
		ctx,
		await.WithLimit(10), // Max 10 concurrent routines.
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
		log.Println(err)
	}

	// Instead of creating groups, you can use the Context suite of functions.
	// These immediately execute a group, returning the error. You lose the ability
	// to chain multiple calls onto a single group, but it improves readability for
	// single calls.
	err := await.SliceReplaceContext(ctx, ids, func(_ context.Context, id string) (string, error) {
		return "replaced", nil
	})
	if err != nil {
		log.Println(err)
	}

	log.Println(ids) // [replaced replaced replaced]

	// Call a function over every item, the Response value returned is collected into responses.
	responses, err := await.SliceReturnContext(ctx, ids, func(ctx context.Context, id string) (Response, error) {
		return exampleClientCall(ctx, id), nil
	})
	if err != nil {
		log.Println(err)
	}

	log.Printf("%+v", responses) // [{ID:replaced Status:true} {ID:replaced Status:true} {ID:replaced Status:true}]
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
```

### Example 2: Handling Panics

```go
package main

import (
	"context"
	"log"

	"github.com/snicol/await"
)

func main() {
	ctx := context.Background()

	g := await.Group(ctx)

	g.Go(func(ctx context.Context) error {
		panic("something went wrong")

		return nil
	})

	err := g.Wait()
	pe, ok := err.(await.PanicError)
	if !ok {
		log.Println("expected a panic error")
		return
	}

	if pe.Error() != "goroutine panicked: something went wrong" {
		log.Println("unexpected error found")
	}

	if pe.Panic != "something went wrong" {
		log.Println("expected underlying panic to be set")
	}
}
```

## Overview of Main Functions and Methods

### Group

`Group` returns a new wrapped `errgroup.Group`, with the given context. It manages passing in context to avoid issues when not using the `errgroup`'s returned context.

### Any

`Any` takes any input, and asynchronously calls `fn`. If any error is encountered, the context is canceled and execution stops.

### Slice

`Slice` takes a slice of any input, and asynchronously calls `fn` over every item. If any error is encountered, the context is cancelled and execution stops.

### Map

`Map` takes a map of any input, and asynchronously calls `fn` over each key/value pair. If any error is encountered, the context is cancelled and execution stops.

### SliceReplace

`SliceReplace` takes a slice of any input, and asynchronously calls `fn` over every item, replacing the value in the slice with the one returned by `fn`. If any error is encountered, the context is cancelled and execution stops.

### MapReplace

`MapReplace` takes a map of any input, and asynchronously calls `fn` over every item, replacing the value in the map with the one returned by `fn`. If any error is encountered, the context is cancelled and execution stops.

### SliceReturn

`SliceReturn` takes a slice of any input, and asynchronously calls `fn` over every item, collecting the return values of each execution to be returned as a new slice when finished. If any error is encountered, the context is cancelled and execution stops.

### MapReturn

`MapReturn` takes a map of any input, and asynchronously calls `fn` over every item, returning a key/value pair or error. If any error is encountered, the context is cancelled and execution stops.

## License

MIT License

Copyright (c) 2023 Scott Nicol

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
