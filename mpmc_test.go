package mpmc

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"math/rand"
)

func TestFastMpmc_SwapBuffer(t *testing.T) {
	tests := []struct {
		name           string
		initialBuffer  []int
		newBuffer      []int
		expectedRet    []int
		expectedSignal bool
	}{
		{
			name:           "Empty buffer to non-empty buffer",
			initialBuffer:  []int{},
			newBuffer:      []int{1, 2, 3},
			expectedRet:    []int{},
			expectedSignal: true,
		},
		{
			name:           "Non-empty buffer to empty buffer",
			initialBuffer:  []int{1, 2, 3},
			newBuffer:      []int{},
			expectedRet:    []int{1, 2, 3},
			expectedSignal: false,
		},
		{
			name:           "Non-empty buffer to non-empty buffer",
			initialBuffer:  []int{1, 2, 3},
			newBuffer:      []int{4, 5, 6},
			expectedRet:    []int{1, 2, 3},
			expectedSignal: false,
		},
		{
			name:           "Empty buffer to empty buffer",
			initialBuffer:  []int{},
			newBuffer:      []int{},
			expectedRet:    []int{},
			expectedSignal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mpmc := NewFastMpmc[int](5)

			// Set initial buffer
			mpmc.buffer = &tt.initialBuffer

			// Prepare new buffer
			newBuffer := tt.newBuffer

			// Call SwapBuffer
			ret := mpmc.SwapBuffer(&newBuffer)

			// Check returned buffer
			if !sliceEqual(*ret, tt.expectedRet) {
				t.Errorf("SwapBuffer() returned %v, want %v", *ret, tt.expectedRet)
			}

			// Check if the new buffer is now in place
			if !sliceEqual(*mpmc.buffer, tt.newBuffer) {
				t.Errorf("After SwapBuffer(), buffer is %v, want %v", *mpmc.buffer, tt.newBuffer)
			}

			// Check if signal was sent when expected
			if tt.expectedSignal {
				select {
				case <-mpmc.popAllCondChan:
					// Signal was sent as expected
				case <-time.After(time.Millisecond * 10):
					t.Error("Expected signal, but none received")
				}
			} else {
				select {
				case <-mpmc.popAllCondChan:
					t.Error("Unexpected signal received")
				case <-time.After(time.Millisecond * 10):
					// No signal, as expected
				}
			}
		})
	}
}

// Helper function to compare slices
func sliceEqual(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func TestFastMpmc_PushSlice(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *FastMpmc[int]
		input    []int
		expected []int
	}{
		{
			name: "Push to empty queue",
			setup: func() *FastMpmc[int] {
				return NewFastMpmc[int](5)
			},
			input:    []int{1, 2, 3},
			expected: []int{1, 2, 3},
		},
		{
			name: "Push to non-empty queue",
			setup: func() *FastMpmc[int] {
				q := NewFastMpmc[int](5)
				q.Push(1, 2)
				return q
			},
			input:    []int{3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
		{
			name: "Push empty slice",
			setup: func() *FastMpmc[int] {
				return NewFastMpmc[int](5)
			},
			input:    []int{},
			expected: []int{},
		},
		{
			name: "Push slice larger than initial capacity",
			setup: func() *FastMpmc[int] {
				return NewFastMpmc[int](3)
			},
			input:    []int{1, 2, 3, 4, 5},
			expected: []int{1, 2, 3, 4, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.setup()
			q.PushSlice(tt.input)

			// Use WaitSwapBuffer to get the contents of the queue
			result := q.SwapBuffer(&[]int{})
			equaqlSlice(t, tt.expected, *result)

			if q.Len() != 0 {
				t.Fatalf("Expected queue to be empty, but it has %d elements", q.Len())
			}
		})
	}
}

func TestFastMpmc_PushSlice_Concurrency(t *testing.T) {
	q := NewFastMpmc[int](100)
	numGoroutines := 10
	itemsPerGoroutine := 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(start int) {
			defer wg.Done()
			slice := make([]int, itemsPerGoroutine)
			for j := 0; j < itemsPerGoroutine; j++ {
				slice[j] = start + j
			}
			q.PushSlice(slice)
		}(i * itemsPerGoroutine)
	}

	wg.Wait()

	// Use WaitSwapBuffer to get the contents of the queue
	result := q.WaitSwapBuffer(&[]int{})

	equal(t, numGoroutines*itemsPerGoroutine, len(*result))

	// Check if all numbers are present
	countMap := make(map[int]bool)
	for _, num := range *result {
		countMap[num] = true
	}
	equal(t, numGoroutines*itemsPerGoroutine, len(countMap))
}

func equaqlSlice[T comparable](t *testing.T, a, b []T) {
	if len(a) != len(b) {
		t.Fatalf("Expected %v, got %v", a, b)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("Expected %v, got %v", a, b)
		}
	}
}

func equal[T comparable](t *testing.T, a, b T) {
	if a != b {
		t.Fatalf("Expected %v, got %v", a, b)
	}
}

func TestWaitPopAll(t *testing.T) {
	t.Run("Non-empty queue", func(t *testing.T) {
		q := NewFastMpmc[int](5)
		q.Push(1, 2, 3)

		result := q.WaitPopAll()
		if len(*result) != 3 || (*result)[0] != 1 || (*result)[1] != 2 || (*result)[2] != 3 {
			t.Errorf("Expected [1, 2, 3], got %v", *result)
		}
	})

	t.Run("Initially empty, then filled queue", func(t *testing.T) {
		q := NewFastMpmc[int](5)

		go func() {
			time.Sleep(100 * time.Millisecond)
			q.Push(4, 5, 6)
		}()

		result := q.WaitPopAll()
		if len(*result) != 3 || (*result)[0] != 4 || (*result)[1] != 5 || (*result)[2] != 6 {
			t.Errorf("Expected [4, 5, 6], got %v", *result)
		}
	})
}

