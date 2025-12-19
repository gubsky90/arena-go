// (moved misplaced methods below)
// buddy_per_chunk_bitmap.go
package alloc

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/thebagchi/arena-go/alloc/cont"
	"github.com/thebagchi/arena-go/res"
)

/*
Buddy Allocator with Per-Chunk Bitmap Metadata
================================================================================

Each chunk (page) layout:
┌──────────────────────────────────────┐  ← chunk.Base()
│ Buddy Bitmap (payload_size/16 bits)  │  ← 1 bit per 16-byte min block
│ e.g., 64KB chunk → ~512 bytes bitmap │
├──────────────────────────────────────┤  ← payloadStart = base + bitmapSize
│ Buddy blocks (16B, 32B, 64B, ...)    │
│ - Power-of-2 sized and aligned       │
│ - Free blocks tracked in global lists│
└──────────────────────────────────────┘

Free lists: global []freeList holding raw pointers into payload area
Bitmap: used only to check "is buddy free?" during Remove() → fast merge

Advantages:
- Metadata safe from user overflow
- Excellent cache locality
- Fast buddy check (bit test)
- Low overhead (~0.8% per page)
- Full coalescing
- User gets exact requested size
================================================================================
*/

const (
	MIN_ORDER      = 4 // 16-byte minimum block
	MIN_BLOCK_SIZE = 1 << MIN_ORDER
)

var (
	MIN_CHUNK_SIZE = (res.PAGE_SIZE * 8 * MIN_BLOCK_SIZE) + res.PAGE_SIZE
)

type FreeList = cont.List[unsafe.Pointer]

// Bitmap is a bit vector for tracking allocation state in a chunk.
type Bitmap struct {
	data []byte // underlying bytes
	bits int    // number of valid bits
}

// NewBitmap creates a Bitmap view over the in-place bitmap at the start of a chunk.
func NewBitmap(chunk []byte) *Bitmap {
	chunkSize := len(chunk)
	numBits := (chunkSize - bitmapSize(chunkSize)) / MIN_BLOCK_SIZE
	return &Bitmap{
		data: chunk[:bitmapSize(chunkSize)],
		bits: numBits,
	}
}

// Bitmap methods
func (b *Bitmap) IsSet(idx int) bool {
	if idx < 0 || idx >= b.bits {
		return false
	}
	byteIdx := idx / 8
	bitIdx := uint(idx % 8)
	return b.data[byteIdx]&(1<<bitIdx) != 0
}

func (b *Bitmap) Set(idx int) {
	if idx < 0 || idx >= b.bits {
		return
	}
	byteIdx := idx / 8
	bitIdx := uint(idx % 8)
	b.data[byteIdx] |= (1 << bitIdx)
}

func (b *Bitmap) Clear(idx int) {
	if idx < 0 || idx >= b.bits {
		return
	}
	byteIdx := idx / 8
	bitIdx := uint(idx % 8)
	b.data[byteIdx] &= ^(1 << bitIdx)
}

func (b *Bitmap) Reset() {
	for i := range b.data {
		b.data[i] = 0
	}
}

type BuddyAllocator struct {
	res   *res.Res
	free  []*FreeList
	order int
	mtx   sync.Mutex
}

func NewBuddyAllocator() *BuddyAllocator {
	a := &BuddyAllocator{
		res:  res.NewRes(MIN_CHUNK_SIZE),
		free: make([]*FreeList, 0, 32),
	}
	a.mtx.Lock()
	a.initializeFreeListList()
	a.mtx.Unlock()
	return a
}

// bitmapSize returns bytes needed for bitmap covering the full chunk
func bitmapSize(chunkSize int) int {
	num := chunkSize / MIN_BLOCK_SIZE
	return (num + 7) / 8 // round up to nearest byte
}

// blockOffset returns first usable byte after the bitmap
func blockOffset(chunk []byte) unsafe.Pointer {
	size := bitmapSize(len(chunk))
	return unsafe.Add(unsafe.Pointer(unsafe.SliceData(chunk)), size)
}

// blockSize returns usable memory after bitmap
func blockSize(chunk []byte) int {
	return len(chunk) - bitmapSize(len(chunk))
}

