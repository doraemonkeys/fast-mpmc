package main

import (
	"context"
	"fmt"
	"github.com/doraemonkeys/fast-mpmc"
	"time"
)

func main() {
	// Create a FastMpmc queue with a minimum capacity of 10
	queue := mpmc.NewFastMpmc[int](10)

	// Start a producer goroutine
	go func() {
		for i := 0; i < 20; i++ {
			queue.Push(i)
			fmt.Printf("Pushed: %d\n", i)
			time.Sleep(100 * time.Millisecond) // 模拟一些延迟
		}
	}()

	// Start a consumer goroutine
	go func() {
		for {
			elements := queue.WaitPopAll()
			if elements != nil {
				for _, elem := range *elements {
					fmt.Printf("Popped: %d\n", elem)
				}
				queue.RecycleBuffer(elements)
			}
		}
	}()

	// Use WaitPopAllContext method with context
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for {
		elements, ok := queue.WaitPopAllContext(ctx)
		if !ok {
			fmt.Println("Context done, exiting...")
			break
		}
		if elements != nil {
			for _, elem := range *elements {
				fmt.Printf("Popped with context: %d\n", elem)
			}
			queue.RecycleBuffer(elements)
		}
	}

	// Use TryPopAll method
	for {
		elements, ok := queue.TryPopAll()
		if !ok {
			fmt.Println("No more elements to pop, exiting...")
			break
		}
		if elements != nil {
			for _, elem := range *elements {
				fmt.Printf("Popped with TryPopAll: %d\n", elem)
			}
			queue.RecycleBuffer(elements)
		}
	}

	// Print queue length and capacity
	fmt.Printf("Queue length: %d\n", queue.Len())
	fmt.Printf("Queue capacity: %d\n", queue.Cap())
}
