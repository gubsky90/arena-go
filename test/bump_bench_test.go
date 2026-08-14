package test

import (
	"testing"
	"unsafe"

	"github.com/gubsky90/arena-go"
	"github.com/gubsky90/arena-go/alloc"
)

// Benchmark basic allocation performance
func BenchmarkBump_Alloc(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(100 * 4096))
	defer a.Delete()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arena.Alloc[int](a)
	}
}

// Benchmark allocation with reset
func BenchmarkBump_AllocWithReset(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(100 * 4096))
	defer a.Delete()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arena.Alloc[int](a)
		if i%1000 == 0 {
			a.Reset()
		}
	}
}

// Benchmark allocation and deallocation (bump doesn't truly free, but test for consistency)
func BenchmarkBump_AllocFree(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(100 * 4096))
	defer a.Delete()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := arena.Alloc[int](a)
		arena.DeleteObject(a, p)
	}
}

// Benchmark large allocations
func BenchmarkBump_AllocLarge(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(100 * 4096))
	defer a.Delete()

	type Large struct {
		data [1024]byte
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arena.Alloc[Large](a)
	}
}

// Benchmark mixed workload
func BenchmarkBump_MixedWorkload(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(100 * 4096))
	defer a.Delete()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arena.Alloc[int](a)
		_ = arena.Alloc[int64](a)
		_ = arena.Alloc[float64](a)
	}
}

// Benchmark Owns operation
func BenchmarkBump_Owns(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(100 * 4096))
	defer a.Delete()

	p := arena.Alloc[int](a)
	*p = 42

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = a.Owns(unsafe.Pointer(p))
	}
}

// Benchmark make slice
func BenchmarkBump_MakeSlice(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1000 * 4096))
	defer a.Delete()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		slice := arena.MakeSlice[int](a, 10, 10)
		slice[0] = i
	}
}

// Benchmark make string
func BenchmarkBump_MakeString(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1000 * 4096))
	defer a.Delete()

	str := "benchmark string for allocation"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := a.MakeString(str)
		_ = s
	}
}

// Benchmark comparison with standard allocator
func BenchmarkBump_VsStandard_Alloc(b *testing.B) {
	b.Run("Bump", func(b *testing.B) {
		a := arena.New(alloc.NewBumpAllocator(100 * 4096))
		defer a.Delete()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = arena.Alloc[int](a)
		}
	})

	b.Run("Standard", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = new(int)
		}
	})
}

// Benchmark batch allocations
func BenchmarkBump_BatchAlloc(b *testing.B) {
	var sizes = []int{10, 100, 1000}

	for _, size := range sizes {
		b.Run(string(rune('0'+size/10)), func(b *testing.B) {
			a := arena.New(alloc.NewBumpAllocator(1000 * 4096))
			defer a.Delete()

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for j := 0; j < size; j++ {
					_ = arena.Alloc[int](a)
				}
				a.Reset()
			}
		})
	}
}

// Benchmark reset overhead
func BenchmarkBump_Reset(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(100 * 4096))
	defer a.Delete()

	// Pre-allocate some objects
	for i := 0; i < 1000; i++ {
		_ = arena.Alloc[int](a)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Reset()
		// Re-allocate to keep state consistent
		for j := 0; j < 1000; j++ {
			_ = arena.Alloc[int](a)
		}
	}
}

// Benchmark struct allocation
func BenchmarkBump_AllocStruct(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(100 * 4096))
	defer a.Delete()

	type TestStruct struct {
		A int
		B string
		C [10]int
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := arena.Alloc[TestStruct](a)
		_ = p
	}
}

// Benchmark array allocation
func BenchmarkBump_AllocArray(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(100 * 4096))
	defer a.Delete()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = arena.Alloc[[100]int](a)
	}
}

// Benchmark stress test
func BenchmarkBump_StressTest(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1000 * 4096))
	defer a.Delete()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			_ = arena.Alloc[int](a)
		}
		a.Reset()
	}
}
