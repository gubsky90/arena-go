package test

import (
	"fmt"
	"math/rand"
	"sync"
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
	a := arena.New(alloc.NewBumpAllocator(2000 * res.PAGE_SIZE))
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
		*ptrs[i] = i%2 == 0
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
	type Struct struct {
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

	t.Logf("TestStruct: Sizeof=%d, Alignof=%d", unsafe.Sizeof(Struct{}), unsafe.Alignof(Struct{}))

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

func TestBump_100KStrings(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
	defer a.Delete()

	const count = 100_000
	strings := make([]string, count)

	for i := range count {
		strings[i] = a.MakeString(fmt.Sprintf("string %d", i))
		if strings[i] == "" {
			t.Fatalf("allocation %d failed", i)
		}
	}

	for i := range count {
		addr := uintptr(unsafe.Pointer(unsafe.StringData(strings[i])))
		t.Logf("index %d: str=%q, addr=%#x", i, strings[i], addr)
	}

	seen := make(map[uintptr]bool)
	for i := range count {
		addr := uintptr(unsafe.Pointer(unsafe.StringData(strings[i])))
		if seen[addr] {
			t.Errorf("duplicate address at index %d: %#x", i, addr)
		}
		seen[addr] = true

		expected := fmt.Sprintf("string %d", i)
		if strings[i] != expected {
			t.Errorf("index %d: expected %s, got %s", i, expected, strings[i])
		}
	}

	t.Logf("Successfully allocated and verified %d strings", count)
}

func TestBump_100KTypesAlignment(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
	defer a.Delete()

	const iterations = 100_000
	counter := 0

	allocInt8 := func() {
		ptr := arena.Alloc[int8](a)
		if ptr == nil {
			t.Fatalf("int8 allocation failed at counter %d", counter)
		}
		*ptr = int8(counter % 128)
		counter++
	}

	allocInt16 := func() {
		ptr := arena.Alloc[int16](a)
		if ptr == nil {
			t.Fatalf("int16 allocation failed at counter %d", counter)
		}
		*ptr = int16(counter % 32768)
		addr := uintptr(unsafe.Pointer(ptr))
		if addr%2 != 0 {
			t.Errorf("int16 counter %d: addr %#x not aligned to 2 bytes", counter, addr)
		}
		counter++
	}

	allocInt32 := func() {
		ptr := arena.Alloc[int32](a)
		if ptr == nil {
			t.Fatalf("int32 allocation failed at counter %d", counter)
		}
		*ptr = int32(counter)
		addr := uintptr(unsafe.Pointer(ptr))
		if addr%4 != 0 {
			t.Errorf("int32 counter %d: addr %#x not aligned to 4 bytes", counter, addr)
		}
		counter++
	}

	allocInt64 := func() {
		ptr := arena.Alloc[int64](a)
		if ptr == nil {
			t.Fatalf("int64 allocation failed at counter %d", counter)
		}
		*ptr = int64(counter)
		addr := uintptr(unsafe.Pointer(ptr))
		if addr%8 != 0 {
			t.Errorf("int64 counter %d: addr %#x not aligned to 8 bytes", counter, addr)
		}
		counter++
	}

	allocators := []func(){allocInt8, allocInt16, allocInt32, allocInt64}

	for range iterations {
		rand.Shuffle(len(allocators), func(i, j int) {
			allocators[i], allocators[j] = allocators[j], allocators[i]
		})
		for _, allocFunc := range allocators {
			allocFunc()
		}
	}

	t.Logf("Successfully allocated and verified alignment for %d iterations of random type allocations", iterations)
}

func TestBump_100KArray1000TypesAlignment(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
	defer a.Delete()

	const iterations = 10_000
	counter := 0

	allocInt8 := func() {
		ptr := arena.Alloc[[1000]int8](a)
		if ptr == nil {
			t.Fatalf("int8 array allocation failed at counter %d", counter)
		}
		for i := range *ptr {
			(*ptr)[i] = int8((counter + i) % 128)
		}
		counter++
	}

	allocInt16 := func() {
		ptr := arena.Alloc[[1000]int16](a)
		if ptr == nil {
			t.Fatalf("int16 array allocation failed at counter %d", counter)
		}
		for i := range *ptr {
			(*ptr)[i] = int16((counter + i) % 32768)
		}
		addr := uintptr(unsafe.Pointer(ptr))
		if addr%2 != 0 {
			t.Errorf("int16 array counter %d: addr %#x not aligned to 2 bytes", counter, addr)
		}
		counter++
	}

	allocInt32 := func() {
		ptr := arena.Alloc[[1000]int32](a)
		if ptr == nil {
			t.Fatalf("int32 array allocation failed at counter %d", counter)
		}
		for i := range *ptr {
			(*ptr)[i] = int32(counter + i)
		}
		addr := uintptr(unsafe.Pointer(ptr))
		if addr%4 != 0 {
			t.Errorf("int32 array counter %d: addr %#x not aligned to 4 bytes", counter, addr)
		}
		counter++
	}

	allocInt64 := func() {
		ptr := arena.Alloc[[1000]int64](a)
		if ptr == nil {
			t.Fatalf("int64 array allocation failed at counter %d", counter)
		}
		for i := range *ptr {
			(*ptr)[i] = int64(counter + i)
		}
		addr := uintptr(unsafe.Pointer(ptr))
		if addr%8 != 0 {
			t.Errorf("int64 array counter %d: addr %#x not aligned to 8 bytes", counter, addr)
		}
		counter++
	}

	allocators := []func(){allocInt8, allocInt16, allocInt32, allocInt64}

	for range iterations {
		rand.Shuffle(len(allocators), func(i, j int) {
			allocators[i], allocators[j] = allocators[j], allocators[i]
		})
		for _, allocFunc := range allocators {
			allocFunc()
		}
	}

	t.Logf("Successfully allocated and verified alignment for %d iterations of random array type allocations", iterations)
}

func TestBump_AppendSlice(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
	defer a.Delete()

	slice := arena.MakeSlice[int](a, 0, 4)

	const count = 100_000
	for i := range count {
		slice = arena.Append(a, slice, i)
	}

	if len(slice) != count {
		t.Errorf("Expected length %d, got %d", count, len(slice))
	}

	for i := range count {
		if slice[i] != i {
			t.Errorf("Expected slice[%d] = %d, got %d", i, i, slice[i])
		}
	}

	t.Logf("Successfully appended %d elements to slice: final length %d, capacity %d", count, len(slice), cap(slice))
}

func TestBump_RandomTypesLambda(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
	defer a.Delete()

	const iterations = 100_000
	counter := 0

	allocInt8 := func() {
		ptr := arena.Alloc[int8](a)
		if ptr == nil {
			t.Fatalf("int8 allocation failed at counter %d", counter)
		}
		if !a.Owns(unsafe.Pointer(ptr)) {
			t.Errorf("int8 allocation not owned by arena at counter %d", counter)
		}
		*ptr = int8(counter % 128)
		counter++
	}

	allocInt16 := func() {
		ptr := arena.Alloc[int16](a)
		if ptr == nil {
			t.Fatalf("int16 allocation failed at counter %d", counter)
		}
		if !a.Owns(unsafe.Pointer(ptr)) {
			t.Errorf("int16 allocation not owned by arena at counter %d", counter)
		}
		*ptr = int16(counter % 32768)
		addr := uintptr(unsafe.Pointer(ptr))
		if addr%2 != 0 {
			t.Errorf("int16 counter %d: addr %#x not aligned to 2 bytes", counter, addr)
		}
		counter++
	}

	allocInt32 := func() {
		ptr := arena.Alloc[int32](a)
		if ptr == nil {
			t.Fatalf("int32 allocation failed at counter %d", counter)
		}
		if !a.Owns(unsafe.Pointer(ptr)) {
			t.Errorf("int32 allocation not owned by arena at counter %d", counter)
		}
		*ptr = int32(counter)
		addr := uintptr(unsafe.Pointer(ptr))
		if addr%4 != 0 {
			t.Errorf("int32 counter %d: addr %#x not aligned to 4 bytes", counter, addr)
		}
		counter++
	}

	allocInt64 := func() {
		ptr := arena.Alloc[int64](a)
		if ptr == nil {
			t.Fatalf("int64 allocation failed at counter %d", counter)
		}
		if !a.Owns(unsafe.Pointer(ptr)) {
			t.Errorf("int64 allocation not owned by arena at counter %d", counter)
		}
		*ptr = int64(counter)
		addr := uintptr(unsafe.Pointer(ptr))
		if addr%8 != 0 {
			t.Errorf("int64 counter %d: addr %#x not aligned to 8 bytes", counter, addr)
		}
		counter++
	}

	allocFloat32 := func() {
		ptr := arena.Alloc[float32](a)
		if ptr == nil {
			t.Fatalf("float32 allocation failed at counter %d", counter)
		}
		if !a.Owns(unsafe.Pointer(ptr)) {
			t.Errorf("float32 allocation not owned by arena at counter %d", counter)
		}
		*ptr = float32(counter) + 0.5
		addr := uintptr(unsafe.Pointer(ptr))
		if addr%4 != 0 {
			t.Errorf("float32 counter %d: addr %#x not aligned to 4 bytes", counter, addr)
		}
		counter++
	}

	allocFloat64 := func() {
		ptr := arena.Alloc[float64](a)
		if ptr == nil {
			t.Fatalf("float64 allocation failed at counter %d", counter)
		}
		if !a.Owns(unsafe.Pointer(ptr)) {
			t.Errorf("float64 allocation not owned by arena at counter %d", counter)
		}
		*ptr = float64(counter) + 0.5
		addr := uintptr(unsafe.Pointer(ptr))
		if addr%8 != 0 {
			t.Errorf("float64 counter %d: addr %#x not aligned to 8 bytes", counter, addr)
		}
		counter++
	}

	allocators := []func(){allocInt8, allocInt16, allocInt32, allocInt64, allocFloat32, allocFloat64}

	for range iterations {
		rand.Shuffle(len(allocators), func(i, j int) {
			allocators[i], allocators[j] = allocators[j], allocators[i]
		})
		for _, allocFunc := range allocators {
			allocFunc()
		}
	}

	t.Logf("Successfully allocated and verified alignment for %d iterations of random type allocations", iterations)
}

func TestBump_ExpandCases(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
	defer a.Delete()

	t.Run("ExpandCase1_SingleElementNoGrowth", func(t *testing.T) {
		slice := arena.MakeSlice[int](a, 2, 4)
		slice[0] = 1
		slice[1] = 2

		originalCap := cap(slice)
		originalPtr := &slice[0]

		slice = arena.Append(a, slice, 3)

		if len(slice) != 3 {
			t.Errorf("Expected length 3, got %d", len(slice))
		}
		if cap(slice) != originalCap {
			t.Errorf("Capacity changed unexpectedly: was %d, now %d", originalCap, cap(slice))
		}
		if &slice[0] != originalPtr {
			t.Errorf("Slice backing changed unexpectedly")
		}
		if slice[0] != 1 || slice[1] != 2 || slice[2] != 3 {
			t.Errorf("Slice values incorrect: %v", slice)
		}
	})

	t.Run("ExpandCase2_SingleElementWithGrowth", func(t *testing.T) {
		slice := arena.MakeSlice[int](a, 2, 2)
		slice[0] = 1
		slice[1] = 2

		originalCap := cap(slice)
		originalPtr := &slice[0]

		slice = arena.Append(a, slice, 3)

		if len(slice) != 3 {
			t.Errorf("Expected length 3, got %d", len(slice))
		}
		if cap(slice) <= originalCap {
			t.Errorf("Capacity did not grow: was %d, now %d", originalCap, cap(slice))
		}
		if &slice[0] == originalPtr {
			t.Errorf("Slice backing should have changed")
		}
		if slice[0] != 1 || slice[1] != 2 || slice[2] != 3 {
			t.Errorf("Slice values incorrect: %v", slice)
		}
	})

	t.Run("ExpandCase3_MultiElementNoGrowth", func(t *testing.T) {
		slice := arena.MakeSlice[int](a, 2, 6)
		slice[0] = 1
		slice[1] = 2

		originalCap := cap(slice)
		originalPtr := &slice[0]

		slice = arena.Append(a, slice, 3, 4)

		if len(slice) != 4 {
			t.Errorf("Expected length 4, got %d", len(slice))
		}
		if cap(slice) != originalCap {
			t.Errorf("Capacity changed unexpectedly: was %d, now %d", originalCap, cap(slice))
		}
		if &slice[0] != originalPtr {
			t.Errorf("Slice backing changed unexpectedly")
		}
		expected := []int{1, 2, 3, 4}
		for i, v := range expected {
			if slice[i] != v {
				t.Errorf("Slice[%d] incorrect: expected %d, got %d", i, v, slice[i])
			}
		}
	})

	t.Run("ExpandCase4_MultiElementWithGrowth", func(t *testing.T) {
		slice := arena.MakeSlice[int](a, 2, 3)
		slice[0] = 1
		slice[1] = 2

		originalCap := cap(slice)
		originalPtr := &slice[0]

		slice = arena.Append(a, slice, 3, 4, 5)

		if len(slice) != 5 {
			t.Errorf("Expected length 5, got %d", len(slice))
		}
		if cap(slice) <= originalCap {
			t.Errorf("Capacity did not grow: was %d, now %d", originalCap, cap(slice))
		}
		if &slice[0] == originalPtr {
			t.Errorf("Slice backing should have changed")
		}
		expected := []int{1, 2, 3, 4, 5}
		for i, v := range expected {
			if slice[i] != v {
				t.Errorf("Slice[%d] incorrect: expected %d, got %d", i, v, slice[i])
			}
		}
	})

	t.Run("ExpandCase5_AppendToEmptySlice", func(t *testing.T) {
		slice := arena.MakeSlice[int](a, 0, 2)

		slice = arena.Append(a, slice, 10)

		if len(slice) != 1 {
			t.Errorf("Expected length 1, got %d", len(slice))
		}
		if cap(slice) < 1 {
			t.Errorf("Capacity should be at least 1, got %d", cap(slice))
		}
		if slice[0] != 10 {
			t.Errorf("Slice[0] incorrect: expected 10, got %d", slice[0])
		}
	})
}

func TestBump_StressTest(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
	defer a.Delete()

	const totalIterations = 10_000_000
	const resetInterval = 10_000

	counter := 0

	allocInt32 := func() {
		ptr := arena.Alloc[int32](a)
		if ptr == nil {
			t.Fatalf("int32 allocation failed at counter %d", counter)
		}
		*ptr = int32(counter)
		counter++
	}

	allocInt64 := func() {
		ptr := arena.Alloc[int64](a)
		if ptr == nil {
			t.Fatalf("int64 allocation failed at counter %d", counter)
		}
		*ptr = int64(counter)
		counter++
	}

	allocFloat64 := func() {
		ptr := arena.Alloc[float64](a)
		if ptr == nil {
			t.Fatalf("float64 allocation failed at counter %d", counter)
		}
		*ptr = float64(counter) + 0.5
		counter++
	}

	allocSlice := func() {
		slice := arena.MakeSlice[int](a, 0, 10)
		for i := 0; i < 5; i++ {
			slice = arena.Append(a, slice, counter+i)
		}
		counter += 5
	}

	allocators := []func(){allocInt32, allocInt64, allocFloat64, allocSlice}

	for i := 0; i < totalIterations; i++ {
		// Randomly select an allocator
		idx := rand.Intn(len(allocators))
		allocators[idx]()

		// Reset arena periodically
		if (i+1)%resetInterval == 0 {
			a.Reset()
			counter = 0 // Reset counter to avoid large numbers
		}
	}

	t.Logf("Stress test completed: %d iterations with periodic resets", totalIterations)
}

func TestBump_ConcurrentAllocations(t *testing.T) {
	const numGoroutines = 10
	const allocationsPerGoroutine = 10000

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines)

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
			defer a.Delete()

			// Allocate and verify various types
			for j := 0; j < allocationsPerGoroutine; j++ {
				intPtr := arena.Alloc[int](a)
				if intPtr == nil {
					errors <- fmt.Errorf("goroutine %d: failed to alloc int %d", id, j)
					return
				}
				*intPtr = j

				int64Ptr := arena.Alloc[int64](a)
				if int64Ptr == nil {
					errors <- fmt.Errorf("goroutine %d: failed to alloc int64 %d", id, j)
					return
				}
				*int64Ptr = int64(j)

				float64Ptr := arena.Alloc[float64](a)
				if float64Ptr == nil {
					errors <- fmt.Errorf("goroutine %d: failed to alloc float64 %d", id, j)
					return
				}
				*float64Ptr = float64(j) + 0.5

				boolPtr := arena.Alloc[bool](a)
				if boolPtr == nil {
					errors <- fmt.Errorf("goroutine %d: failed to alloc bool %d", id, j)
					return
				}
				*boolPtr = j%2 == 0

				str := a.MakeString(fmt.Sprintf("goroutine %d string %d", id, j))
				if str == "" {
					errors <- fmt.Errorf("goroutine %d: failed to make string %d", id, j)
					return
				}

				arrayPtr := arena.Alloc[[10]int](a)
				if arrayPtr == nil {
					errors <- fmt.Errorf("goroutine %d: failed to alloc array %d", id, j)
					return
				}
				for k := range *arrayPtr {
					(*arrayPtr)[k] = j + k
				}

				// Verify immediately
				if *intPtr != j {
					errors <- fmt.Errorf("goroutine %d: int verification failed for %d", id, j)
					return
				}
				if *int64Ptr != int64(j) {
					errors <- fmt.Errorf("goroutine %d: int64 verification failed for %d", id, j)
					return
				}
				if *float64Ptr != float64(j)+0.5 {
					errors <- fmt.Errorf("goroutine %d: float64 verification failed for %d", id, j)
					return
				}
				if *boolPtr != (j%2 == 0) {
					errors <- fmt.Errorf("goroutine %d: bool verification failed for %d", id, j)
					return
				}
				expectedStr := fmt.Sprintf("goroutine %d string %d", id, j)
				if str != expectedStr {
					errors <- fmt.Errorf("goroutine %d: string verification failed for %d: got %q, expected %q", id, j, str, expectedStr)
					return
				}
				for k := range *arrayPtr {
					if (*arrayPtr)[k] != j+k {
						errors <- fmt.Errorf("goroutine %d: array verification failed for %d, index %d", id, j, k)
						return
					}
				}
			}

			a.Reset()
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		if err != nil {
			t.Error(err)
		}
	}

	t.Logf("Concurrent allocations completed successfully: %d goroutines, %d allocations each", numGoroutines, allocationsPerGoroutine)
}

