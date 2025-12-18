// slab.go
package alloc

import (
	"math/bits"
	"sync"
	"unsafe"

	"github.com/thebagchi/arena-go/alloc/cont"
	"github.com/thebagchi/arena-go/res"
)

const (
	NUM_CLASSES = 17
)

var SIZE_CLASSES = [NUM_CLASSES]uint64{
	16, 32, 64, 128, 256, 512, 1024, 2048, 4096, 8192, 16384,
	32768, 65536, 131072, 262144, 524288, 1048576,
}

type SlabAllocator struct {
	res       *res.Res
	bins      [NUM_CLASSES]*Bin
	freePages *cont.List[*res.Page]
	mtx       sync.Mutex
}

// Bin manages slabs for a single object size class
type Bin struct {
	size      uint64
	align     uint64
	exhausted *cont.List[*Slab]
	available *cont.List[*Slab] // multiple available slabs (LIFO)
	mtx       sync.Mutex
}

type Slab struct {
	page     *res.Page
	freed    unsafe.Pointer // head of singly-linked in-object free list
	usable   int
	capacity int
}

func NewSlabAllocator() *SlabAllocator {
	a := &SlabAllocator{
		res:       res.NewRes(0),
		freePages: cont.NewList[*res.Page](),
	}

	for i := range a.bins {
		a.bins[i] = &Bin{
			size:      SIZE_CLASSES[i],
			align:     8,
			exhausted: cont.NewList[*Slab](),
			available: cont.NewList[*Slab](),
		}
	}

	return a
}

func (a *SlabAllocator) findSizeClass(size, align uint64) uint64 {
	if size == 0 {
		size = 8
	}
	if align == 0 {
		align = 8
	}
	size = ((size + align - 1) / align) * align

	for _, s := range SIZE_CLASSES {
		if s >= size {
			return s
		}
	}
	return res.RoundPow2(size)
}

func indexForSize(classSize uint64) int {
	if classSize <= SIZE_CLASSES[0] {
		return 0
	}
	return bits.TrailingZeros64(classSize) - 4
}

func (a *SlabAllocator) Alloc(size, align uint64) unsafe.Pointer {
	cs := a.findSizeClass(size, align)
	id := indexForSize(cs)

	if id >= NUM_CLASSES {
		return a.res.Alloc(cs, align)
	}

	bin := a.bins[id]
	return bin.Alloc(a)
}

// AllocatePage returns a page for a bin, either from the free pool or by allocating a new one
func (a *SlabAllocator) AllocatePage(size uint64) *res.Page {
	a.mtx.Lock()
	if !a.freePages.IsEmpty() {
		page, _ := a.freePages.PopBack()
		a.mtx.Unlock()
		return page
	}
	a.mtx.Unlock()

	// Allocate new page from res
	var (
		minObjects  = max(64, res.PAGE_SIZE/int(size))
		chunkSize   = minObjects * int(size)
		roundedSize = ((chunkSize + res.PAGE_SIZE - 1) / res.PAGE_SIZE) * res.PAGE_SIZE
	)

	ptr := a.res.Alloc(uint64(roundedSize), 8)
	if ptr == nil {
		return nil
	}

	page, ok := a.res.FindPage(ptr)
	if !ok {
		return nil
	}

	return page
}

func (a *SlabAllocator) Free(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}

	a.mtx.Lock()
	defer a.mtx.Unlock()

	for _, bin := range a.bins {
		if bin != nil && bin.Owns(ptr, a.res) {
			bin.Free(ptr, a)
			return
		}
	}
}

func (b *Bin) Owns(ptr unsafe.Pointer, res *res.Res) bool {
	if ptr == nil {
		return false
	}
	addr := uintptr(ptr)

	if !b.available.IsEmpty() {
		for it := b.available.Iter(); ; {
			s, ok := it.Next()
			if !ok {
				break
			}
			if b.slabContains(s, addr) {
				return true
			}
		}
	}

	if !b.exhausted.IsEmpty() {
		for it := b.exhausted.Iter(); ; {
			s, ok := it.Next()
			if !ok {
				break
			}
			if b.slabContains(s, addr) {
				return true
			}
		}
	}

	return false
}