func TestWaitPopAllContext(t *testing.T) {
	t.Run("Non-empty queue", func(t *testing.T) {
		q := NewFastMpmc[int](5)
		q.Push(1, 2, 3)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		result, ok := q.WaitPopAllContext(ctx)
		if !ok {
			t.Error("Expected success, got failure")
		}
		if len(*result) != 3 || (*result)[0] != 1 || (*result)[1] != 2 || (*result)[2] != 3 {
			t.Errorf("Expected [1, 2, 3], got %v", *result)
		}
	})

	t.Run("Context timeout", func(t *testing.T) {
		q := NewFastMpmc[int](5)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		result, ok := q.WaitPopAllContext(ctx)
		if ok || result != nil {
			t.Errorf("Expected failure and nil result, got success or non-nil result")
		}
	})

	t.Run("Initially empty, then filled queue", func(t *testing.T) {
		q := NewFastMpmc[int](5)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		go func() {
			time.Sleep(100 * time.Millisecond)
			q.Push(4, 5, 6)
		}()

		result, ok := q.WaitPopAllContext(ctx)
		if !ok {
			t.Error("Expected success, got failure")
		}
		if len(*result) != 3 || (*result)[0] != 4 || (*result)[1] != 5 || (*result)[2] != 6 {
			t.Errorf("Expected [4, 5, 6], got %v", *result)
		}
	})
}

func TestTryPopAll(t *testing.T) {
	t.Run("Non-empty queue", func(t *testing.T) {
		q := NewFastMpmc[int](5)
		q.Push(1, 2, 3)

		result, ok := q.TryPopAll()
		if !ok {
			t.Error("Expected success, got failure")
		}
		if len(*result) != 3 || (*result)[0] != 1 || (*result)[1] != 2 || (*result)[2] != 3 {
			t.Errorf("Expected [1, 2, 3], got %v", *result)
		}
	})

	t.Run("Empty queue", func(t *testing.T) {
		q := NewFastMpmc[int](5)

		result, ok := q.TryPopAll()
		if ok || result != nil {
			t.Errorf("Expected failure and nil result, got success or non-nil result")
		}
	})
}

func TestFastMpmc_Len(t *testing.T) {
	q := NewFastMpmc[int](5)

	if q.Len() != 0 {
		t.Errorf("Expected length 0, got %d", q.Len())
	}

	q.Push(1, 2, 3)
	if q.Len() != 3 {
		t.Errorf("Expected length 3, got %d", q.Len())
	}
}

func TestFastMpmc_Cap(t *testing.T) {
	q := NewFastMpmc[int](5)

	if q.Cap() != 5 {
		t.Errorf("Expected capacity 5, got %d", q.Cap())
	}

	q.Push(1, 2, 3, 4, 5, 6)
	if q.Cap() <= 5 {
		t.Errorf("Expected capacity > 5, got %d", q.Cap())
	}
}

func TestFastMpmc_IsEmpty(t *testing.T) {
	q := NewFastMpmc[int](5)

	if !q.IsEmpty() {
		t.Error("Expected queue to be empty")
	}

	q.Push(1)
	if q.IsEmpty() {
		t.Error("Expected queue to be non-empty")
	}

	_, _ = q.TryPopAll()
	if !q.IsEmpty() {
		t.Error("Expected queue to be empty after popping all elements")
	}
}

func TestFastMpmc_LenNoLock(t *testing.T) {
	q := NewFastMpmc[int](5)

	q.Push(1, 2, 3)
	q.bufferMu.Lock()
	defer q.bufferMu.Unlock()

	if q.Len() != 3 {
		t.Errorf("Expected length 3, got %d", q.Len())
	}
}

func TestFastMpmc_CapNoLock(t *testing.T) {
	q := NewFastMpmc[int](5)

	q.bufferMu.Lock()
	defer q.bufferMu.Unlock()

	if q.Cap() != 5 {
		t.Errorf("Expected capacity 5, got %d", q.Cap())
	}
}

func TestFastMpmc_IsEmptyNoLock(t *testing.T) {
	q := NewFastMpmc[int](5)

	q.bufferMu.Lock()
	defer q.bufferMu.Unlock()

	if !q.IsEmpty() {
		t.Error("Expected queue to be empty")
	}

	*q.buffer = append(*q.buffer, 1)
	if q.IsEmpty() {
		t.Error("Expected queue to be non-empty")
	}
}

func TestConcurrentSwapAndWaitSwap(t *testing.T) {
	t.Parallel()
	const (
		numWorkers     = 10
		numOperations  = 1000
		bufferCapacity = 100
	)

	queue := NewFastMpmc[int](bufferCapacity)
	var wg sync.WaitGroup

	// 启动多个 goroutine 进行并发操作
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				switch rand.Intn(3) {
				case 0:
					// Push 操作
					queue.Push(rand.Int())
				case 1:
					// SwapBuffer 操作
					newBuffer := make([]int, 0, bufferCapacity)
					oldBuffer := queue.SwapBuffer(&newBuffer)
					t.Logf("SwapBuffer: old buffer len %d, new buffer len %d", len(*oldBuffer), len(newBuffer))
				case 2:
					// WaitSwapBuffer 操作
					ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
					defer cancel()
					newBuffer := make([]int, 0, bufferCapacity)
					oldBuffer, ok := queue.WaitSwapBufferContext(ctx, &newBuffer)
					if ok {
						t.Logf("WaitSwapBuffer: old buffer len %d, new buffer len %d", len(*oldBuffer), len(newBuffer))
					} else {
						t.Log("WaitSwapBuffer: timeout")
					}
				}
			}
		}()
	}

	// 等待所有 goroutine 完成
	wg.Wait()
}

