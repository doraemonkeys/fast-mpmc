package mpmc

import (
	"context"
	"sync"
)

type FastMpmc[T any] struct {
	bufMinCap  int
	buffer     *[]T
	bufferPool *sync.Pool
	// chan waiting for notification before emptying the buffer
	popAllCondChan chan struct{}
	bufferMu       *sync.Mutex
}

// NewFastMpmc creates a new MPMC queue with the specified minimum buffer capacity.
func NewFastMpmc[T any](bufMinCap int) *FastMpmc[T] {
	var buffer = make([]T, 0, bufMinCap)
	return &FastMpmc[T]{
		bufMinCap:      bufMinCap,
		buffer:         &buffer,
		bufferMu:       &sync.Mutex{},
		popAllCondChan: make(chan struct{}, 1),
		bufferPool: &sync.Pool{
			New: func() interface{} {
				var buffer = make([]T, 0, bufMinCap)
				return &buffer
			},
		},
	}
}

// enableSignal give a signal indicating that the buffer is not empty
func (b *FastMpmc[T]) enableSignal() {
	select {
	case b.popAllCondChan <- struct{}{}:
	default:
	}
}

// Push adds one or more elements to the queue.
func (b *FastMpmc[T]) Push(v ...T) {
	b.PushSlice(v)
}

// PushSlice adds a slice of elements to the queue.
func (b *FastMpmc[T]) PushSlice(values []T) {
	if len(values) == 0 {
		return
	}
	b.bufferMu.Lock()
	*b.buffer = append(*b.buffer, values...)
	if len(*b.buffer) == len(values) {
		b.enableSignal()
	}
	b.bufferMu.Unlock()
}

// popAll removes and returns all elements from the queue.
//
// If the queue is empty, it returns nil, this happens only with a low probability when SwapBuffer is called.
//
// The caller should wait the signal before calling this function.
func (b *FastMpmc[T]) popAll() *[]T {
	var newBuffer = b.bufferPool.Get().(*[]T)
	var ret *[]T

	b.bufferMu.Lock()
	defer b.bufferMu.Unlock()

	// low probability
	if len(*b.buffer) == 0 {
		b.bufferPool.Put(newBuffer)
		return nil
	}
	ret = b.buffer
	b.buffer = newBuffer
	return ret
}

// swapWhenNotEmpty swaps the current buffer with the provided new buffer only if the current buffer is not empty.
// If the swap is performed, it returns the old buffer. If the current buffer is empty, it returns nil.
//
// The caller should wait the signal before calling this function.
func (b *FastMpmc[T]) swapWhenNotEmpty(newBuffer *[]T) *[]T {
	var ret *[]T

	b.bufferMu.Lock()
	defer b.bufferMu.Unlock()

	if len(*b.buffer) == 0 {
		// low probability
		return nil
	}

	ret = b.buffer
	b.buffer = newBuffer

	// case 1: non-empty buffer -> non-empty buffer
	if len(*newBuffer) != 0 {
		b.enableSignal()
	}
	// case 2: non-empty buffer -> empty buffer, do nothing

	return ret
}

// WaitSwapBuffer waits for a signal indicating that the buffer is not empty,
// then swaps the current buffer with the provided new buffer and returns the old buffer.
// This method blocks until the buffer is not empty and the swap is performed.
//
// Note: The function will directly replace the old buffer with the new buffer without clearing the new buffer's elements.
func (b *FastMpmc[T]) WaitSwapBuffer(newBuffer *[]T) *[]T {
	for range b.popAllCondChan {
		if elements := b.swapWhenNotEmpty(newBuffer); elements != nil {
			return elements
		}
	}
	panic("unreachable")
}

// WaitSwapBufferContext like WaitSwapBuffer but with a context.
func (b *FastMpmc[T]) WaitSwapBufferContext(ctx context.Context, newBuffer *[]T) (*[]T, bool) {
	for {
		select {
		case <-b.popAllCondChan:
			if elements := b.swapWhenNotEmpty(newBuffer); elements != nil {
				return elements, true
			}
		case <-ctx.Done():
			return nil, false
		}
	}
}

// SwapBuffer swaps the current buffer with a new buffer and returns the old buffer even if the current buffer is empty.
//
// Note: The function will directly replace the old buffer with the new buffer without clearing the new buffer's elements.
func (b *FastMpmc[T]) SwapBuffer(newBuffer *[]T) *[]T {
	var ret *[]T

	b.bufferMu.Lock()
	defer b.bufferMu.Unlock()

	ret = b.buffer
	b.buffer = newBuffer

	if len(*ret) == 0 && len(*newBuffer) != 0 {
		// case 1: empty buffer -> non-empty buffer
		b.enableSignal()
	} else if len(*ret) != 0 && len(*newBuffer) == 0 {
		// case 2: non-empty buffer -> empty buffer
		select {
		case <-b.popAllCondChan:
		default:
			// Uh-oh, there's a poor kid who can't read the buffer.
		}
	}

	// case 3: empty buffer -> empty buffer, do nothing
	// case 4: non-empty buffer -> non-empty buffer, do nothing

	return ret
}

// RecycleBuffer returns the given buffer to the buffer pool for reuse.
// This helps to reduce memory allocations by reusing previously allocated buffers.
func (b *FastMpmc[T]) RecycleBuffer(buffer *[]T) {
	*buffer = (*buffer)[:0]
	b.bufferPool.Put(buffer)
}

// WaitPopAll waits for elements to be available and then removes and returns all elements from the queue.
// After calling this function, you can call RecycleBuffer to recycle the buffer.
func (b *FastMpmc[T]) WaitPopAll() *[]T {
	for range b.popAllCondChan {
		if elements := b.popAll(); elements != nil {
			return elements
		}
	}
	panic("unreachable")
}

// WaitPopAllContext like WaitPopAll but with a context.
func (b *FastMpmc[T]) WaitPopAllContext(ctx context.Context) (*[]T, bool) {
	for {
		select {
		case <-b.popAllCondChan:
			if elements := b.popAll(); elements != nil {
				return elements, true
			}
		case <-ctx.Done():
			return nil, false
		}
	}
}

// TryPopAll tries to remove and return all elements from the queue without blocking. Returns the elements and a boolean indicating success.
func (b *FastMpmc[T]) TryPopAll() (*[]T, bool) {
	select {
	case <-b.popAllCondChan:
		if elements := b.popAll(); elements != nil {
			return elements, true
		}
		return nil, false
	default:
		return nil, false
	}
}

func (b *FastMpmc[T]) Len() int {
	return len(*b.buffer)
}

func (b *FastMpmc[T]) Cap() int {
	return cap(*b.buffer)
}

func (b *FastMpmc[T]) IsEmpty() bool {
	return len(*b.buffer) == 0
}