func (b *Bin) slabContains(s *Slab, addr uintptr) bool {
	if s.page == nil {
		return false
	}
	var (
		start = uintptr(unsafe.Pointer(unsafe.SliceData(s.page.Base())))
		end   = start + uintptr(len(s.page.Base()))
	)
	return addr >= start && addr < end
}

func (b *Bin) removeFromList(l *cont.List[*Slab], target *Slab) {
	for node := l.Front(); node != nil; node = node.Next {
		if node.Value == target {
			l.Remove(node)
			return
		}
	}
}

func (b *Bin) Alloc(allocator *SlabAllocator) unsafe.Pointer {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// Fast path: most recent available slab
	if !b.available.IsEmpty() {
		s, ok := b.available.PopBack()
		if !ok {
			return nil // Should not happen, but handle gracefully
		}
		ptr := b.allocFromSlab(s)
		if ptr == nil {
			b.exhausted.PushBack(s)
		} else {
			b.available.PushBack(s)
		}
		return ptr
	}

	// Try free page pool via allocator
	page := allocator.AllocatePage(b.size)
	if page == nil {
		return nil
	}

	s := b.initSlab(page)
	b.available.PushBack(s)
	return b.allocFromSlab(s)
}

func (b *Bin) initSlab(page *res.Page) *Slab {
	s := &Slab{
		page:     page,
		freed:    nil,
		usable:   0,
		capacity: len(page.Base()) / int(b.size),
	}

	// Build singly-linked free list in objects
	base := unsafe.SliceData(page.Base())
	var prev unsafe.Pointer
	for i := s.capacity - 1; i >= 0; i-- {
		temp := unsafe.Add(unsafe.Pointer(base), i*int(b.size))
		*(*unsafe.Pointer)(temp) = prev
		prev = temp
	}
	s.freed = prev
	s.usable = s.capacity

	return s
}

func (b *Bin) allocFromSlab(s *Slab) unsafe.Pointer {
	if s.freed == nil {
		return nil
	}
	ptr := s.freed
	s.freed = *(*unsafe.Pointer)(ptr)
	s.usable = s.usable - 1
	return ptr
}

func (b *Bin) Free(ptr unsafe.Pointer, allocator *SlabAllocator) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	s := b.findSlab(ptr)
	if s == nil {
		return
	}

	// Prepend to in-object free list
	*(*unsafe.Pointer)(ptr) = s.freed
	s.freed = ptr
	s.usable = s.usable + 1

	if s.usable == s.capacity {
		// Completely empty → return page to free pool
		b.removeFromList(b.available, s)
		b.removeFromList(b.exhausted, s)

		allocator.mtx.Lock()
		allocator.freePages.PushBack(s.page)
		allocator.mtx.Unlock()
	} else {
		// Promote to available list
		b.removeFromList(b.exhausted, s)
		b.available.PushBack(s)
	}
}

func (b *Bin) findSlab(ptr unsafe.Pointer) *Slab {
	addr := uintptr(ptr)

	// Check available list first
	if !b.available.IsEmpty() {
		for it := b.available.Iter(); ; {
			s, ok := it.Next()
			if !ok {
				break
			}
			if b.slabContains(s, addr) {
				return s
			}
		}
	}

	// Check exhausted list
	if !b.exhausted.IsEmpty() {
		for it := b.exhausted.Iter(); ; {
			s, ok := it.Next()
			if !ok {
				break
			}
			if b.slabContains(s, addr) {
				return s
			}
		}
	}

	return nil
}

func (a *SlabAllocator) Reset() {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	a.res.Reset()
	a.freePages.Reset()

	for _, bin := range a.bins {
		if bin != nil {
			bin.mtx.Lock()
			bin.exhausted.Reset()
			bin.available.Reset()
			bin.mtx.Unlock()
		}
	}
}

func (a *SlabAllocator) Delete() {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	if a.res != nil {
		a.res.Delete()
		a.res = nil
	}
	a.freePages = nil
}

func (a *SlabAllocator) Owns(ptr unsafe.Pointer) bool {
	if ptr == nil || a.res == nil {
		return false
	}
	return a.res.Owns(ptr)
}