func TestEdgeCaseSwapAndWaitSwap(t *testing.T) {
	const bufferCapacity = 100
	queue := NewFastMpmc[int](bufferCapacity)

	// 准备一个非空缓冲区
	queue.Push(1)

	var wg sync.WaitGroup
	wg.Add(2)

	// 同时调用 SwapBuffer 和 WaitSwapBuffer
	go func() {
		defer wg.Done()
		newBuffer := make([]int, 0, bufferCapacity)
		oldBuffer := queue.SwapBuffer(&newBuffer)
		t.Logf("SwapBuffer: old buffer len %d, new buffer len %d", len(*oldBuffer), len(newBuffer))
	}()

	go func() {
		defer wg.Done()
		newBuffer := make([]int, 0, bufferCapacity)
		oldBuffer := queue.WaitSwapBuffer(&newBuffer)
		t.Logf("WaitSwapBuffer: old buffer len %d, new buffer len %d", len(*oldBuffer), len(newBuffer))
	}()

	wg.Wait()
}

func TestConcurrentSwapAndWaitSwap2(t *testing.T) {
	t.Parallel()
	const (
		iterations = 2000
		bufSize    = 10
	)

	queue := NewFastMpmc[int](bufSize)
	var wg sync.WaitGroup

	startCh := make(chan struct{})

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			newBuf := make([]int, 0, bufSize)
			<-startCh
			for j := 0; j < iterations; j++ {
				oldBuf := queue.SwapBuffer(&newBuf)
				if len(*oldBuf) > 0 {
					newBuf = (*oldBuf)[:0]
				}
			}
		}()
	}

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			newBuf := make([]int, 0, bufSize)
			<-startCh
			for j := 0; j < iterations; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				oldBuf, ok := queue.WaitSwapBufferContext(ctx, &newBuf)
				cancel()
				if ok && len(*oldBuf) > 0 {
					newBuf = (*oldBuf)[:0]
				}
			}
		}()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		<-startCh
		for i := 0; i < iterations*10; i++ {
			queue.Push(i)
			time.Sleep(time.Microsecond)
		}
	}()

	close(startCh)

	wg.Wait()

	t.Log("Test completed without deadlocks or panics")
}

func TestTryPopAll2(t *testing.T) {
	t.Run("Race condition: non-empty to empty buffer", func(t *testing.T) {
		q := NewFastMpmc[int](10)
		q.Push(1, 2, 3)

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			time.Sleep(10 * time.Millisecond) // Delay to ensure SwapBuffer runs first
			result, ok := q.TryPopAll()
			if !ok && result != nil {
				t.Error("Expected TryPopAll to fail")
			}
		}()

		go func() {
			defer wg.Done()
			newBuffer := make([]int, 0, 10)
			q.SwapBuffer(&newBuffer)
		}()

		wg.Wait()

		// Verify the queue is now empty
		result, ok := q.TryPopAll()
		if ok {
			t.Error("Expected TryPopAll to fail")
		}
		if result != nil {
			t.Error("Expected nil result")
		}
	})
}

func TestTryPopAllConcurrent(t *testing.T) {
	t.Parallel()

	q := NewFastMpmc[int](10)
	var wg sync.WaitGroup
	iterations := 3000

	// 测试正常情况
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			q.Push(i)
			time.Sleep(time.Microsecond)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			elements, ok := q.TryPopAll()
			if ok {
				q.RecycleBuffer(elements)
			}
			time.Sleep(time.Microsecond)
		}
	}()

	// 测试边缘情况: SwapBuffer 与 TryPopAll 并发
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			newBuffer := make([]int, 0, q.bufMinCap)
			oldBuffer := q.SwapBuffer(&newBuffer)
			q.RecycleBuffer(oldBuffer)
			time.Sleep(time.Microsecond)
		}
	}()

	// 额外的 goroutine 来增加并发性和竞争条件的可能性
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			if rand.Float32() < 0.5 {
				q.Push(i)
			} else {
				elements, ok := q.TryPopAll()
				if ok {
					q.RecycleBuffer(elements)
				}
			}
			time.Sleep(time.Microsecond)
		}
	}()

	// 专门测试 case <-b.popAllCondChan return nil, false 的情况
	nilFalseCount := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations*10; i++ { // 增加迭代次数以提高捕获小概率事件的机会
			select {
			case <-q.popAllCondChan:
				elements, ok := q.TryPopAll()
				if !ok && elements == nil {
					nilFalseCount++
				}
				if ok && elements != nil {
					q.RecycleBuffer(elements)
				}
			default:
				// Do nothing
			}
			time.Sleep(time.Microsecond)
		}
	}()

	wg.Wait()

	if nilFalseCount == 0 {
		t.Errorf("Failed to capture the case where TryPopAll returns (nil, false) after receiving a signal")
	} else {
		t.Logf("Captured %d instances of TryPopAll returning (nil, false) after receiving a signal", nilFalseCount)
	}
}