func BenchmarkBump_Allocations(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(res.PAGE_SIZE))
	defer a.Delete()

	b.ReportAllocs()

	for b.Loop() {
		// Allocate native types
		arena.Alloc[int](a)
		arena.Alloc[int8](a)
		arena.Alloc[int16](a)
		arena.Alloc[int32](a)
		arena.Alloc[int64](a)
		arena.Alloc[uint](a)
		arena.Alloc[float32](a)
		arena.Alloc[float64](a)
		arena.Alloc[bool](a)

		// Allocate arrays
		arena.Alloc[[100]int](a)
		arena.Alloc[[100]byte](a)
		arena.Alloc[[100]int8](a)
		arena.Alloc[[100]int16](a)
		arena.Alloc[[100]int32](a)
		arena.Alloc[[100]int64](a)
		arena.Alloc[[100]uint](a)
		arena.Alloc[[100]float32](a)
		arena.Alloc[[100]float64](a)
		arena.Alloc[[100]bool](a)

		// Allocate slices
		arena.MakeSlice[int](a, 100, 200)
		arena.MakeSlice[string](a, 100, 200)
		arena.MakeSlice[int8](a, 100, 200)
		arena.MakeSlice[int16](a, 100, 200)
		arena.MakeSlice[int32](a, 100, 200)
		arena.MakeSlice[int64](a, 100, 200)
		arena.MakeSlice[uint](a, 100, 200)
		arena.MakeSlice[float32](a, 100, 200)
		arena.MakeSlice[float64](a, 100, 200)
		arena.MakeSlice[bool](a, 100, 200)

		// Allocate strings
		a.MakeString("benchmark string")

		// Allocate structs
		type TestStruct struct {
			f01 [100]int
			f02 [100]int8
			f03 [100]int16
			f04 [100]int32
			f05 [100]int64
			f06 [100]uint
			f07 [100]float32
			f08 [100]float64
			f09 [100]bool
			f10 [100]string
			f11 [100]byte
			f12 int
			f13 int8
			f14 int16
			f15 int32
			f16 int64
			f17 uint
			f18 float32
			f19 float64
			f20 bool
			f21 string
			f22 byte
		}
		arena.Alloc[TestStruct](a)

		// Allocate empty struct
		type Empty struct{}
		arena.Alloc[Empty](a)

		a.Reset()
	}
}
