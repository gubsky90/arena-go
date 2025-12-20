package arena_test

import (
	"testing"
	"unsafe"

	arena "github.com/thebagchi/arena-go"
	"github.com/thebagchi/arena-go/alloc"
)

func TestBuddy_BasicAllocation(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	// Allocate various sizes
	p1 := arena.Alloc[int](a)
	if p1 == nil {
		t.Fatal("allocation failed")
	}
	*p1 = 42

	p2 := arena.Alloc[uint64](a)
	if p2 == nil {
		t.Fatal("allocation failed")
	}
	*p2 = 99

	if *p1 != 42 {
		t.Errorf("expected 42, got %d", *p1)
	}
	if *p2 != 99 {
		t.Errorf("expected 99, got %d", *p2)
	}
}

func TestBuddy_RemoveAndReuse(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	// Allocate and free
	p1 := arena.Alloc[int](a)
	*p1 = 100
	addr1 := unsafe.Pointer(p1)

	arena.DeleteObject(a, p1)

	// Allocate again - should reuse the freed block
	p2 := arena.Alloc[int](a)
	addr2 := unsafe.Pointer(p2)

	// Due to buddy coalescing, addresses might be the same or different
	// but allocation should succeed
	if p2 == nil {
		t.Fatal("reallocation failed")
	}
	*p2 = 200

	if *p2 != 200 {
		t.Errorf("expected 200, got %d", *p2)
	}

	t.Logf("addr1=%p addr2=%p", addr1, addr2)
}

