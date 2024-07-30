# fast-mpmc
[![Go Reference](https://pkg.go.dev/badge/github.com/doraemonkeys/fast-mpmc.svg)](https://pkg.go.dev/github.com/doraemonkeys/fast-mpmc) [![Go Report Card](https://goreportcard.com/badge/github.com/doraemonkeys/fast-mpmc)](https://goreportcard.com/report/github.com/doraemonkeys/fast-mpmc)


This repository contains an implementation of a high-performance Multi-Producer Multi-Consumer (MPMC) queue in Go, named `FastMpmc`. This queue is designed for scenarios where multiple goroutines need to push and pop elements concurrently with minimal contention and high throughput. The implementation leverages the mechanism of buffer swapping to effectively reduce memory allocation and lock contention.


## Installation
To use FastMpmc in your Go project, you can install it using:

```bash
go get github.com/doraemonkeys/fast-mpmc
```

## Usage

### Creating a Queue

```go
queue := NewFastMpmc[int](8) // Creates a new queue with a minimum buffer capacity of 8
```

### Pushing Elements

```go
queue.Push(1, 2, 3) // Pushes multiple elements to the queue
```

### Popping Elements

```go
elements := queue.WaitPopAll() // Waits for elements to be available and pops all elements
```

### Context-Based Popping

```go
ctx, cancel := context.WithTimeout(context.Background(), time.Second)
defer cancel()

elements, ok := queue.WaitPopAllContext(ctx)
if ok {
    // Process elements
} else {
    // Handle timeout or cancellation
}
```

### Buffer Recycling

```go
queue.RecycleBuffer(elements) // Recycles the buffer for reuse
```


## Benchmark

This benchmark compares the performance of the Fast MPMC queue implementation against Go's built-in channels for various data sizes.

| Test Case               | Operations per Second | ns/op  | Allocations per Operation |
|-------------------------|-----------------------|--------|---------------------------|
| Fast MPMC Small         | 40,664,594            | 28.50  | 0                         |
| Channel Small           | 11,229,351            | 101.3  | 0                         |
| Fast MPMC Medium        | 22,075,785            | 55.59  | 0                         |
| Channel Medium          | 5,726,166             | 211.9  | 0                         |
| Fast MPMC Large         | 17,678,229            | 67.47  | 0                         |
| Channel Large           | 2,915,758             | 405.0  | 0                         |
| Fast MPMC Huge          | 15,389,725            | 77.57  | 0                         |
| Channel Huge            | 6,356,145             | 369.6  | 0                         |


Test results show that the Fast MPMC implementation demonstrates significant performance improvements over Go's built-in channels, especially for larger data sizes. 



