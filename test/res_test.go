package test

import (
	"testing"
	"unsafe"

	"github.com/thebagchi/arena-go/res"
)

// TestResAllocChunkSwitching verifies Res allocator handles chunk switching correctly
func TestRes_AllocChunkSwitching(t *testing.T) {
	r := res.NewRes(10000)

	ptrs := make([]unsafe.Pointer, 30)
	for i := 0; i < 30; i++ {
		ptrs[i] = r.Alloc(2000, 8)
	}

	// Verify no duplicate addresses
	seenAddrs := make(map[uintptr]bool)
	for i := 0; i < 30; i++ {
		ptrAddr := uintptr(ptrs[i])
		if seenAddrs[ptrAddr] {
			t.Errorf("duplicate address found at allocation %d: %p", i, ptrs[i])
		}
		seenAddrs[ptrAddr] = true
	}
}

// TestResAllocBasic tests basic allocation functionality
func TestRes_AllocBasic(t *testing.T) {
	r := res.NewRes(4096)

	ptr := r.Alloc(64, 8)
	if ptr == nil {
		t.Fatal("Alloc returned nil")
	}

	// Verify pointer is usable
	*((*int)(ptr)) = 42
	if *((*int)(ptr)) != 42 {
		t.Error("failed to write to allocated memory")
	}
}

// TestResAllocAlignment tests alignment constraints
func TestRes_AllocAlignment(t *testing.T) {
	r := res.NewRes(4096)

	tests := []struct {
		name      string
		size      uint64
		alignment uint64
	}{
		{"8-byte alignment", 64, 8},
		{"16-byte alignment", 256, 16},
		{"32-byte alignment", 512, 32},
		{"64-byte alignment", 1024, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ptr := r.Alloc(tt.size, tt.alignment)
			if ptr == nil {
				t.Fatal("Alloc returned nil")
			}

			addr := uintptr(ptr)
			if addr%uintptr(tt.alignment) != 0 {
				t.Errorf("pointer not aligned: %p (alignment %d)", ptr, tt.alignment)
			}
		})
	}
}

// TestResAllocLarge tests allocation of large objects
func TestRes_AllocLarge(t *testing.T) {
	r := res.NewRes(100000)

	ptr := r.Alloc(50000, 8)
	if ptr == nil {
		t.Fatal("Alloc returned nil for large allocation")
	}

	// Verify we can write across the allocation
	data := (*[50000]byte)(ptr)
	data[0] = 1
	data[49999] = 2
	if data[0] != 1 || data[49999] != 2 {
		t.Error("failed to access large allocated memory")
	}
}

// TestResReset tests reset functionality
func TestRes_Reset(t *testing.T) {
	r := res.NewRes(4096)

	ptr1 := r.Alloc(100, 8)
	r.Reset()
	ptr2 := r.Alloc(100, 8)

	// After reset, next allocation might reuse same chunk
	if ptr1 == nil || ptr2 == nil {
		t.Fatal("Alloc returned nil")
	}
}

// TestResCurrent tests current chunk tracking
func TestRes_Current(t *testing.T) {
	r := res.NewRes(4096)

	initial := r.Current()
	if initial != 0 {
		t.Errorf("initial chunk should be 0, got %d", initial)
	}

	// Force chunk switch by allocating large blocks multiple times
	r.Alloc(3000, 8)
	r.Alloc(3000, 8) // This should trigger new chunk

	second := r.Current()
	if second <= 0 {
		t.Errorf("chunk should have switched, still at %d", second)
	}
}

// TestResChunks tests chunks access
func TestRes_Chunks(t *testing.T) {
	r := res.NewRes(4096)

	chunks := r.Chunks()
	if len(chunks) == 0 {
		t.Fatal("no chunks available")
	}

	// Allocate to potentially create more chunks
	r.Alloc(2000, 8)
	r.Alloc(2000, 8)
	r.Alloc(2000, 8)

	chunks = r.Chunks()
	if len(chunks) < 2 {
		t.Errorf("expected at least 2 chunks, got %d", len(chunks))
	}
}

// TestPageNew tests Page creation
func TestPage_New(t *testing.T) {
	p := res.NewPage(4096)
	if p == nil {
		t.Fatal("NewPage returned nil")
	}
	defer p.Delete()
}

// TestPageBase tests Page base access
func TestPage_Base(t *testing.T) {
	p := res.NewPage(4096)
	defer p.Delete()

	base := p.Base()
	if len(base) == 0 {
		t.Fatal("Base returned empty slice")
	}
}

// TestPageDelete tests Page cleanup
func TestPage_Delete(t *testing.T) {
	p := res.NewPage(4096)
	base1 := p.Base()
	if len(base1) == 0 {
		t.Fatal("Base returned empty slice before delete")
	}

	p.Delete()
	base2 := p.Base()
	if len(base2) != 0 {
		t.Errorf("Base should be empty after Delete, got len %d", len(base2))
	}
}

// TestResMultipleAllocations tests multiple allocations with different sizes
func TestRes_MultipleAllocations(t *testing.T) {
	r := res.NewRes(10000)

	sizes := []uint64{64, 128, 256, 512, 1024}
	ptrs := make([]unsafe.Pointer, 0)

	for i, size := range sizes {
		for j := 0; j < 5; j++ {
			ptr := r.Alloc(size, 8)
			if ptr == nil {
				t.Fatalf("allocation %d (size %d) returned nil", i*5+j, size)
			}
			ptrs = append(ptrs, ptr)
		}
	}

	// Verify no duplicates
	seenAddrs := make(map[uintptr]bool)
	for idx, ptr := range ptrs {
		addr := uintptr(ptr)
		if seenAddrs[addr] {
			t.Errorf("duplicate address at allocation %d: %p", idx, ptr)
		}
		seenAddrs[addr] = true
	}
}

// TestResOwns tests ownership checks
func TestRes_Owns(t *testing.T) {
	r := res.NewRes(4096)
	defer r.Delete()

	ptr := r.Alloc(100, 8)
	if ptr == nil {
		t.Fatal("Alloc returned nil")
	}

	if !r.Owns(ptr) {
		t.Error("Owns should return true for allocated pointer")
	}

	// Test external pointer
	external := new(int)
	if r.Owns(unsafe.Pointer(external)) {
		t.Error("Owns should return false for external pointer")
	}
}

// TestResDelete tests cleanup
func TestRes_Delete(t *testing.T) {
	r := res.NewRes(4096)

	ptr1 := r.Alloc(100, 8)
	if ptr1 == nil {
		t.Fatal("Alloc returned nil")
	}

	r.Delete()

	// After delete, should not own previously allocated memory
	if r.Owns(ptr1) {
		t.Error("Owns should return false after Delete")
	}
}
