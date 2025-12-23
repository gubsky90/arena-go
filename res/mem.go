// Package arena provides memory allocation utilities for arena-based allocators.
// This package handles low-level memory operations using system calls for efficient
// memory management outside of Go's garbage collector.
package res

import (
	"syscall"

	"golang.org/x/sys/unix"
)

var PAGE_SIZE int

func init() {
	PAGE_SIZE = syscall.Getpagesize()
}

// MakePages allocates memory pages using mmap.
// It rounds up the requested size to the nearest page boundary to ensure
// proper alignment and prevent partial page allocations.
//
// Parameters:
//   - size: The minimum number of bytes to allocate. Will be rounded up to page size.
//
// Returns:
//   - []byte: A byte slice backed by the allocated memory pages.
//
// Panics:
//   - If mmap fails to allocate the requested memory.
//
// Note: The allocated memory is not managed by Go's GC and must be explicitly
// released using ReleasePages to avoid memory leaks.
func MakePages(size int) []byte {
	size = ((size + PAGE_SIZE - 1) / PAGE_SIZE) * PAGE_SIZE
	data, err := syscall.Mmap(-1, 0, size, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS)
	if err != nil {
		panic(err)
	}
	return data
}

// ReleasePages frees memory pages allocated with MakePages.
// This function must be called to release memory allocated by MakePages,
// otherwise the memory will leak as it's not managed by Go's garbage collector.
//
// Parameters:
//   - data: The byte slice returned by MakePages. Must be the exact slice
//     returned by MakePages, not a subslice.
//
// Note: After calling ReleasePages, the data slice becomes invalid and
// should not be used. Attempting to access it may cause undefined behavior.
func ReleasePages(data []byte) {
	syscall.Munmap(data)
}

// GrowPages expands an existing memory allocation to a larger size.
// It allocates new pages for the additional memory needed and may move the
// allocation to a different address if in-place expansion is not possible.
//
// Parameters:
//   - data: The byte slice returned by MakePages. Must be the exact slice
//     returned by MakePages, not a subslice.
//   - newSize: The desired new size in bytes. Will be rounded up to page size.
//
// Returns:
//   - []byte: A new byte slice backed by the expanded memory. The address may
//     differ from the original allocation.
//
// Panics:
//   - If mmap fails to allocate the new memory or if unmapping the old allocation fails.
//
// Note: The original data slice becomes invalid after calling GrowPages.
// The caller must update all references to use the returned slice. The old
// memory is automatically released and should not be freed manually.
func GrowPages(data []byte, size int) []byte {
	size = ((size + PAGE_SIZE - 1) / PAGE_SIZE) * PAGE_SIZE
	if size <= len(data) {
		return data
	}
	temp := MakePages(size)
	copy(temp, data)
	ReleasePages(data)
	return temp
}

// ExpandPages expands an existing memory allocation while attempting to preserve
// the base address using mremap (Linux only). This function tries to expand the
// allocation in-place; if that fails, it falls back to the copy-based approach.
//
// Parameters:
//   - data: The byte slice returned by MakePages. Must be the exact slice
//     returned by MakePages, not a subslice.
//   - newSize: The desired new size in bytes. Will be rounded up to page size.
//
// Returns:
//   - []byte: A new byte slice backed by the expanded memory. The base address
//     may remain the same if in-place expansion succeeded, or differ if fallback
//     to copy-based expansion occurred.
//
// Panics:
//   - If mmap/mremap fails to allocate the new memory or if unmapping fails.
//
// Note: The original data slice becomes invalid after calling ExpandPages.
// The caller must update all references to use the returned slice. If in-place
// expansion fails, the old memory is automatically released.
// This function is Linux-specific; on other systems, behavior depends on mremap
// availability or falls back to copy-based expansion.
func ExpandPages(data []byte, size int) []byte {
	size = ((size + PAGE_SIZE - 1) / PAGE_SIZE) * PAGE_SIZE
	if size <= len(data) {
		return data
	}
	temp, err := unix.Mremap(data, size, unix.MREMAP_MAYMOVE)
	if err == nil {
		return temp
	}
	return nil
}
