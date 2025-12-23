package test

import (
	"testing"
	"unsafe"

	"github.com/thebagchi/arena-go"
	"github.com/thebagchi/arena-go/alloc"
	"github.com/thebagchi/arena-go/res"
)

func TestBuddy_100KInt64(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 100_000
	ptrs := make([]*int64, count)

	for i := range count {
		ptrs[i] = arena.Alloc[int64](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = int64(i)
	}

	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		t.Logf("index %d: ptr=%p, addr=%#x", i, ptrs[i], addr)
	}

	seen := make(map[uintptr]bool)
	pageCounts := make(map[uintptr]int)
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		if *ptrs[i] != int64(i) {
			t.Errorf("index %d: expected %d, got %d", i, i, *ptrs[i])
		}
		page := addr / uintptr(res.PAGE_SIZE)
		pageCounts[page]++
	}

	// Log page usage
	for page, count := range pageCounts {
		pageAddr := page * uintptr(res.PAGE_SIZE)
		t.Logf("Page %#x: %d allocations", pageAddr, count)
	}

	t.Logf("Successfully allocated and verified %d int64 values", count)
}

func TestBuddy_ReallocOrder(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 128
	// First allocation
	ptrs := make([]*int64, count)
	for i := range count {
		ptrs[i] = arena.Alloc[int64](a)
		if ptrs[i] == nil {
			t.Fatalf("first allocation %d failed", i)
		}
		*ptrs[i] = int64(i)
	}

	// Delete all pointers
	for _, ptr := range ptrs {
		arena.DeleteObject(a, ptr)
	}

	// Second allocation and verify in one loop
	for i := range count {
		ptr := arena.Alloc[int64](a)
		if ptr == nil {
			t.Fatalf("second allocation %d failed", i)
		}
		*ptr = int64(i + 1000) // Different value to distinguish

		// Verify address matches first allocation
		addr := uintptr(unsafe.Pointer(ptr))
		if addr != uintptr(unsafe.Pointer(ptrs[i])) {
			temp := uintptr(unsafe.Pointer(ptrs[i]))
			t.Errorf("allocation order mismatch at index %d: expected addr %#x, got %#x", i, temp, addr)
		}
		if *ptr != int64(i+1000) {
			t.Errorf("value mismatch at index %d: expected %d, got %d", i, i+1000, *ptr)
		}
	}

	t.Logf("Successfully verified reallocation order for %d int64 values", count)
}