// blockIndex returns the min-block index within the payload for a pointer
func blockIndex(ptr uintptr, chunk []byte) int {
	offset := int(ptr - uintptr(blockOffset(chunk)))
	if offset < 0 || offset >= blockSize(chunk) {
		return -1
	}
	return offset / MIN_BLOCK_SIZE
}

// registerAllChunksAsFree treats every chunk as one large free buddy block
func (a *BuddyAllocator) initializeFreeListList() {
	for i := 0; i < len(a.free); i++ {
		fl := a.free[i]
		if fl != nil && !fl.IsEmpty() {
			fl.Reset()
		}
	}
	a.free = a.free[:0]
	a.order = 0
	for _, chunk := range a.res.Chunks() {
		usable := blockSize(chunk.Base())
		if usable < MIN_BLOCK_SIZE {
			continue
		}
		order := int(res.Log2(uint64(usable)))
		if res.Pow2(uint64(order+1)) <= uint64(usable) {
			order++
		}
		if order < MIN_ORDER {
			continue
		}
		payload := blockOffset(chunk.Base())
		if order > a.order {
			a.order = order
		}
		list := a.getFreeList(order)
		list.PushBack(payload)
	}
}

func (a *BuddyAllocator) Alloc(size, align uint64) unsafe.Pointer {
	if size == 0 {
		return nil
	}
	if align == 0 {
		align = 8
	}

	align = res.RoundPow2(align)
	var (
		blockSize     = max(res.RoundPow2(size), align)
		requiredOrder = int(res.Log2(blockSize))
	)
	a.mtx.Lock()
	defer a.mtx.Unlock()
	order := requiredOrder
	for order <= a.order || order <= 60 {
		list := a.getFreeList(order)
		for node := list.Back(); node != nil; node = node.Prev {
			blockPtr := node.Value
			// Check bitmap for this block
			for _, chunk := range a.res.Chunks() {
				var (
					baseAddr    = uintptr(unsafe.Pointer(unsafe.SliceData(chunk.Base())))
					payloadAddr = uintptr(blockOffset(chunk.Base()))
					blockAddr   = uintptr(blockPtr)
				)
				if blockAddr >= payloadAddr && blockAddr < baseAddr+uintptr(len(chunk.Base())) {
					idx := blockIndex(uintptr(blockPtr), chunk.Base())
					if idx >= 0 {
						bitmap := NewBitmap(chunk.Base())
						if !bitmap.IsSet(idx) {
							// Remove from free list and use this block
							list.Remove(node)
							// Split if needed
							splitOrder := order
							for splitOrder > requiredOrder {
								a.split(splitOrder, blockPtr)
								splitOrder--
							}
							a.markAllocated(blockPtr)
							return blockPtr
						}
					}
					break
				}
			}
		}
		order++
	}

	// Grow
	growSize := max(max(int(res.RoundPow2(size*2)), int(blockSize)), MIN_CHUNK_SIZE)
	if growSize > MIN_CHUNK_SIZE {
		rem := growSize % MIN_CHUNK_SIZE
		if rem != 0 {
			growSize += MIN_CHUNK_SIZE - rem
		}
	}

	a.res.Alloc(uint64(growSize), align)
	a.initializeFreeListList()

	// Final attempt
	order = requiredOrder
	for order <= a.order {
		list := a.getFreeList(order)
		if !list.IsEmpty() {
			blockPtr, ok := list.PopBack()
			if ok {
				for order > requiredOrder {
					a.split(order, blockPtr)
					order--
				}
				a.markAllocated(blockPtr)
				return blockPtr
			}
		}
		order++
	}

	return nil
}

func (a *BuddyAllocator) split(order int, blockPtr unsafe.Pointer) {
	var (
		offset   = res.Pow2(uint64(order - 1))
		buddyPtr = unsafe.Add(blockPtr, int(offset))
		list     = a.getFreeList(order - 1)
	)
	list.PushBack(blockPtr)
	list.PushBack(buddyPtr)
}