func TestWaitPopAll3(t *testing.T) {

	t.Run("Panic case", func(t *testing.T) {
		q := NewFastMpmc[int](5)

		// 使用一个 channel 来同步 goroutine
		done := make(chan bool)

		go func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
				done <- true
			}()

			// 关闭 channel 以触发 panic
			close(q.popAllCondChan)
			q.WaitPopAll()
		}()

		select {
		case <-done:
			// 测试通过
		case <-time.After(time.Second):
			t.Fatal("Test timed out")
		}
	})
}

func TestWaitSwapBuffer(t *testing.T) {
	t.Run("Normal case", func(t *testing.T) {
		q := NewFastMpmc[int](5)
		q.Push(1, 2, 3)

		newBuffer := make([]int, 0, 5)
		result := q.WaitSwapBuffer(&newBuffer)

		if len(*result) != 3 || (*result)[0] != 1 || (*result)[2] != 3 {
			t.Errorf("Unexpected result: %v", *result)
		}
	})

	t.Run("Panic case", func(t *testing.T) {
		q := NewFastMpmc[int](5)

		done := make(chan bool)

		go func() {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("The code did not panic")
				}
				done <- true
			}()

			newBuffer := make([]int, 0, 5)
			// 关闭 channel 以触发 panic
			close(q.popAllCondChan)
			q.WaitSwapBuffer(&newBuffer)
		}()

		select {
		case <-done:
			// 测试通过
		case <-time.After(time.Second):
			t.Fatal("Test timed out")
		}
	})
}

func TestNewFastMpmc(t *testing.T) {
	mq := NewFastMpmc[int](10)
	if mq == nil {
		t.Fatal("Expected non-nil SimpleMQ")
	}
	if mq.bufMinCap != 10 {
		t.Fatalf("Expected bufMinCap to be 10, got %d", mq.bufMinCap)
	}
	if mq.buffer == nil {
		t.Fatal("Expected non-nil buffer")
	}
	if len(*mq.buffer) != 0 {
		t.Fatalf("Expected buffer length to be 0, got %d", len(*mq.buffer))
	}
}

func TestPushAndPopAll(t *testing.T) {
	mq := NewFastMpmc[int](10)

	// Test Push
	mq.Push(1, 2, 3)
	if len(*mq.buffer) != 3 {
		t.Fatalf("Expected buffer length to be 3, got %d", len(*mq.buffer))
	}

	// Test popAll
	buffer := mq.popAll()
	if len(*buffer) != 3 {
		t.Fatalf("Expected popped buffer length to be 3, got %d", len(*buffer))
	}
	if len(*mq.buffer) != 0 {
		t.Fatalf("Expected buffer length to be 0 after popAll, got %d", len(*mq.buffer))
	}
}

func TestPushSlice(t *testing.T) {
	mq := NewFastMpmc[int](10)

	// Test PushSlice
	mq.PushSlice([]int{4, 5, 6})
	if len(*mq.buffer) != 3 {
		t.Fatalf("Expected buffer length to be 3, got %d", len(*mq.buffer))
	}

	// Test popAll
	buffer := mq.popAll()
	if len(*buffer) != 3 {
		t.Fatalf("Expected popped buffer length to be 3, got %d", len(*buffer))
	}
	if len(*mq.buffer) != 0 {
		t.Fatalf("Expected buffer length to be 0 after popAll, got %d", len(*mq.buffer))
	}
}

func TestSwapBuffer(t *testing.T) {
	mq := NewFastMpmc[int](10)

	// Push some elements
	mq.Push(7, 8, 9)
	if len(*mq.buffer) != 3 {
		t.Fatalf("Expected buffer length to be 3, got %d", len(*mq.buffer))
	}

	// Create a new buffer to swap
	newBuffer := make([]int, 0, 10)
	newBuffer = append(newBuffer, 10, 11, 12)

	// Swap buffer
	oldBuffer := mq.SwapBuffer(&newBuffer)
	if len(*oldBuffer) != 3 {
		t.Fatalf("Expected old buffer length to be 3, got %d", len(*oldBuffer))
	}
	if len(*mq.buffer) != 3 {
		t.Fatalf("Expected new buffer length to be 3, got %d", len(*mq.buffer))
	}
}

func TestEmptyPopAll(t *testing.T) {
	mq := NewFastMpmc[int](10)

	// Test popAll on empty buffer
	buffer := mq.popAll()
	if buffer != nil {
		t.Fatalf("Expected nil buffer, got %v", buffer)
	}
}

func TestEnableSignal(t *testing.T) {
	mq := NewFastMpmc[int](10)

	// Test signal enabling
	mq.Push(13, 14, 15)
	select {
	case <-mq.popAllCondChan:
		// Expected to receive a signal
	default:
		t.Fatal("Expected to receive a signal, but did not")
	}

	// Test signal disabling
	mq.popAll()
	select {
	case <-mq.popAllCondChan:
		t.Fatal("Did not expect to receive a signal, but did")
	default:
		// Expected to not receive a signal
	}
}

func TestSimpleMQ(t *testing.T) {
	mq := NewFastMpmc[int](10)

	var wg sync.WaitGroup
	numProducers := 5
	numConsumers := 5
	numMessages := 100

	// Producer goroutines
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(producerID int) {
			defer wg.Done()
			for j := 0; j < numMessages; j++ {
				mq.Push(producerID*numMessages + j)
			}
		}(i)
	}

	// Consumer goroutines
	for i := 0; i < numConsumers; i++ {
		wg.Add(1)
		go func(consumerID int) {
			defer wg.Done()
			for {
				select {
				case <-mq.popAllCondChan:
					buffer := mq.popAll()
					if buffer == nil {
						continue
					}
					for _, value := range *buffer {
						t.Logf("Consumer %d received: %d", consumerID, value)
					}
					mq.RecycleBuffer(buffer)
				case <-time.After(1 * time.Second):
					return
				}
			}
		}(i)
	}

	wg.Wait()
}

