# fast-mpmc
[![Go Reference](https://pkg.go.dev/badge/github.com/doraemonkeys/fast-mpmc.svg)](https://pkg.go.dev/github.com/doraemonkeys/fast-mpmc) [![Go Report Card](https://goreportcard.com/badge/github.com/doraemonkeys/fast-mpmc)](https://goreportcard.com/report/github.com/doraemonkeys/fast-mpmc) [![Coverage Status](https://coveralls.io/repos/github/doraemonkeys/fast-mpmc/badge.svg)](https://coveralls.io/github/doraemonkeys/fast-mpmc)


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
go test -bench Benchmark -benchmem
```

- windows

```bash
goos: windows
goarch: amd64
pkg: github.com/doraemonkeys/fast-mpmc
cpu: AMD Ryzen 7 4800H with Radeon Graphics
BenchmarkSPSC/Fast_MPMC_Small-16                72769170                17.51 ns/op            0 B/op          0 allocs/op
BenchmarkSPSC/Channel_Small-16                  16214835                73.37 ns/op            0 B/op          0 allocs/op
BenchmarkSPSC/Fast_MPMC_Medium-16               68834967                16.71 ns/op            0 B/op          0 allocs/op
BenchmarkSPSC/Channel_Medium-16                 16432368                74.25 ns/op            0 B/op          0 allocs/op
BenchmarkSPSC/Fast_MPMC_Large-16                69326091                16.11 ns/op            0 B/op          0 allocs/op
BenchmarkSPSC/Channel_Large-16                  16778406                71.25 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Small-16                13308343                82.31 ns/op           63 B/op          0 allocs/op
BenchmarkMPMC/Channel_Small-16                   3996104               312.7 ns/op            28 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Medium-16               17923876                60.62 ns/op           33 B/op          0 allocs/op
BenchmarkMPMC/Channel_Medium-16                  2717090               461.1 ns/op            40 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Large-16                17639613                72.51 ns/op           27 B/op          0 allocs/op
BenchmarkMPMC/Channel_Large-16                   1831418               652.5 ns/op            29 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Huge-16                 14638912                87.83 ns/op           10 B/op          0 allocs/op
BenchmarkMPMC/Channel_Huge-16                    2172172               565.0 ns/op             6 B/op          0 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Small-16               4057039               265.6 ns/op            73 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Small-16                 2940981               422.9 ns/op            61 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Medium-16              3931312               308.7 ns/op            54 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Medium-16                1950024               614.7 ns/op            52 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Large-16               3723310               318.1 ns/op            54 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Large-16                 1670511               744.4 ns/op            56 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Huge-16                3670563               315.8 ns/op            33 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Huge-16                  1781544               667.4 ns/op            31 B/op          2 allocs/op
BenchmarkMPSC/Fast_MPSC_Small-16                        10851253               135.1 ns/op            62 B/op          0 allocs/op
BenchmarkMPSC/Channel_Small-16                           6047389               242.4 ns/op            19 B/op          0 allocs/op
BenchmarkMPSC/Fast_MPSC_Medium-16                       17619308               109.7 ns/op            41 B/op          0 allocs/op
BenchmarkMPSC/Channel_Medium-16                          3884724               307.9 ns/op            14 B/op          0 allocs/op
BenchmarkMPSC/Fast_MPSC_Large-16                        20526925                96.46 ns/op           28 B/op          0 allocs/op
BenchmarkMPSC/Channel_Large-16                           3625776               337.2 ns/op             7 B/op          0 allocs/op
BenchmarkMPSC/Fast_MPSC_Huge-16                         17166325                84.54 ns/op           20 B/op          0 allocs/op
BenchmarkMPSC/Channel_Huge-16                            3159135               371.0 ns/op             2 B/op          0 allocs/op
BenchmarkMPSC_BigStruct/Fast_MPSC_Small-16               6375385               178.3 ns/op           161 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Channel_Small-16                 4070065               339.7 ns/op            50 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Fast_MPSC_Medium-16              4285758               276.1 ns/op            62 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Channel_Medium-16                2782287               420.7 ns/op            34 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Fast_MPSC_Large-16               4046803               296.6 ns/op            38 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Channel_Large-16                 2605978               468.6 ns/op            29 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Fast_MPSC_Huge-16                3895831               308.7 ns/op            31 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Channel_Huge-16                  2391663               506.9 ns/op            26 B/op          2 allocs/op
PASS
ok      github.com/doraemonkeys/fast-mpmc       60.197s
```

- Linux

```bash
$ go test -bench Benchmark -benchmem 
goos: linux
goarch: amd64
pkg: github.com/doraemonkeys/fast-mpmc
cpu: Intel(R) Core(TM) i5-7500 CPU @ 3.40GHz
BenchmarkSPSC/Fast_MPMC_Small-4                 59615301                19.65 ns/op            0 B/op          0 allocs/op
BenchmarkSPSC/Channel_Small-4                   20554114                57.23 ns/op            0 B/op          0 allocs/op
BenchmarkSPSC/Fast_MPMC_Medium-4                59662388                19.63 ns/op            0 B/op          0 allocs/op
BenchmarkSPSC/Channel_Medium-4                  20900962                57.08 ns/op            0 B/op          0 allocs/op
BenchmarkSPSC/Fast_MPMC_Large-4                 59236758                19.67 ns/op            0 B/op          0 allocs/op
BenchmarkSPSC/Channel_Large-4                   20903209                57.19 ns/op            0 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Small-4                 17090604                86.90 ns/op           58 B/op          0 allocs/op
BenchmarkMPMC/Channel_Small-4                    7754822               191.8 ns/op            29 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Medium-4                22709574                65.07 ns/op           56 B/op          0 allocs/op
BenchmarkMPMC/Channel_Medium-4                   6956114               195.9 ns/op            31 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Large-4                 21509295                55.73 ns/op           51 B/op          0 allocs/op
BenchmarkMPMC/Channel_Large-4                    7008406               195.7 ns/op            31 B/op          0 allocs/op
BenchmarkMPMC/Fast_MPMC_Huge-4                  13608062                79.07 ns/op            5 B/op          0 allocs/op
BenchmarkMPMC/Channel_Huge-4                     7824103               182.8 ns/op             6 B/op          0 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Small-4                7606657               160.3 ns/op            78 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Small-4                  3924454               328.4 ns/op            53 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Medium-4               7501035               160.4 ns/op            62 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Medium-4                 3825278               327.7 ns/op            52 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Large-4                6374690               164.8 ns/op            71 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Large-4                  3823546               328.2 ns/op            52 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Fast_MPMC_Huge-4                 6615960               179.5 ns/op            35 B/op          2 allocs/op
BenchmarkMPMC_BigStruct/Channel_Huge-4                   3984849               310.0 ns/op            30 B/op          2 allocs/op
BenchmarkMPSC/Fast_MPSC_Small-4                         13322949               116.3 ns/op            53 B/op          0 allocs/op
BenchmarkMPSC/Channel_Small-4                            8814874               166.2 ns/op            24 B/op          0 allocs/op
BenchmarkMPSC/Fast_MPSC_Medium-4                        22931544                99.90 ns/op           37 B/op          0 allocs/op
BenchmarkMPSC/Channel_Medium-4                          10084273               141.4 ns/op            10 B/op          0 allocs/op
BenchmarkMPSC/Fast_MPSC_Large-4                         18490930                81.83 ns/op           13 B/op          0 allocs/op
BenchmarkMPSC/Channel_Large-4                           10736006               146.6 ns/op             5 B/op          0 allocs/op
BenchmarkMPSC/Fast_MPSC_Huge-4                          12715917                97.14 ns/op            2 B/op          0 allocs/op
BenchmarkMPSC/Channel_Huge-4                            10741873               131.3 ns/op             2 B/op          0 allocs/op
BenchmarkMPSC_BigStruct/Fast_MPSC_Small-4                6470430               169.2 ns/op            84 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Channel_Small-4                  4083324               326.9 ns/op            50 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Fast_MPSC_Medium-4               7002794               174.3 ns/op            67 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Channel_Medium-4                 4389544               292.0 ns/op            36 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Fast_MPSC_Large-4                7210298               163.0 ns/op            46 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Channel_Large-4                  4443619               281.0 ns/op            30 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Fast_MPSC_Huge-4                 6590770               180.4 ns/op            30 B/op          2 allocs/op
BenchmarkMPSC_BigStruct/Channel_Huge-4                   4439965               275.8 ns/op            27 B/op          2 allocs/op
PASS
ok      github.com/doraemonkeys/fast-mpmc       58.128s
```