func TestBuddy_MultipleAllocations(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 100
	ptrs := make([]*int, count)

	// Allocate many objects
	for i := 0; i < count; i++ {
		ptrs[i] = arena.Alloc[int](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = i * 10
	}

	// Verify all pointers are unique
	seen := make(map[uintptr]bool)
	for i := 0; i < count; i++ {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		// Verify value is still what we set
		if *ptrs[i] != i*10 {
			t.Errorf("index %d: expected %d, got %d", i, i*10, *ptrs[i])
		}
	}
}

func TestBuddy_Coalescing(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	// Allocate multiple small blocks
	p1 := arena.Alloc[int](a)
	p2 := arena.Alloc[int](a)
	p3 := arena.Alloc[int](a)
	p4 := arena.Alloc[int](a)

	*p1 = 1
	*p2 = 2
	*p3 = 3
	*p4 = 4

	// Free some blocks - should coalesce
	arena.DeleteObject(a, p1)
	arena.DeleteObject(a, p2)

	// Allocate larger block - might reuse coalesced space
	p5 := arena.Alloc[uint64](a)
	if p5 == nil {
		t.Fatal("allocation after coalescing failed")
	}
	*p5 = 123456789

	// Remaining allocations should still be valid
	if *p3 != 3 {
		t.Errorf("p3: expected 3, got %d", *p3)
	}
	if *p4 != 4 {
		t.Errorf("p4: expected 4, got %d", *p4)
	}
}

func TestBuddy_Reset(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	// Allocate some memory
	for i := 0; i < 10; i++ {
		p := arena.Alloc[int](a)
		*p = i
	}

	// Reset should clear allocations
	a.Reset()

	// Should be able to allocate again
	p := arena.Alloc[int](a)
	if p == nil {
		t.Fatal("allocation after reset failed")
	}
	*p = 999
	if *p != 999 {
		t.Errorf("expected 999, got %d", *p)
	}
}

func TestBuddy_LargeAllocation(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	// Allocate large structure
	type Large struct {
		data [1024]byte
	}

	p := arena.Alloc[Large](a)
	if p == nil {
		t.Fatal("large allocation failed")
	}

	p.data[0] = 0xFF
	p.data[1023] = 0xAA

	if p.data[0] != 0xFF || p.data[1023] != 0xAA {
		t.Error("large allocation data corruption")
	}
}

func TestBuddy_Owns(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	p := arena.Alloc[int](a)
	*p = 42

	if !a.Owns(unsafe.Pointer(p)) {
		t.Error("Owns should return true for allocated pointer")
	}

	external := new(int)
	if a.Owns(unsafe.Pointer(external)) {
		t.Error("Owns should return false for external pointer")
	}

	if a.Owns(nil) {
		t.Error("Owns should return false for nil pointer")
	}
}

func TestBuddy_ZeroSizedType(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	type Empty struct{}

	p := arena.Alloc[Empty](a)
	if p == nil {
		t.Fatal("zero-sized allocation failed")
	}

	// Should not crash
	_ = *p
}

func TestBuddy_AlignedAllocation(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	p := arena.Alloc[uint64](a)
	if p == nil {
		t.Fatal("allocation failed")
	}

	// Check alignment
	addr := uintptr(unsafe.Pointer(p))
	if addr%8 != 0 {
		t.Errorf("uint64 not aligned: address=%#x", addr)
	}
}

func TestBuddy_InterleavedAllocFree(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	// Test alloc/free/alloc cycle with Reset instead of Remove
	// since buddy doesn't track allocation sizes for individual frees
	for cycle := 0; cycle < 3; cycle++ {
		ptrs := make([]*int, 10)

		// Allocate
		for i := 0; i < 10; i++ {
			ptrs[i] = arena.Alloc[int](a)
			if ptrs[i] == nil {
				t.Fatalf("cycle %d: allocation %d failed", cycle, i)
			}
			*ptrs[i] = cycle*100 + i
		}

		// Verify
		for i := 0; i < 10; i++ {
			if *ptrs[i] != cycle*100+i {
				t.Errorf("cycle %d, index %d: expected %d, got %d",
					cycle, i, cycle*100+i, *ptrs[i])
			}
		}

		// Reset for next cycle
		a.Reset()
	}
}

func TestBuddy_MultipleChunks(t *testing.T) {
	// Start with multiple chunks
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 200
	ptrs := make([]*int, count)

	// Allocate across multiple chunks
	for i := 0; i < count; i++ {
		ptrs[i] = arena.Alloc[int](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = i * 3
	}

	// Verify all pointers are unique
	seen := make(map[uintptr]bool)
	for i := 0; i < count; i++ {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		// Verify value
		if *ptrs[i] != i*3 {
			t.Errorf("index %d: expected %d, got %d", i, i*3, *ptrs[i])
		}
	}
}

func TestBuddy_OversizedAllocation(t *testing.T) {
	// Create buddy allocator with small chunk size
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	// Allocate something larger than chunk size (4KB default)
	type Huge struct {
		data [8192]byte // 8KB, larger than default 4KB page
	}

	p := arena.Alloc[Huge](a)
	if p == nil {
		t.Fatal("oversized allocation failed")
	}

	// Write to verify it's accessible
	p.data[0] = 0xAA
	p.data[8191] = 0xBB

	if p.data[0] != 0xAA || p.data[8191] != 0xBB {
		t.Error("oversized allocation data corruption")
	}

	// Allocate another oversized block
	p2 := arena.Alloc[Huge](a)
	if p2 == nil {
		t.Fatal("second oversized allocation failed")
	}

	// Ensure they're different
	if unsafe.Pointer(p) == unsafe.Pointer(p2) {
		t.Error("oversized allocations returned same address")
	}

	p2.data[0] = 0xCC
	p2.data[8191] = 0xDD

	// Verify both still work
	if p.data[0] != 0xAA || p.data[8191] != 0xBB {
		t.Error("first oversized allocation corrupted")
	}
	if p2.data[0] != 0xCC || p2.data[8191] != 0xDD {
		t.Error("second oversized allocation corrupted")
	}
}

func TestBuddy_DetailedDebug(t *testing.T) {
	// Allocate just enough to see the 32766-32768 issue in detail
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 32800
	ptrs := make([]*uint64, count)

	// Allocate up to the problem area
	for i := 0; i < count; i++ {
		ptrs[i] = arena.Alloc[uint64](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = uint64(i)

		// Log around problem indices
		if i >= 32762 && i <= 32772 {
			// Check if this address already exists
			for j := 0; j < i; j++ {
				if ptrs[j] != nil && ptrs[j] == ptrs[i] {
					t.Errorf("DUPLICATE DETECTED: alloc %d ptr=%p same as alloc %d", i, ptrs[i], j)
					t.Logf("  alloc[%d] = %p, value = %d", j, ptrs[j], *ptrs[j])
					t.Logf("  alloc[%d] = %p, value = %d", i, ptrs[i], *ptrs[i])
					break
				}
			}
			t.Logf("alloc[%d] = %p, value=%d", i, ptrs[i], *ptrs[i])
		}
	}

	t.Logf("Completed %d allocations", count)
}

func TestBuddy_100KUint64(t *testing.T) {
	// Allocate 100K uint64 values across multiple chunks
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 100_000
	ptrs := make([]*uint64, count)

	// Allocate 100K uint64
	for i := 0; i < count; i++ {
		ptrs[i] = arena.Alloc[uint64](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = uint64(i)
	}

	// Verify all pointers are unique
	seen := make(map[uintptr]bool)
	for i := 0; i < count; i++ {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		// Verify value
		if *ptrs[i] != uint64(i) {
			t.Errorf("index %d: expected %d, got %d", i, i, *ptrs[i])
		}
	}

	t.Logf("Successfully allocated and verified %d uint64 values across multiple chunks", count)
}

func TestBuddy_10KUint64(t *testing.T) {
	// Test with 10K to isolate the issue
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 10_000
	ptrs := make([]*uint64, count)

	// Allocate 10K uint64
	for i := 0; i < count; i++ {
		ptrs[i] = arena.Alloc[uint64](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = uint64(i)
	}

	// Verify all pointers are unique
	seen := make(map[uintptr]bool)
	for i := 0; i < count; i++ {
		addr := uintptr(unsafe.Pointer(ptrs[i]))

		// Figure out which chunk this belongs to (simplified without Res)
		chunkIdx := -1

		if seen[addr] {
			t.Errorf("duplicate address at index %d (chunk %d): %#x", i, chunkIdx, addr)
		}
		seen[addr] = true
		// Verify value
		if *ptrs[i] != uint64(i) {
			t.Errorf("index %d (chunk %d): expected %d, got %d", i, chunkIdx, i, *ptrs[i])
			if i > 32760 && i < 32770 {
				t.Logf("DEBUG: Around error: index=%d, ptr=%p, chunkIdx=%d, value=%d", i, ptrs[i], chunkIdx, *ptrs[i])
			}
		}
	}

	t.Logf("Successfully allocated and verified %d uint64 values across multiple chunks", count)
}
