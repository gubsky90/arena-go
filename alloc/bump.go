package alloc

import (
	"sync"
	"unsafe"

	"github.com/thebagchi/arena-go/res"
)

type BumpAllocator struct {
	res *res.Res
	mtx sync.Mutex
}

// NewBumpAllocator creates a new bump allocator.
func NewBumpAllocator(size int) *BumpAllocator {
	r := res.NewRes(size)
	return &BumpAllocator{
		res: r,
	}
}

// Alloc allocates memory of the specified size and alignment.
// It delegates to the Res allocator.
func (b *BumpAllocator) Alloc(size, align uint64) unsafe.Pointer {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	return b.res.Alloc(size, align)
}

// Reset resets the allocator to its initial state, allowing reuse of allocated memory.
// Note: All previously allocated pointers become invalid and should not be used.
func (b *BumpAllocator) Reset() {
	b.mtx.Lock()
	b.res.Reset()
	b.mtx.Unlock()
}

// Delete frees all memory allocated by the allocator.
// Note: All previously allocated pointers become invalid and should not be used.
func (b *BumpAllocator) Delete() {
	b.mtx.Lock()
	b.res.Delete()
	b.mtx.Unlock()
}

// Remove is a no-op for bump allocator, as individual deallocations are not supported.
// Note: This does not invalidate any pointers.
func (b *BumpAllocator) Remove(ptr unsafe.Pointer) {
	// no op for bump allocator
}

// Owns checks if the given pointer belongs to memory managed by this allocator.
func (b *BumpAllocator) Owns(ptr unsafe.Pointer) bool {
	return b.res.Owns(ptr)
}