func TestSwapBuffer2(t *testing.T) {
	mq := NewFastMpmc[int](10)

	// Initial buffer with some elements
	initialBuffer := []int{1, 2, 3, 4, 5}
	mq.PushSlice(initialBuffer)

	// New buffer to swap in
	newBuffer := make([]int, 0, 10)
	swappedBuffer := mq.SwapBuffer(&newBuffer)

	// Check if the swapped buffer is the initial buffer
	if len(*swappedBuffer) != len(initialBuffer) {
		t.Errorf("Expected swapped buffer length to be %d, got %d", len(initialBuffer), len(*swappedBuffer))
	}

	// Check if the new buffer is now the current buffer
	if len(*mq.buffer) != 0 {
		t.Errorf("Expected current buffer length to be 0, got %d", len(*mq.buffer))
	}
}

func TestPushSlice2(t *testing.T) {
	mq := NewFastMpmc[int](10)

	// Push some elements
	elements := []int{1, 2, 3, 4, 5}
	mq.PushSlice(elements)

	// Check if the buffer contains the elements
	mq.bufferMu.Lock()
	defer mq.bufferMu.Unlock()
	if len(*mq.buffer) != len(elements) {
		t.Errorf("Expected buffer length to be %d, got %d", len(elements), len(*mq.buffer))
	}

	for i, v := range *mq.buffer {
		if v != elements[i] {
			t.Errorf("Expected buffer element at index %d to be %d, got %d", i, elements[i], v)
		}
	}
}

func TestPush(t *testing.T) {
	mq := NewFastMpmc[int](10)

	// Push some elements
	mq.Push(1, 2, 3, 4, 5)

	// Check if the buffer contains the elements
	mq.bufferMu.Lock()
	defer mq.bufferMu.Unlock()
	expected := []int{1, 2, 3, 4, 5}
	if len(*mq.buffer) != len(expected) {
		t.Errorf("Expected buffer length to be %d, got %d", len(expected), len(*mq.buffer))
	}

	for i, v := range *mq.buffer {
		if v != expected[i] {
			t.Errorf("Expected buffer element at index %d to be %d, got %d", i, expected[i], v)
		}
	}
}

func TestWaitSwapBuffer_BasicFunctionality(t *testing.T) {
	queue := NewFastMpmc[int](10)
	newBuffer := make([]int, 0, 10)

	// Push some elements to the queue
	queue.Push(1, 2, 3)

	// Swap buffer and verify
	oldBuffer := queue.WaitSwapBuffer(&newBuffer)
	if len(*oldBuffer) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(*oldBuffer))
	}
	if (*oldBuffer)[0] != 1 || (*oldBuffer)[1] != 2 || (*oldBuffer)[2] != 3 {
		t.Fatalf("unexpected buffer contents: %v", *oldBuffer)
	}
}

func TestWaitSwapBuffer_BlockingBehavior(t *testing.T) {
	queue := NewFastMpmc[int](10)
	newBuffer := make([]int, 0, 10)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		// This should block until an element is pushed
		queue.WaitSwapBuffer(&newBuffer)
	}()

	// Give the goroutine some time to block
	time.Sleep(100 * time.Millisecond)

	// Push an element to unblock the goroutine
	queue.Push(1)

	wg.Wait()
}

func TestWaitSwapBuffer_EmptyBuffer(t *testing.T) {
	queue := NewFastMpmc[int](10)
	newBuffer := make([]int, 0, 10)

	// This should block until an element is pushed
	go func() {
		time.Sleep(100 * time.Millisecond)
		queue.Push(1)
	}()

	oldBuffer := queue.WaitSwapBuffer(&newBuffer)
	if len(*oldBuffer) != 1 {
		t.Fatalf("expected 1 element, got %d", len(*oldBuffer))
	}
	if (*oldBuffer)[0] != 1 {
		t.Fatalf("unexpected buffer contents: %v", *oldBuffer)
	}
}

func TestWaitSwapBufferContext_Cancel(t *testing.T) {
	queue := NewFastMpmc[int](10)
	newBuffer := make([]int, 0, 10)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	oldBuffer, ok := queue.WaitSwapBufferContext(ctx, &newBuffer)
	if ok {
		t.Fatal("expected context to cancel, but it didn't")
	}
	if oldBuffer != nil {
		t.Fatalf("expected nil buffer, got %v", oldBuffer)
	}
}

