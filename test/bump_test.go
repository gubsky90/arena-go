package test

import (
	"testing"
	"unsafe"

	"github.com/thebagchi/arena-go"
	"github.com/thebagchi/arena-go/alloc"
	"github.com/thebagchi/arena-go/res"
)

func TestBump_100KInt64(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
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
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		if *ptrs[i] != int64(i) {
			t.Errorf("index %d: expected %d, got %d", i, i, *ptrs[i])
		}
	}

	t.Logf("Successfully allocated and verified %d int64 values", count)
}

func TestBump_1MInt64(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(2000 * res.PAGE_SIZE)) // Start with 2000 pages (~8MB)
	defer a.Delete()

	const count = 1_000_000
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
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true
		if *ptrs[i] != int64(i) {
			t.Errorf("index %d: expected %d, got %d", i, i, *ptrs[i])
		}
	}

	t.Logf("Successfully allocated and verified %d int64 values", count)
}

func TestBump_100KInt32(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
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

	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		t.Logf("index %d: ptr=%p, addr=%#x", i, ptrs[i], addr)
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

func TestBump_100KInt16(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
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

	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		t.Logf("index %d: ptr=%p, addr=%#x", i, ptrs[i], addr)
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

func TestBump_100KInt8(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
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

	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		t.Logf("index %d: ptr=%p, addr=%#x", i, ptrs[i], addr)
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

func TestBump_100KEmpty(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
	defer a.Delete()

	type Empty struct{}

	t.Logf("Empty struct: Sizeof=%d, Alignof=%d", unsafe.Sizeof(Empty{}), unsafe.Alignof(Empty{}))

	const count = 100_000
	ptrs := make([]*Empty, count)

	for i := range count {
		ptrs[i] = arena.Alloc[Empty](a)
		if ptrs[i] == nil {
			t.Fatalf("zero-sized allocation %d failed", i)
		}
	}

	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		t.Logf("index %d: ptr=%p, addr=%#x", i, ptrs[i], addr)
	}

	for i, ptr := range ptrs {
		if ptr == nil {
			t.Errorf("zero-sized allocation at index %d is nil", i)
		}
	}

	t.Logf("Successfully allocated and verified %d zero-sized values", count)
}

func TestBump_100KByte100(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
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

	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		t.Logf("index %d: ptr=%p, addr=%#x", i, ptrs[i], addr)
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

func TestBump_100KFloat32(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
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

	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		t.Logf("index %d: ptr=%p, addr=%#x", i, ptrs[i], addr)
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

func TestBump_100KFloat64(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
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

	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		t.Logf("index %d: ptr=%p, addr=%#x", i, ptrs[i], addr)
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

func TestBump_100KBool(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
	defer a.Delete()

	const count = 100_000
	ptrs := make([]*bool, count)

	for i := range count {
		ptrs[i] = arena.Alloc[bool](a)
		if ptrs[i] == nil {
			t.Fatalf("allocation %d failed", i)
		}
		*ptrs[i] = i%2 == 0 // Alternate between true and false
	}

	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		t.Logf("index %d: ptr=%p, addr=%#x", i, ptrs[i], addr)
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

func TestBump_100KTestStruct(t *testing.T) {
	type TestStruct struct {
		f1 int8
		f2 int16
		f3 int32
		f4 int64
		f5 bool
		f6 float32
		f7 float64
	}

	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
	defer a.Delete()

	t.Logf("TestStruct: Sizeof=%d, Alignof=%d", unsafe.Sizeof(TestStruct{}), unsafe.Alignof(TestStruct{}))

	const count = 100_000
	ptrs := make([]*TestStruct, count)

	for i := range count {
		ptrs[i] = arena.Alloc[TestStruct](a)
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

	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		t.Logf("index %d: ptr=%p, addr=%#x", i, ptrs[i], addr)
	}

	seen := make(map[uintptr]bool)
	for i := range count {
		addr := uintptr(unsafe.Pointer(ptrs[i]))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true

		expected := TestStruct{
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
