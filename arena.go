// arena/arena.go
//
// Package arena provides high-performance, zero-GC memory allocators with multiple strategies.
//
// Thread Safety:
//   - All allocators (Bump, Slab, Buddy) are thread-safe and can be used concurrently
//   - Alloc() operations are serialized with mutexes to prevent data races
//   - Reset() and Delete() should NOT be called concurrently with Alloc() or with each other
//   - Multiple Arena instances are completely independent and require no synchronization
//
// Memory Model:
//   - All memory is allocated via mmap and lives outside Go's garbage collector
//   - Memory is never returned to the OS until Delete() is called
//   - Reset() clears allocations but retains underlying memory pages
//
// Allocator Strategies:
//   - BUMP: Fastest, best for batch allocations or when arena is reset frequently
//   - SLAB: Best for fixed-size objects with high allocation/free turnover
//   - BUDDY: Most flexible, good for varied-size allocations with power-of-2 sizes
package arena

import (
	"unsafe"
)

// ---------------------------------------------------------------
// Public API – one arena for all types
// ---------------------------------------------------------------

// Arena is the beautiful multi-type facade.
// Thread-safe: Multiple goroutines can safely call Alloc concurrently.
// The underlying allocator handles synchronization internally.
type Arena struct {
	Allocator
}

// New creates an arena from an Allocator implementation.
func New(alloc Allocator) *Arena {
	return &Arena{Allocator: alloc}
}

func (a *Arena) Reset() {
	a.Allocator.Reset()
}
func (a *Arena) Delete() {
	a.Allocator.Delete()
}

// Owns checks if the given pointer belongs to memory managed by this arena.
// Returns true if the pointer was allocated by this arena and is still valid.
// Returns false for nil pointers or pointers not managed by this arena.
func (a *Arena) Owns(ptr unsafe.Pointer) bool {
	return a.Allocator.Owns(ptr)
}

// ---------------------------------------------------------------
// Internal raw allocators (all support growing)
// ---------------------------------------------------------------

type Allocator interface {
	Alloc(size, align uint64) unsafe.Pointer
	Reset()
	Delete()
	Remove(ptr unsafe.Pointer)
	Owns(ptr unsafe.Pointer) bool
}

// OwnsPtr checks if the given pointer to a value belongs to memory managed by this arena.
// This is a convenience wrapper around Owns that eliminates the need for unsafe.Pointer casts.
func OwnsPtr[T any](a *Arena, ptr *T) bool {
	return a.Allocator.Owns(unsafe.Pointer(ptr))
}

// OwnsSlice checks if the underlying array of the given slice belongs to memory managed by this arena.
// Returns false for nil or empty slices.
func OwnsSlice[T any](a *Arena, slice []T) bool {
	if len(slice) == 0 {
		return false
	}
	return a.Owns(unsafe.Pointer(unsafe.SliceData(slice)))
}

// OwnsString checks if the underlying data of the given string belongs to memory managed by this arena.
// Returns false for empty strings.
func OwnsString(a *Arena, s string) bool {
	if len(s) == 0 {
		return false
	}
	return a.Allocator.Owns(unsafe.Pointer(unsafe.StringData(s)))
}
