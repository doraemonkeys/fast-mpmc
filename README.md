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
queue := mpmc.NewFastMpmc[int](8) // Creates a new queue with a minimum buffer capacity of 8
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

### Or Swap the Buffer

```go
elements := queue.WaitSwapBuffer(buffer) // Waits for elements to be available and swaps the buffer
```

### Buffer Recycling

```go
queue.RecycleBuffer(elements) // Recycles the buffer for reuse
```


## Benchmark

This benchmark compares the performance of the Fast MPMC queue implementation against Go's built-in channels for various data sizes.

| Test Case               | Iterations | ns/op  | Allocations per Operation |
|-------------------------|-----------------------|--------|---------------------------|
| Fast MPMC Small         | 40,664,594            | 28.50  | 0                         |
| Channel Small           | 11,229,351            | 101.3  | 0                         |
| Fast MPMC Medium        | 22,075,785            | 55.59  | 0                         |
| Channel Medium          | 5,726,166             | 211.9  | 0                         |
| Fast MPMC Large         | 17,678,229            | 67.47  | 0                         |
| Channel Large           | 2,915,758             | 405.0  | 0                         |
| Fast MPMC Huge          | 15,389,725            | 77.57  | 0                         |
| Channel Huge            | 6,356,145             | 369.6  | 0                         |


Test results show that the Fast MPMC implementation demonstrates significant performance improvements over Go's built-in channels. 



### Raw output

go version: v1.23.0

```bash
go test -bench BenchmarkMPMC -benchmem 
```

- windows

```bash
goos: windows
goarch: amd64
pkg: github.com/doraemonkeys/fast-mpmc
cpu: AMD Ryzen 7 4800H with Radeon Graphics
BenchmarkMPMC/Fast_MPMC_Small-16                43947350                26.10 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Small-16                  12461265                97.85 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Medium-16               22681822                53.17 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Medium-16                  5591316               211.7 ns/op             0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Large-16                17941268                67.81 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Large-16                   2530562               466.4 ns/op             0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Huge-16                 15446877                78.12 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Huge-16                    3268629               370.0 ns/op             0 B/op          0 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Small-16               5307976               226.0 ns/op            25 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Small-16                 3206842               372.0 ns/op            24 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Medium-16              4165419               295.2 ns/op            25 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Medium-16                2147991               557.6 ns/op            24 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Large-16               3882062               311.4 ns/op            25 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Channel_Large-16                 1596138               756.4 ns/op            23 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Huge-16                3753025               320.6 ns/op            24 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Channel_Huge-16                  2145730               569.1 ns/op            23 B/op          1 allocs/op
PASS
ok      github.com/doraemonkeys/fast-mpmc       25.515s
```

- Linux

```bash
$ go test -bench BenchmarkMPMC -benchmem 
goos: linux
goarch: amd64
pkg: github.com/doraemonkeys/fast-mpmc
cpu: Intel(R) Core(TM) i5-7500 CPU @ 3.40GHz
BenchmarkMPMC/Fast_MPMC_Small-4                 46346176                25.47 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Small-4                   18180032                65.78 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Medium-4                23993593                51.93 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Medium-4                  13948807                90.23 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Large-4                 17105738                72.16 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Large-4                   13289521                89.99 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Huge-4                  14390824                85.36 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Huge-4                    13927892                86.49 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Small-4                7913755               152.2 ns/op            25 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Small-4                  5698142               209.3 ns/op            24 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Medium-4               7223360               159.1 ns/op            27 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Channel_Medium-4                 5243253               229.5 ns/op            23 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Large-4                6507098               160.8 ns/op            32 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Channel_Large-4                  5212800               230.1 ns/op            23 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Huge-4                 6733812               174.6 ns/op            26 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Channel_Huge-4                   5010307               238.0 ns/op            23 B/op          1 allocs/op
PASS
ok      github.com/doraemonkeys/fast-mpmc       23.616s
```