func TestWaitSwapBufferConcurrent(t *testing.T) {
	t.Parallel()
	const (
		pushGoRoutines  = 10
		swapGoRoutines  = 5
		itemsPerRoutine = 1000
		bufferMinCap    = 100
	)

	queue := NewFastMpmc[int](bufferMinCap)
	var wg sync.WaitGroup

	// Start push goroutines
	for i := 0; i < pushGoRoutines; i++ {
		wg.Add(1)
		go func(routineID int) {
			defer wg.Done()
			for j := 0; j < itemsPerRoutine; j++ {
				queue.Push(routineID*itemsPerRoutine + j)
			}
		}(i)
	}

	// Start swap goroutines
	receivedItems := make([]int, 0, pushGoRoutines*itemsPerRoutine)
	var receivedMu sync.Mutex

	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()
	for i := 0; i < swapGoRoutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			newBuffer := queue.bufferPool.Get().(*[]int)
			for {

				swapped, ok := queue.WaitSwapBufferContext(ctx, newBuffer)
				if !ok {
					break
				}
				if len(*swapped) == 0 {
					t.Errorf("Expected swapped buffer to contain items, got 0")
					return
				}

				receivedMu.Lock()
				receivedItems = append(receivedItems, *swapped...)
				receivedMu.Unlock()

				if len(receivedItems) >= pushGoRoutines*itemsPerRoutine {
					cancle()
					break
				}
				*swapped = (*swapped)[:0]
				newBuffer = swapped
			}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Give some time for the last swap operations to complete
	time.Sleep(100 * time.Millisecond)

	// Verify results
	if len(receivedItems) != pushGoRoutines*itemsPerRoutine {
		t.Errorf("Expected %d items, got %d", pushGoRoutines*itemsPerRoutine, len(receivedItems))
	}

	// Create a map to check for duplicates and missing items
	itemMap := make(map[int]bool)
	for _, item := range receivedItems {
		if itemMap[item] {
			t.Errorf("Duplicate item found: %d", item)
		}
		itemMap[item] = true
	}

	for i := 0; i < pushGoRoutines*itemsPerRoutine; i++ {
		if !itemMap[i] {
			t.Errorf("Missing item: %d", i)
		}
	}
}

func BenchmarkSPSC(b *testing.B) {
	benchCases := []struct {
		name      string
		batchSize int
	}{
		{"Small", 1},
		{"Medium", 100},
		{"Large", 10000},
	}

	for _, bc := range benchCases {
		b.Run("Fast MPMC_"+bc.name, func(b *testing.B) {
			benchmarkSPSCFastMPMC(b, bc.batchSize)
		})
		b.Run("Channel_"+bc.name, func(b *testing.B) {
			benchmarkSPSCChannel(b, bc.batchSize)
		})
	}
}

func benchmarkSPSCFastMPMC(b *testing.B, _ int) {
	mq := NewFastMpmc[int](benchmarkMqInitialSize)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			mq.Push(i)
		}
	}()

	count := 0
	for count < b.N {
		items, _ := mq.WaitPopAllContext(ctx)
		count += len(*items)
		mq.RecycleBuffer(items)
	}
}

func benchmarkSPSCChannel(b *testing.B, _ int) {
	ch := make(chan int, benchmarkMqInitialSize*2)

	b.ResetTimer()
	go func() {
		for i := 0; i < b.N; i++ {
			ch <- i
		}
		close(ch)
	}()

	count := 0
	for range ch {
		count++
	}
}

func BenchmarkMPMC(b *testing.B) {
	benchCases := []struct {
		name      string
		batchSize int
		producers int
		consumers int
	}{
		{"Small", 1, 2, 2},
		{"Medium", 100, 4, 4},
		{"Large", 1000, 8, 8},
		{"Huge", 10000, 16, 2},
	}

	for _, bc := range benchCases {
		b.Run("Fast MPMC_"+bc.name, func(b *testing.B) {
			benchmarkMPMCSimpleMQ(b, bc.batchSize, bc.producers, bc.consumers)
		})
		b.Run("Channel_"+bc.name, func(b *testing.B) {
			benchmarkMPMCChannel(b, bc.batchSize, bc.producers, bc.consumers)
		})
	}
}

const benchmarkMqInitialSize = 20

func benchmarkMPMCSimpleMQ(b *testing.B, _, producers, consumers int) {
	mq := NewFastMpmc[int](benchmarkMqInitialSize)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	itemsPerProducer := b.N / producers

	b.ResetTimer()

	// Producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		if i == producers-1 {
			itemsPerProducer += b.N % producers
		}
		go func(nums int) {
			defer wg.Done()
			for j := 0; j < nums; j++ {
				mq.Push(j)
			}
		}(itemsPerProducer)
	}

	var wg2 sync.WaitGroup
	// Consumers
	consumedItems := make([]int, consumers)
	consumedIDs := make([]map[int64]bool, consumers)
	for i := 0; i < consumers; i++ {
		consumedIDs[i] = make(map[int64]bool)
	}
	for i := 0; i < consumers; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			var temp = make([]int, 0, benchmarkMqInitialSize)
			buffer := &temp
			for {
				*buffer = (*buffer)[:0]
				old, ok := mq.WaitSwapBufferContext(ctx, buffer)
				if !ok && mq.Len() == 0 {
					return
				}
				if old != nil {
					consumedItems[id] += len(*old)
					for i := 0; i < len(*old); i++ {
						consumedIDs[id][int64((*old)[i])] = true
					}
					buffer = old
				}
			}
		}(i)
	}

	wg.Wait()
	cancel() // Stop consumers
	wg2.Wait()

	totalConsumed := 0
	for _, count := range consumedItems {
		totalConsumed += count
	}
	if totalConsumed != b.N {
		b.Fatalf("Expected to consume %d items, but consumed %d, unique items: %v", b.N, totalConsumed, consumedIDs)
	}
}

