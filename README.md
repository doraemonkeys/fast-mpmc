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
> go test -bench BenchmarkMPMC -benchmem       
goos: windows
goarch: amd64
pkg: github.com/doraemonkeys/fast-mpmc
cpu: AMD Ryzen 7 4800H with Radeon Graphics
BenchmarkMPMC/Fast_MPMC_Small-16                46308444                26.37 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Small-16                  11415883               100.9 ns/op             0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Medium-16               23198266                52.50 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Medium-16                  5687140               211.6 ns/op             0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Large-16                18620673                64.22 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Large-16                   2574243               457.7 ns/op             0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Huge-16                 15965472                76.85 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Huge-16                    3297562               368.2 ns/op             0 B/op          0 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Small-16               5394382               230.5 ns/op            31 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Small-16                 3516858               346.6 ns/op            24 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Medium-16              4043533               298.9 ns/op            49 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Medium-16                2220840               553.1 ns/op            24 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Large-16               3797011               316.5 ns/op            41 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Large-16                 1600464               746.6 ns/op            23 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Huge-16                3438199               331.7 ns/op            67 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Huge-16                  2134294               565.4 ns/op            23 B/op          1 allocs/op
PASS
ok      github.com/doraemonkeys/fast-mpmc       25.512s
```

- Linux

```bash
$ go test -bench BenchmarkMPMC -benchmem 
goos: linux
goarch: amd64
pkg: github.com/doraemonkeys/fast-mpmc
cpu: Intel(R) Core(TM) i5-7500 CPU @ 3.40GHz
BenchmarkMPMC/Fast_MPMC_Small-4                 41757321                27.87 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Small-4                   18139478                65.86 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Medium-4                22657053                53.89 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Medium-4                  13064905                90.76 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Large-4                 15813302                76.95 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Large-4                   13235518                91.26 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Huge-4                  14104390                88.98 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Channel_Huge-4                    14115940                87.20 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Small-4                7798296               151.9 ns/op            37 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Small-4                  5686855               210.4 ns/op            24 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Medium-4               6704970               167.6 ns/op            44 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Medium-4                 5181726               230.9 ns/op            23 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Large-4                5566794               181.9 ns/op            59 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Large-4                  5194020               231.0 ns/op            23 B/op          1 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Huge-4                 6311290               197.9 ns/op            73 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Huge-4                   4984615               240.8 ns/op            23 B/op          1 allocs/op
PASS
ok      github.com/doraemonkeys/fast-mpmc       23.160s
```

