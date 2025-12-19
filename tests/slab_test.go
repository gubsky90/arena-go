package arena_test

import (
	"testing"
	"unsafe"

	arena "github.com/thebagchi/arena-go"
	"github.com/thebagchi/arena-go/alloc"
)

func TestSlab_BasicAllocation(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	// Allocate small objects
	p1 := arena.Alloc[int](a)
	if p1 == nil {
		t.Fatal("allocation failed")
	}
	*p1 = 42

	p2 := arena.Alloc[int64](a)
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

func TestSlab_SizeClasses(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	// Test allocations of various sizes using MakeSlice
	type TestCase struct {
		name string
		size int
	}

	cases := []TestCase{
		{"tiny", 1},
		{"small", 16},
		{"medium", 256},
		{"large", 4096},
		{"xlarge", 262144},
	}

	for _, tc := range cases {
		p := arena.MakeSlice[byte](a, tc.size, tc.size)
		if len(p) == 0 {
			t.Fatalf("%s: allocation failed", tc.name)
		}
		// Write to allocated memory to verify it's accessible
		p[0] = 42
	}
}

func TestSlab_MultipleObjects(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	const count = 100
	ptrs := make([]*int, count)

	// Allocate many small objects
	for i := 0; i < count; i++ {
		ptrs[i] = arena.Alloc[int](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = i * 10
	}

	// Verify all values
	for i := 0; i < count; i++ {
		if *ptrs[i] != i*10 {
			t.Errorf("index %d: expected %d, got %d", i, i*10, *ptrs[i])
		}
	}
}

func TestSlab_RemoveAndReuse(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	// Allocate and free
	p1 := arena.Alloc[int](a)
	*p1 = 100
	addr1 := unsafe.Pointer(p1)

	arena.DeleteObject(a, p1)

	// Allocate again - should potentially reuse the freed block
	p2 := arena.Alloc[int](a)
	addr2 := unsafe.Pointer(p2)

	if p2 == nil {
		t.Fatal("reallocation failed")
	}
	*p2 = 200

	if *p2 != 200 {
		t.Errorf("expected 200, got %d", *p2)
	}

	t.Logf("addr1=%p addr2=%p (may or may not reuse)", addr1, addr2)
}

func TestSlab_InterleavedAllocFree(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	// Allocate, free, allocate pattern
	ptrs := make([]*int, 100)

	for i := 0; i < 100; i++ {
		ptrs[i] = arena.Alloc[int](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = i
	}

	// Free every other one
	for i := 0; i < 100; i += 2 {
		arena.DeleteObject(a, ptrs[i])
		ptrs[i] = nil
	}

	// Allocate more - should fill freed slots
	for i := 0; i < 50; i++ {
		p := arena.Alloc[int](a)
		if p == nil {
			t.Fatalf("reallocation %d failed", i)
		}
		*p = i + 1000
	}

	// Verify remaining originals
	for i := 1; i < 100; i += 2 {
		if *ptrs[i] != i {
			t.Errorf("index %d: expected %d, got %d", i, i, *ptrs[i])
		}
	}
}

func TestSlab_DifferentTypes(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	// Allocate different types
	pi := arena.Alloc[int](a)
	*pi = 42

	ps := arena.Alloc[string](a)
	*ps = "hello"

	pf := arena.Alloc[float64](a)
	*pf = 3.14

	pb := arena.Alloc[bool](a)
	*pb = true

	// Verify
	if *pi != 42 {
		t.Errorf("int: expected 42, got %d", *pi)
	}
	if *ps != "hello" {
		t.Errorf("string: expected 'hello', got %q", *ps)
	}
	if *pf != 3.14 {
		t.Errorf("float64: expected 3.14, got %f", *pf)
	}
	if *pb != true {
		t.Errorf("bool: expected true, got %v", *pb)
	}
}

func TestSlab_StructAllocation(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	type TestStruct struct {
		A int
		B string
		C float64
	}

	p := arena.Alloc[TestStruct](a)
	if p == nil {
		t.Fatal("struct allocation failed")
	}

	p.A = 42
	p.B = "test"
	p.C = 3.14

	if p.A != 42 || p.B != "test" || p.C != 3.14 {
		t.Error("struct field values incorrect")
	}
}

func TestSlab_ArrayAllocation(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	// Allocate array
	arr := arena.MakeSlice[int](a, 100, 100)
	if len(arr) == 0 {
		t.Fatal("array allocation failed")
	}

	// Fill with values
	for i := 0; i < 100; i++ {
		arr[i] = i * 2
	}

	// Verify
	for i := 0; i < 100; i++ {
		if arr[i] != i*2 {
			t.Errorf("index %d: expected %d, got %d", i, i*2, arr[i])
		}
	}
}

func TestSlab_Owns(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	p := arena.Alloc[int](a)
	if !a.Owns(unsafe.Pointer(p)) {
		t.Error("allocator should own allocated pointer")
	}

	// Test with unowned pointer
	var x int
	if a.Owns(unsafe.Pointer(&x)) {
		t.Error("allocator should not own stack pointer")
	}
}

func TestSlab_Reset(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	// Allocate before reset
	p1 := arena.Alloc[int](a)
	*p1 = 42

	a.Reset()

	// Allocate after reset
	p2 := arena.Alloc[int](a)
	if p2 == nil {
		t.Fatal("allocation after reset failed")
	}
	*p2 = 99

	if *p2 != 99 {
		t.Errorf("expected 99, got %d", *p2)
	}
}

func TestSlab_ZeroSizedAllocation(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	// Allocate zero-length array
	arr := arena.MakeSlice[int](a, 0, 0)
	if len(arr) != 0 {
		t.Error("zero-length allocation should be empty")
	}
}

func TestSlab_LargeAllocations(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	// Allocate increasingly large arrays
	sizes := []int{1024, 4096, 16384, 65536, 262144}

	for _, size := range sizes {
		arr := arena.MakeSlice[byte](a, size, size)
		if len(arr) == 0 {
			t.Fatalf("allocation of size %d failed", size)
		}

		// Write and verify at boundaries
		arr[0] = 42
		arr[size-1] = 99

		if arr[0] != 42 {
			t.Errorf("size %d: first byte incorrect", size)
		}
		if arr[size-1] != 99 {
			t.Errorf("size %d: last byte incorrect", size)
		}
	}
}

func TestSlab_ConcurrentAllocations(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	// Note: actual concurrent access would need proper synchronization
	// This test just verifies many rapid allocations work
	const count = 100
	ptrs := make([]*int, count)

	for i := 0; i < count; i++ {
		ptrs[i] = arena.Alloc[int](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = i
	}

	// Verify
	for i := 0; i < count; i++ {
		if *ptrs[i] != i {
			t.Errorf("index %d: expected %d, got %d", i, i, *ptrs[i])
		}
	}
}

func TestSlab_MixedSizes(t *testing.T) {
	a := arena.New(alloc.NewSlabAllocator())
	defer a.Delete()

	// Allocate mix of small and large objects
	for i := 0; i < 20; i++ {
		// Small object
		ps := arena.Alloc[int](a)
		if ps == nil {
			t.Fatalf("small allocation %d failed", i)
		}
		*ps = i

		// Medium object
		pm := arena.MakeSlice[byte](a, 256, 256)
		if len(pm) == 0 {
			t.Fatalf("medium allocation %d failed", i)
		}
		pm[0] = byte(i % 256)

		// Large object every 5 iterations
		if i%5 == 0 {
			pl := arena.MakeSlice[int](a, 100, 100)
			if len(pl) == 0 {
				t.Fatalf("large allocation %d failed", i)
			}
			pl[0] = i * 10
		}
	}
}