func (a *BuddyAllocator) Remove(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}

	a.mtx.Lock()
	defer a.mtx.Unlock()

	addr := uintptr(ptr)
	chunk, found := a.res.FindPage(ptr)
	if !found || chunk == nil {
		return
	}

	var (
		bitmap = NewBitmap(chunk.Base())
		offset = uintptr(blockOffset(chunk.Base()))
		idx    = blockIndex(uintptr(ptr), chunk.Base())
	)
	if idx < 0 {
		return
	}
	bitmap.Clear(idx)

	var (
		ptrOffset = int(addr - offset)
		order     = int(res.Log2(uint64(ptrOffset + MIN_BLOCK_SIZE)))
	)
	for res.Pow2(uint64(order+1)) <= uint64(ptrOffset+int(res.Pow2(uint64(order)))) {
		order++
	}

	blockPtr := ptr
	for {
		blockSize := uintptr(res.Pow2(uint64(order)))
		blockBase := addr &^ (blockSize - 1)
		buddyAddr := blockBase
		if addr == blockBase {
			buddyAddr += blockSize
		} else {
			buddyAddr -= blockSize
		}

		buddyIdx := blockIndex(uintptr(buddyAddr), chunk.Base())
		// Only merge if buddy is free (bitmap clear) and buddy is in the free list
		if buddyIdx < 0 || bitmap.IsSet(buddyIdx) {
			break
		}

		list := a.getFreeList(order)
		foundBuddy := false
		for node := list.Front(); node != nil; node = node.Next {
			if uintptr(node.Value) == buddyAddr {
				list.Remove(node)
				foundBuddy = true
				break
			}
		}
		if !foundBuddy {
			break // buddy not actually free in list
		}

		// Clear buddy's bit since it is being merged
		bitmap.Clear(buddyIdx)

		if addr > buddyAddr {
			addr = buddyAddr
			// compute pointer relative to payload start to avoid uintptr->unsafe.Pointer conversion
			offset := int(buddyAddr - offset)
			blockPtr = unsafe.Add(blockOffset(chunk.Base()), offset)
		}
		order++
	}
	list := a.getFreeList(order)
	list.PushBack(blockPtr)
}

func (a *BuddyAllocator) Reset() {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.res.Reset()
	a.initializeFreeListList()
}

func (a *BuddyAllocator) Delete() {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	if a.res != nil {
		a.res.Delete()
		a.res = nil
	}
	a.free = nil
	a.order = 0
}

func (a *BuddyAllocator) Owns(ptr unsafe.Pointer) bool {
	if ptr == nil || a.res == nil {
		return false
	}

	a.mtx.Lock()
	defer a.mtx.Unlock()

	return a.res.Owns(ptr)
}

func (a *BuddyAllocator) getFreeList(order int) *FreeList {
	idx := order - MIN_ORDER
	if idx < 0 {
		panic("invalid buddy order")
	}
	if idx >= len(a.free) {
		// Grow the slice
		newSize := idx + 16
		a.free = append(a.free, make([]*FreeList, newSize-len(a.free))...)
	}
	fl := a.free[idx]
	if fl == nil {
		fl = cont.NewList[unsafe.Pointer]()
		a.free[idx] = fl
	}
	return fl
}

func (a *BuddyAllocator) markAllocated(blockPtr unsafe.Pointer) {
	for _, chunk := range a.res.Chunks() {
		var (
			baseAddr    = uintptr(unsafe.Pointer(unsafe.SliceData(chunk.Base())))
			payloadAddr = uintptr(blockOffset(chunk.Base()))
			blockAddr   = uintptr(blockPtr)
		)
		if blockAddr >= payloadAddr && blockAddr < baseAddr+uintptr(len(chunk.Base())) {
			idx := blockIndex(uintptr(blockPtr), chunk.Base())
			if idx >= 0 {
				bitmap := NewBitmap(chunk.Base())
				bitmap.Set(idx)
			}
			break
		}
	}
}

func (a *BuddyAllocator) DebugDump() {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	fmt.Println("=== Buddy Allocator Free Lists ===")
	for i := 0; i < len(a.free); i++ {
		fl := a.free[i]
		if fl == nil || fl.IsEmpty() {
			continue
		}
		var (
			order = i + MIN_ORDER
			size  = 1 << order
		)
		fmt.Printf("Order %2d | Size: %8d bytes | Count: %d\n", order, size, fl.Length())
		it := fl.Iter()
		for ptr, ok := it.Next(); ok; ptr, ok = it.Next() {
			fmt.Printf("  → %p\n", ptr)
		}
	}
	fmt.Println("==================================")
}
