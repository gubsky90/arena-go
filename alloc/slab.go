package alloc

import (
	"unsafe"

	"github.com/thebagchi/arena-go/res"
)

type SlabAllocator struct {
	blockSize uintptr
	res       *res.Res
}

func NewSlabAllocator(r *res.Res, blockSize, totalBytes int) *SlabAllocator {
	if blockSize < 16 {
		blockSize = 16
	}
	blockSize = (blockSize + 15) &^ 15
	s := &SlabAllocator{blockSize: uintptr(blockSize), res: r}
	// dummy implementation, no actual allocation
	return s
}

func (s *SlabAllocator) Alloc(size, align uint64) unsafe.Pointer {
	// dummy
	return nil
}

func (s *SlabAllocator) Reset() {
	// dummy
}

func (s *SlabAllocator) Delete() {
	// dummy
}

func (s *SlabAllocator) Remove(ptr unsafe.Pointer) {
	// no op for slab allocator
}

func (s *SlabAllocator) Owns(ptr unsafe.Pointer) bool {
	// TODO: implement when slab allocator is fully implemented
	return s.res.Owns(ptr)
}