func TestBuddy_FragmentationRecovery(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 256
	// First allocation of 256 pointers
	ptrs := make([]*int64, count)
	for i := range count {
		ptrs[i] = arena.Alloc[int64](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = int64(i)
	}

	// Free even pointers
	for i := range count {
		if i%2 == 0 {
			arena.DeleteObject(a, ptrs[i])
		}
	}

	// Allocate 128 new pointers
	for i := range 128 {
		ptr := arena.Alloc[int64](a)
		if ptr == nil {
			t.Fatalf("second allocation %d failed", i)
		}
		*ptr = int64(i + 2000)

		// Verify address matches the freed even pointer
		addr := uintptr(unsafe.Pointer(ptr))
		if addr != uintptr(unsafe.Pointer(ptrs[i*2])) {
			temp := uintptr(unsafe.Pointer(ptrs[i*2]))
			t.Errorf("allocation order mismatch at index %d: expected addr %#x, got %#x", i, temp, addr)
		}
		if *ptr != int64(i+2000) {
			t.Errorf("value mismatch at index %d: expected %d, got %d", i, i+2000, *ptr)
		}
	}

	t.Logf("Successfully verified allocation order after freeing even pointers")

	// Free odd pointers
	for i := range count {
		if i%2 == 1 {
			arena.DeleteObject(a, ptrs[i])
		}
	}

	// Allocate 128 more pointers
	for i := range 128 {
		ptr := arena.Alloc[int64](a)
		if ptr == nil {
			t.Fatalf("third allocation %d failed", i)
		}
		*ptr = int64(i + 3000)

		// Verify address matches the freed odd pointer
		addr := uintptr(unsafe.Pointer(ptr))
		if addr != uintptr(unsafe.Pointer(ptrs[i*2+1])) {
			temp := uintptr(unsafe.Pointer(ptrs[i*2+1]))
			t.Errorf("allocation order mismatch at index %d: expected addr %#x, got %#x", i, temp, addr)
		}
		if *ptr != int64(i+3000) {
			t.Errorf("value mismatch at index %d: expected %d, got %d", i, i+3000, *ptr)
		}
	}

	t.Logf("Successfully verified allocation order after freeing odd pointers")
}

func TestBuddy_100KInt32(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 100_000
	ptrs := make([]*int32, count)

	for i := range count {
		ptrs[i] = arena.Alloc[int32](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = int32(i)
	}

	seen := make(map[uintptr]bool)
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		if *ptrs[i] != int32(i) {
			t.Errorf("index %d: expected %d, got %d", i, i, *ptrs[i])
		}
	}

	t.Logf("Successfully allocated and verified %d int32 values", count)
}

func TestBuddy_100KInt16(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 100_000
	ptrs := make([]*int16, count)

	for i := range count {
		ptrs[i] = arena.Alloc[int16](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = int16(i)
	}

	seen := make(map[uintptr]bool)
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		if *ptrs[i] != int16(i) {
			t.Errorf("index %d: expected %d, got %d", i, i, *ptrs[i])
		}
	}

	t.Logf("Successfully allocated and verified %d int16 values", count)
}

func TestBuddy_100KInt8(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 100_000
	ptrs := make([]*int8, count)

	for i := range count {
		ptrs[i] = arena.Alloc[int8](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = int8(i)
	}

	seen := make(map[uintptr]bool)
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		if *ptrs[i] != int8(i) {
			t.Errorf("index %d: expected %d, got %d", i, i, *ptrs[i])
		}
	}

	t.Logf("Successfully allocated and verified %d int8 values", count)
}

func TestBuddy_100KEmpty(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	type Empty struct{}

	const count = 100_000
	ptrs := make([]*Empty, count)

	for i := range count {
		ptrs[i] = arena.Alloc[Empty](a)
		if ptrs[i] == nil {
			t.Fatalf("zero-sized allocation %d failed", i)
		}
	}

	for i, ptr := range ptrs {
		if ptr == nil {
			t.Errorf("zero-sized allocation at index %d is nil", i)
		}
	}

	t.Logf("Successfully allocated and verified %d zero-sized values", count)
}

func TestBuddy_100KByte100(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 100_000
	ptrs := make([]*[100]byte, count)

	for i := range count {
		ptrs[i] = arena.Alloc[[100]byte](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		copy((*ptrs[i])[:], []byte{byte(i % 256), byte((i / 256) % 256), byte((i / 65536) % 256)})
	}

	seen := make(map[uintptr]bool)
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		expected := [3]byte{byte(i % 256), byte((i / 256) % 256), byte((i / 65536) % 256)}
		actual := [3]byte{(*ptrs[i])[0], (*ptrs[i])[1], (*ptrs[i])[2]}
		if actual != expected {
			t.Errorf("index %d: expected %v, got %v", i, expected, actual)
		}
	}

	t.Logf("Successfully allocated and verified %d [100]byte values", count)
}

func TestBuddy_100KFloat32(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 100_000
	ptrs := make([]*float32, count)

	for i := range count {
		ptrs[i] = arena.Alloc[float32](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = float32(i) + 0.5
	}

	seen := make(map[uintptr]bool)
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		expected := float32(i) + 0.5
		if *ptrs[i] != expected {
			t.Errorf("index %d: expected %f, got %f", i, expected, *ptrs[i])
		}
	}

	t.Logf("Successfully allocated and verified %d float32 values", count)
}

func TestBuddy_100KFloat64(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 100_000
	ptrs := make([]*float64, count)

	for i := range count {
		ptrs[i] = arena.Alloc[float64](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = float64(i) + 0.5
	}

	seen := make(map[uintptr]bool)
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		expected := float64(i) + 0.5
		if *ptrs[i] != expected {
			t.Errorf("index %d: expected %f, got %f", i, expected, *ptrs[i])
		}
	}

	t.Logf("Successfully allocated and verified %d float64 values", count)
}

func TestBuddy_100KBool(t *testing.T) {
	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 100_000
	ptrs := make([]*bool, count)

	for i := range count {
		ptrs[i] = arena.Alloc[bool](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = i%2 == 0
	}

	seen := make(map[uintptr]bool)
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		expected := i%2 == 0
		if *ptrs[i] != expected {
			t.Errorf("index %d: expected %t, got %t", i, expected, *ptrs[i])
		}
	}

	t.Logf("Successfully allocated and verified %d bool values", count)
}

func TestBuddy_100KTestStruct(t *testing.T) {
	type Struct struct {
		f1 int8
		f2 int16
		f3 int32
		f4 int64
		f5 bool
		f6 float32
		f7 float64
	}

	a := arena.New(alloc.NewBuddyAllocator())
	defer a.Delete()

	const count = 100_000
	ptrs := make([]*Struct, count)

	for i := range count {
		ptrs[i] = arena.Alloc[Struct](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		ptrs[i].f1 = int8(i % 128)
		ptrs[i].f2 = int16(i % 32768)
		ptrs[i].f3 = int32(i)
		ptrs[i].f4 = int64(i)
		ptrs[i].f5 = i%2 == 0
		ptrs[i].f6 = float32(i) + 0.5
		ptrs[i].f7 = float64(i) + 0.5
	}

	seen := make(map[uintptr]bool)
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true

		expected := Struct{
			f1: int8(i % 128),
			f2: int16(i % 32768),
			f3: int32(i),
			f4: int64(i),
			f5: i%2 == 0,
			f6: float32(i) + 0.5,
			f7: float64(i) + 0.5,
		}

		actual := *ptrs[i]
		if actual != expected {
			t.Errorf("index %d: expected %+v, got %+v", i, expected, actual)
		}
	}

	t.Logf("Successfully allocated and verified %d TestStruct values", count)
}
