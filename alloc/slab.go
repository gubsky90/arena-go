// slab.go
package alloc

import (
	"sync"
	"unsafe"

	"github.com/gubsky90/arena-go/res"
)

const (
	NUM_CLASSES = 32
)

var SIZE_CLASSES = [NUM_CLASSES]uint64{
	16, 32, 64, 128,
	256, 512, 1024, 2048,
	4096, 8192, 16384, 32768,
	65536, 131072, 262144, 524288,
	1048576, 2097152, 4194304, 8388608,
	16777216, 33554432, 67108864, 134217728,
	268435456, 536870912, 1073741824, 2147483648,
	4294967296, 8589934592, 17179869184, 34359738368,
}

type SlabAllocator struct {
	mtx       sync.Mutex
	res       *res.Res
	bins      [NUM_CLASSES]*Bin
	freePages []*res.Page
}

// Bin manages slabs for a single object size class
type Bin struct {
	mtx       sync.Mutex
	size      uint64
	align     uint64
	exhausted []*Slab
	available []*Slab // multiple available slabs (LIFO)
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
		freePages: make([]*res.Page, 0),
	}

	for i := range a.bins {
		a.bins[i] = &Bin{
			size:      SIZE_CLASSES[i],
			align:     8,
			exhausted: make([]*Slab, 0),
			available: make([]*Slab, 0),
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
	// Explicitly search for the size class instead of calculating the index
	// This ensures we only return valid indices for sizes actually in SIZE_CLASSES
	for i, s := range SIZE_CLASSES {
		if s == classSize {
			return i
		}
	}
	// If size not found in SIZE_CLASSES, return NUM_CLASSES to indicate direct allocation
	return NUM_CLASSES
}

func (a *SlabAllocator) Alloc(size, align uint64) unsafe.Pointer {
	cs := a.findSizeClass(size, align)
	id := indexForSize(cs)

	if id >= NUM_CLASSES {
		return a.res.Alloc(cs, align)
	}

	bin := a.bins[id]
	return a.binAlloc(bin)
}

// AllocatePage returns a page for a bin, either from the free pool or by allocating a new one
func (a *SlabAllocator) AllocatePage(size uint64) *res.Page {
	// Calculate the needed size with proper logic for different size ranges
	var roundedSize int

	if size <= uint64(res.PAGE_SIZE) {
		// For small sizes: fit within one page boundary
		// Calculate how many objects fit in one page
		objectsPerPage := max(1, res.PAGE_SIZE/int(size))
		chunkSize := objectsPerPage * int(size)
		// Round up to page boundary
		roundedSize = ((chunkSize + res.PAGE_SIZE - 1) / res.PAGE_SIZE) * res.PAGE_SIZE
	} else {
		// For sizes >= PAGE_SIZE: allocate in multiples of page size
		// Round size up to nearest multiple of PAGE_SIZE
		roundedSize = int(((size + uint64(res.PAGE_SIZE) - 1) / uint64(res.PAGE_SIZE)) * uint64(res.PAGE_SIZE))
	}

	a.mtx.Lock()
	// Try to find a free page that's large enough
	for i := len(a.freePages) - 1; i >= 0; i-- {
		page := a.freePages[i]
		if len(page.Base()) >= roundedSize {
			// Found a suitable page, remove from pool
			a.freePages = append(a.freePages[:i], a.freePages[i+1:]...)
			a.mtx.Unlock()
			// Return reused page directly - initSlab will rebuild free list from scratch
			// No need to zero the page (expensive and unnecessary)
			return page
		}
	}
	a.mtx.Unlock()

	// No suitable page in free pool, allocate new one
	// Create a new page with direct mmap allocation to avoid sharing
	// This prevents multiple slabs from interfering with each other
	page := a.res.New(int(roundedSize))
	if page == nil {
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
		if bin != nil && a.binOwns(bin, ptr) {
			a.binFree(bin, ptr)
			return
		}
	}
}

func (a *SlabAllocator) binOwns(b *Bin, ptr unsafe.Pointer) bool {
	if ptr == nil {
		return false
	}
	addr := uintptr(ptr)

	for _, s := range b.available {
		if a.slabContains(s, addr) {
			return true
		}
	}

	for _, s := range b.exhausted {
		if a.slabContains(s, addr) {
			return true
		}
	}

	return false
}

func (a *SlabAllocator) slabContains(s *Slab, addr uintptr) bool {
	if s.page == nil {
		return false
	}
	var (
		start = uintptr(unsafe.Pointer(unsafe.SliceData(s.page.Base())))
		end   = start + uintptr(len(s.page.Base()))
	)
	return addr >= start && addr < end
}

func (a *SlabAllocator) removeFromSlice(s *[]*Slab, target *Slab) {
	for i, slab := range *s {
		if slab == target {
			*s = append((*s)[:i], (*s)[i+1:]...)
			return
		}
	}
}

func (a *SlabAllocator) binAlloc(b *Bin) unsafe.Pointer {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	// Fast path: most recent available slab
	if len(b.available) > 0 {
		s := b.available[len(b.available)-1]
		ptr := a.allocFromSlab(s)
		if ptr != nil {
			return ptr
		}
		// Slab exhausted, move to exhausted list and allocate new slab
		b.available = b.available[:len(b.available)-1]
		b.exhausted = append(b.exhausted, s)
	}

	// Try free page pool via allocator
	page := a.AllocatePage(b.size)
	if page == nil {
		return nil
	}

	s := a.initSlab(b, page)
	b.available = append(b.available, s)
	return a.allocFromSlab(s)
}

func (a *SlabAllocator) initSlab(b *Bin, page *res.Page) *Slab {
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

func (a *SlabAllocator) allocFromSlab(s *Slab) unsafe.Pointer {
	if s.freed == nil {
		return nil
	}
	ptr := s.freed
	s.freed = *(*unsafe.Pointer)(ptr)
	s.usable = s.usable - 1
	return ptr
}

func (a *SlabAllocator) binFree(b *Bin, ptr unsafe.Pointer) {
	b.mtx.Lock()
	defer b.mtx.Unlock()

	s := a.findSlab(b, ptr)
	if s == nil {
		return
	}

	// Prepend to in-object free list
	*(*unsafe.Pointer)(ptr) = s.freed
	s.freed = ptr
	s.usable = s.usable + 1

	if s.usable == s.capacity {
		// Completely empty → return page to free pool
		a.removeFromSlice(&b.available, s)
		a.removeFromSlice(&b.exhausted, s)

		a.freePages = append(a.freePages, s.page)
	} else {
		// Promote to available list
		a.removeFromSlice(&b.exhausted, s)
		b.available = append(b.available, s)
	}
}

func (a *SlabAllocator) findSlab(b *Bin, ptr unsafe.Pointer) *Slab {
	addr := uintptr(ptr)

	// Check available list first
	for _, s := range b.available {
		if a.slabContains(s, addr) {
			return s
		}
	}

	// Check exhausted list
	for _, s := range b.exhausted {
		if a.slabContains(s, addr) {
			return s
		}
	}

	return nil
}

func (a *SlabAllocator) Reset() {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	// Lock all bins first to prevent concurrent access during reset
	for _, bin := range a.bins {
		if bin != nil {
			bin.mtx.Lock()
		}
	}

	// Now reset under lock
	if a.res != nil {
		a.res.Reset()
	}

	// Clear free pages pool
	a.freePages = nil

	// Clear all bin slab lists
	for _, bin := range a.bins {
		if bin != nil {
			bin.exhausted = nil
			bin.available = nil
		}
	}

	// Unlock all bins
	for _, bin := range a.bins {
		if bin != nil {
			bin.mtx.Unlock()
		}
	}
}

func (a *SlabAllocator) Delete() {
	a.mtx.Lock()
	defer a.mtx.Unlock()

	// Lock all bins first to prevent concurrent access during delete
	for _, bin := range a.bins {
		if bin != nil {
			bin.mtx.Lock()
		}
	}

	// Delete resource manager
	if a.res != nil {
		a.res.Delete()
		a.res = nil
	}

	// Clear free pages pool and bins
	a.freePages = nil

	for i := range a.bins {
		a.bins[i] = nil
	}

	// Unlock all bins (though they won't be used after this)
	for _, bin := range a.bins {
		if bin != nil {
			bin.mtx.Unlock()
		}
	}
}

func (a *SlabAllocator) Remove(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}

	// Find which bin owns this pointer and free it
	for i := range a.bins {
		bin := a.bins[i]
		if bin == nil {
			continue
		}

		// Check if slab is in this bin
		bin.mtx.Lock()
		slab := a.findSlab(bin, ptr)
		bin.mtx.Unlock()

		if slab != nil {
			a.binFree(bin, ptr)
			return
		}
	}
}

func (a *SlabAllocator) Owns(ptr unsafe.Pointer) bool {
	if ptr == nil || a.res == nil {
		return false
	}
	return a.res.Owns(ptr)
}
