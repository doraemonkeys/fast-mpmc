package mpmc

import (
	"context"
	"sync"
	"testing"
	"time"
)

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
					mq.bufferPool.Put(buffer)
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
		b.Run("SimpleMQ_"+bc.name, func(b *testing.B) {
			benchmarkSPSCSimpleMQ(b, bc.batchSize)
		})
		b.Run("Channel_"+bc.name, func(b *testing.B) {
			benchmarkSPSCChannel(b, bc.batchSize)
		})
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
		b.Run("SimpleMQ_"+bc.name, func(b *testing.B) {
			benchmarkMPMCSimpleMQ(b, bc.batchSize, bc.producers, bc.consumers)
		})
		b.Run("Channel_"+bc.name, func(b *testing.B) {
			benchmarkMPMCChannel(b, bc.batchSize, bc.producers, bc.consumers)
		})
	}
}

const benchmarkMqInitialSize = 20

func benchmarkSPSCSimpleMQ(b *testing.B, _ int) {
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
	for i := 0; i < consumers; i++ {
		wg2.Add(1)
		go func(id int) {
			defer wg2.Done()
			for {
				items, ok := mq.WaitPopAllContext(ctx)
				if !ok && mq.Len() == 0 {
					return
				}
				if items != nil {
					consumedItems[id] += len(*items)
					mq.RecycleBuffer(items)
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
		b.Fatalf("Expected to consume %d items, but consumed %d", b.N, totalConsumed)
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
	var consumerWg sync.WaitGroup
	for i := 0; i < consumers; i++ {
		consumerWg.Add(1)
		go func(id int) {
			defer consumerWg.Done()
			for range ch {
				consumedItems[id]++
			}
		}(i)
	}

	consumerWg.Wait()

	totalConsumed := 0
	for _, count := range consumedItems {
		totalConsumed += count
	}
	if totalConsumed != b.N {
		b.Fatalf("Expected to consume %d items, but consumed %d", b.N, totalConsumed)
	}
}
