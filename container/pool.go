// pool.go — Type-safe, zero-GC object pool using Arena allocator
package container

import (
	"sync"
	"unsafe"

	arena "github.com/thebagchi/arena-go"
)

// Pool[T] is a high-performance, type-safe object pool that allocates from an Arena.
// It reuses freed objects via an internal free list, reducing allocation pressure.
// Perfect for AST nodes, query plans, protobuf messages, and other frequently allocated objects.
//
// Thread Safety:
//   - All operations (Alloc, Free, Reset) are thread-safe
//   - Multiple goroutines can safely allocate and free concurrently
//   - Pool shares the Arena's lifecycle - when Arena is deleted, all Pool memory is freed
type Pool[T any] struct {
	mtx         sync.Mutex
	arena       *arena.Arena
	size        uintptr
	allocations *Vec[*T]
}

// NewPool creates a new object pool for type T that allocates from the given Arena.
// All allocations are 16-byte aligned for optimal performance.
func NewPool[T any](a *arena.Arena) *Pool[T] {
	size := unsafe.Sizeof(*new(T))
	if size == 0 {
		size = 1
	}
	// Align to 16 bytes (cache line friendly)
	const alignment uintptr = 16
	size = (size + alignment - 1) &^ (alignment - 1)

	return &Pool[T]{
		arena:       a,
		size:        size,
		allocations: NewVec[*T](a),
	}
}

// Alloc returns a freshly zeroed T from the pool.
// If the free list is not empty, reuses a previously freed object.
// Otherwise, allocates new memory from the Arena.
func (p *Pool[T]) Alloc() *T {
	p.mtx.Lock()
	defer p.mtx.Unlock()

	if p.allocations.Len() > 0 {
		ptr, _ := p.allocations.Pop() // bool always true since Len() > 0
		// Zero the memory for safety using unsafe slice view
		clear(unsafe.Slice((*byte)(unsafe.Pointer(ptr)), p.size))
		return ptr
	}

	// Allocate from the Arena
	ptr := (*T)(p.arena.Alloc(uint64(p.size), 16))
	return ptr
}

// Free returns an object to the pool's free list for reuse.
// The object must have been allocated by this Pool's Alloc() method.
// It's safe to call Free(nil).
//
// Note: Freed objects are not returned to the Arena - they're held in the
// Pool's free list until Reset() is called or the Arena is deleted.
func (p *Pool[T]) Free(obj *T) {
	if obj == nil {
		return
	}
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.allocations.Push(obj)
}

// Reset clears the free list, making all freed objects eligible for reuse.
// This does not free memory back to the Arena - use Arena.Reset() for that.
// Note: The free list capacity is retained and will not shrink.
func (p *Pool[T]) Reset() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.allocations.Clear()
}

// Len returns the number of objects currently in the free list.
func (p *Pool[T]) Len() int {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	return p.allocations.Len()
}

// Cap returns the current capacity of the free list.
// Useful for monitoring pool memory usage and growth patterns.
func (p *Pool[T]) Cap() int {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	return p.allocations.Cap()
}