func benchmarkMPMCChannel(b *testing.B, _, producers, consumers int) {
	ch := make(chan int, benchmarkMqInitialSize*2)

	var wg sync.WaitGroup
	itemsPerProducer := b.N / producers

	b.ResetTimer()

	// Producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		if i == producers-1 {
			itemsPerProducer += b.N % producers
		}
		go func(nums int) {
			defer wg.Done()
			for j := 0; j < nums; j++ {
				ch <- j
			}
		}(itemsPerProducer)
	}

	// Closer
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Consumers
	consumedItems := make([]int, consumers)
	consumedIDs := make([]map[int64]bool, consumers)
	for i := 0; i < consumers; i++ {
		consumedIDs[i] = make(map[int64]bool)
	}
	var consumerWg sync.WaitGroup
	for i := 0; i < consumers; i++ {
		consumerWg.Add(1)
		go func(id int) {
			defer consumerWg.Done()
			for item := range ch {
				consumedItems[id]++
				consumedIDs[id][int64(item)] = true
			}
		}(i)
	}

	consumerWg.Wait()

	totalConsumed := 0
	for _, count := range consumedItems {
		totalConsumed += count
	}
	if totalConsumed != b.N {
		b.Fatalf("Expected to consume %d items, but consumed %d, unique items: %v", b.N, totalConsumed, consumedIDs)
	}
}

type TestItem struct {
	ID   int64
	Name string
	Data [128]byte
}

func BenchmarkMPMC_BigStruct(b *testing.B) {
	benchCases := []struct {
		name      string
		batchSize int
		producers int
		consumers int
	}{
		{"Small", 1, 2, 2},
		{"Medium", 100, 4, 4},
		{"Large", 1000, 8, 8},
		{"Huge", 10000, 16, 2},
	}

	for _, bc := range benchCases {
		b.Run("Fast MPMC_"+bc.name, func(b *testing.B) {
			benchmarkMPMCSimpleMQ_BigStruct(b, bc.batchSize, bc.producers, bc.consumers)
		})
		b.Run("Channel_"+bc.name, func(b *testing.B) {
			benchmarkMPMCChannel_BigStruct(b, bc.batchSize, bc.producers, bc.consumers)
		})
	}
}

func benchmarkMPMCSimpleMQ_BigStruct(b *testing.B, _, producers, consumers int) {
	mq := NewFastMpmc[TestItem](benchmarkMqInitialSize)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	itemsPerProducer := b.N / producers

	b.ResetTimer()

	// Producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		if i == producers-1 {
			itemsPerProducer += b.N % producers
		}
		go func(nums int) {
			defer wg.Done()
			for j := 0; j < nums; j++ {
				item := TestItem{
					ID:   int64(j),
					Name: fmt.Sprintf("Item %d", j),
				}
				mq.Push(item)
			}
		}(itemsPerProducer)
	}

	var wg2 sync.WaitGroup
	// Consumers
	consumedItems := make([]int, consumers)
	consumedIDs := make([]map[int64]bool, consumers)
	for i := 0; i < consumers; i++ {
		consumedIDs[i] = make(map[int64]bool)
	}
	for i := 0; i < consumers; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			var temp = make([]TestItem, 0, benchmarkMqInitialSize)
			buffer := &temp
			for {
				*buffer = (*buffer)[:0]
				old, ok := mq.WaitSwapBufferContext(ctx, buffer)
				if !ok && mq.Len() == 0 {
					return
				}
				if old != nil {
					consumedItems[id] += len(*old)
					for i := 0; i < len(*old); i++ {
						consumedIDs[id][(*old)[i].ID] = true
					}
					buffer = old
				}
			}
		}(i)
	}

	wg.Wait()
	cancel() // Stop consumers
	wg2.Wait()

	totalConsumed := 0
	for _, count := range consumedItems {
		totalConsumed += count
	}
	if totalConsumed != b.N {
		b.Fatalf("Expected to consume %d items, but consumed %d, unique items: %v", b.N, totalConsumed, consumedIDs)
	}
}

func benchmarkMPMCChannel_BigStruct(b *testing.B, _, producers, consumers int) {
	ch := make(chan TestItem, benchmarkMqInitialSize*2)

	var wg sync.WaitGroup
	itemsPerProducer := b.N / producers

	b.ResetTimer()

	// Producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		if i == producers-1 {
			itemsPerProducer += b.N % producers
		}
		go func(nums int) {
			defer wg.Done()
			for j := 0; j < nums; j++ {
				item := TestItem{
					ID:   int64(j),
					Name: fmt.Sprintf("Item %d", j),
				}
				ch <- item
			}
		}(itemsPerProducer)
	}

	// Closer
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Consumers
	consumedItems := make([]int, consumers)
	consumedIDs := make([]map[int64]bool, consumers)
	for i := 0; i < consumers; i++ {
		consumedIDs[i] = make(map[int64]bool)
	}
	var consumerWg sync.WaitGroup
	for i := 0; i < consumers; i++ {
		consumerWg.Add(1)
		go func(id int) {
			defer consumerWg.Done()
			for item := range ch {
				consumedItems[id]++
				consumedIDs[id][item.ID] = true
			}
		}(i)
	}

	consumerWg.Wait()

	totalConsumed := 0
	for _, count := range consumedItems {
		totalConsumed += count
	}
	if totalConsumed != b.N {
		b.Fatalf("Expected to consume %d items, but consumed %d, unique items: %v", b.N, totalConsumed, consumedIDs)
	}
}

func BenchmarkMPSC(b *testing.B) {
	benchCases := []struct {
		name      string
		batchSize int
		producers int
	}{
		{"Small", 1, 2},
		{"Medium", 100, 4},
		{"Large", 1000, 8},
		{"Huge", 10000, 16},
	}

	for _, bc := range benchCases {
		b.Run("Fast MPSC_"+bc.name, func(b *testing.B) {
			benchmarkMPSCSimpleMQ(b, bc.batchSize, bc.producers)
		})
		b.Run("Channel_"+bc.name, func(b *testing.B) {
			benchmarkMPSCChannel(b, bc.batchSize, bc.producers)
		})
	}
}

