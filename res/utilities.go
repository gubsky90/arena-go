package res

import (
	"math/bits"
	"unsafe"
)

// RoundPow2 returns the smallest power of two that is >= n.
func RoundPow2(n uint64) uint64 {
	return uint64(1) << uint64(bits.UintSize-bits.LeadingZeros64(n-1))
}

// Pow2 returns 2^n.
func Pow2(n uint64) uint64 {
	return 1 << n
}

// Log2 returns log2(n).
func Log2(n uint64) uint64 {
	return uint64(bits.Len64(n) - 1)
}

// AsUnsafePointer converts a pointer to unsafe.Pointer.
// This is a generic helper that eliminates the need for explicit unsafe.Pointer casts.
//
// Example:
//
//	ptr := MakeObject[int](a)
//	unsafePtr := AsUnsafePointer(ptr)
func AsUnsafePointer[T any](ptr *T) unsafe.Pointer {
	return unsafe.Pointer(ptr)
}

// AsUnsafePointerSlice converts a slice to unsafe.Pointer pointing to its underlying array.
// Returns nil for empty slices.
//
// Example:
//
//	slice := MakeSlice[int](a, 10, 20)
//	slicePtr := AsUnsafePointerSlice(slice)
func AsUnsafePointerSlice[T any](slice []T) unsafe.Pointer {
	if len(slice) == 0 {
		return nil
	}
	return unsafe.Pointer(unsafe.SliceData(slice))
}

// AsUnsafePointerString converts a string to unsafe.Pointer pointing to its underlying data.
// Returns nil for empty strings.
//
// Example:
//
//	str := "hello"
//	strPtr := AsUnsafePointerString(str)
func AsUnsafePointerString(s string) unsafe.Pointer {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Pointer(unsafe.StringData(s))
}

// UnsafeBytes converts a string to a byte slice without copying (unsafe).
func UnsafeBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// UnsafeString converts a byte slice to a string without copying (unsafe).
func UnsafeString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}

// Ptr converts a pointer to unsafe.Pointer.
// This is a generic helper that eliminates the need for explicit unsafe.Pointer casts.
func Ptr[T any](ptr *T) unsafe.Pointer {
	return unsafe.Pointer(ptr)
}

// SlicePtr converts a slice to unsafe.Pointer pointing to its underlying array.
// Returns nil for empty slices.
func SlicePtr[T any](slice []T) unsafe.Pointer {
	if len(slice) == 0 {
		return nil
	}
	return unsafe.Pointer(unsafe.SliceData(slice))
}

// StringPtr converts a string to unsafe.Pointer pointing to its underlying data.
// Returns nil for empty strings.
func StringPtr(s string) unsafe.Pointer {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Pointer(unsafe.StringData(s))
}
