package main

import (
	"context"
	"fmt"
	"github.com/doraemonkeys/fast-mpmc"
	"time"
)

func Example() {
	// Create a new Fast-MPMC queue with a minimum buffer capacity of 10
	queue := mpmc.NewFastMpmc[int](10)

	// Push elements to the queue
	queue.Push(1, 2, 3)

	// Pop all elements from the queue
	elements := queue.WaitPopAll()
	fmt.Println(*elements) // [1 2 3]

	// Recycle the buffer
	queue.RecycleBuffer(elements)

	// Push more elements to the queue
	queue.Push(4, 5, 6)

	// Pop all elements with a context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	elements, ok := queue.WaitPopAllContext(ctx)
	if ok {
		fmt.Println(*elements) // [4 5 6]
	} else {
		fmt.Println("Failed to pop elements within the timeout")
	}

	// Recycle the buffer
	queue.RecycleBuffer(elements)

	// Output:
	// [1 2 3]
	// [4 5 6]
}