func benchmarkMPSCSimpleMQ(b *testing.B, _, producers int) {
	mq := NewFastMpmc[int](benchmarkMqInitialSize)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	itemsPerProducer := b.N / producers

	b.ResetTimer()

	// Producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		if i == producers-1 {
			itemsPerProducer += b.N % producers
		}
		go func(nums int) {
			defer wg.Done()
			for j := 0; j < nums; j++ {
				mq.Push(j)
			}
		}(itemsPerProducer)
	}

	// Single Consumer
	consumed := 0
	itemIDs := make(map[int64]bool)
	go func() {
		wg.Wait()
		cancel() // Stop consumer when all producers are done
	}()

	var temp = make([]int, 0, benchmarkMqInitialSize)
	buffer := &temp
	for {
		*buffer = (*buffer)[:0]
		old, ok := mq.WaitSwapBufferContext(ctx, buffer)
		if !ok && mq.Len() == 0 {
			break
		}
		if old != nil {
			consumed += len(*old)
			for i := 0; i < len(*old); i++ {
				itemIDs[int64((*old)[i])] = true
			}
			buffer = old
		}
	}

	if consumed != b.N {
		b.Fatalf("Expected to consume %d items, but consumed %d, unique items: %d", b.N, consumed, len(itemIDs))
	}
}

func benchmarkMPSCChannel(b *testing.B, _, producers int) {
	ch := make(chan int, benchmarkMqInitialSize*2)

	var wg sync.WaitGroup
	itemsPerProducer := b.N / producers

	b.ResetTimer()

	// Producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		if i == producers-1 {
			itemsPerProducer += b.N % producers
		}
		go func(nums int) {
			defer wg.Done()
			for j := 0; j < nums; j++ {
				ch <- j
			}
		}(itemsPerProducer)
	}

	// Closer
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Single Consumer
	consumed := 0
	itemIDs := make(map[int64]bool)
	for item := range ch {
		consumed++
		itemIDs[int64(item)] = true
	}

	if consumed != b.N {
		b.Fatalf("Expected to consume %d items, but consumed %d, unique items: %d", b.N, consumed, len(itemIDs))
	}
}
func BenchmarkMPSC_BigStruct(b *testing.B) {
	benchCases := []struct {
		name      string
		batchSize int
		producers int
	}{
		{"Small", 1, 2},
		{"Medium", 100, 4},
		{"Large", 1000, 8},
		{"Huge", 10000, 16},
	}

	for _, bc := range benchCases {
		b.Run("Fast MPSC_"+bc.name, func(b *testing.B) {
			benchmarkMPSCSimpleMQ_BigStruct(b, bc.batchSize, bc.producers)
		})
		b.Run("Channel_"+bc.name, func(b *testing.B) {
			benchmarkMPSCChannel_BigStruct(b, bc.batchSize, bc.producers)
		})
	}
}

func benchmarkMPSCSimpleMQ_BigStruct(b *testing.B, _, producers int) {
	mq := NewFastMpmc[TestItem](benchmarkMqInitialSize)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	itemsPerProducer := b.N / producers

	b.ResetTimer()

	// Producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		if i == producers-1 {
			itemsPerProducer += b.N % producers
		}
		go func(nums int) {
			defer wg.Done()
			for j := 0; j < nums; j++ {
				mq.Push(TestItem{
					ID:   int64(j),
					Name: fmt.Sprintf("Item %d", j),
				})
			}
		}(itemsPerProducer)
	}

	// Single Consumer
	consumed := 0
	itemIDs := make(map[int64]bool)
	go func() {
		wg.Wait()
		cancel() // Stop consumer when all producers are done
	}()

	var temp = make([]TestItem, 0, benchmarkMqInitialSize)
	buffer := &temp
	for {
		*buffer = (*buffer)[:0]
		old, ok := mq.WaitSwapBufferContext(ctx, buffer)
		if !ok && mq.Len() == 0 {
			break
		}
		if old != nil {
			consumed += len(*old)
			for i := 0; i < len(*old); i++ {
				itemIDs[(*old)[i].ID] = true
			}
			buffer = old
		}
	}

	if consumed != b.N {
		b.Fatalf("Expected to consume %d items, but consumed %d, unique items: %d", b.N, consumed, len(itemIDs))
	}
}

func benchmarkMPSCChannel_BigStruct(b *testing.B, _, producers int) {
	ch := make(chan TestItem, benchmarkMqInitialSize*2)

	var wg sync.WaitGroup
	itemsPerProducer := b.N / producers

	b.ResetTimer()

	// Producers
	for i := 0; i < producers; i++ {
		wg.Add(1)
		if i == producers-1 {
			itemsPerProducer += b.N % producers
		}
		go func(nums int) {
			defer wg.Done()
			for j := 0; j < nums; j++ {
				ch <- TestItem{
					ID:   int64(j),
					Name: fmt.Sprintf("Item %d", j),
				}
			}
		}(itemsPerProducer)
	}

	// Closer
	go func() {
		wg.Wait()
		close(ch)
	}()

	// Single Consumer
	consumed := 0
	itemIDs := make(map[int64]bool)
	for item := range ch {
		consumed++
		itemIDs[item.ID] = true
	}

	if consumed != b.N {
		b.Fatalf("Expected to consume %d items, but consumed %d, unique items: %d", b.N, consumed, len(itemIDs))
	}
}
