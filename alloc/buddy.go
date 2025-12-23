// (moved misplaced methods below)
// buddy_per_chunk_bitmap.go
package alloc

import (
	"sync"
	"unsafe"

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
	MIN_BITMAP_SIZE = res.PAGE_SIZE
	MIN_DATA_SIZE   = MIN_BITMAP_SIZE * 8 * MIN_BLOCK_SIZE
	MIN_CHUNK_SIZE  = 64 * 1024 // 64KB - reasonable default, was (MIN_DATA_SIZE + MIN_BITMAP_SIZE) which is too large
	USE_RES         = true      // Set to true to use Res manager, false for direct make()
)

// Chunk wraps a byte slice with cached address bounds and size information
// Uses a complete binary tree with bitset metadata stored at chunk start
type Chunk struct {
	data         []byte  // underlying byte slice with bitset header + payload
	startAddress uintptr // start address
	endAddress   uintptr // end address
	blockSize    int     // data/block size (usable payload, after bitset)
	order        int     // tree order (height)
	bitsSize     int     // size of bitset in bytes
}

func (c *Chunk) Base() []byte {
	return c.data
}

// PayloadOffset returns the offset where payload begins after bitset
func (c *Chunk) PayloadOffset() int {
	return c.bitsSize
}

// SetBit sets a bit in the bitset at given position
func (c *Chunk) SetBit(idx int) {
	var (
		byteIdx = idx / 8
		bitIdx  = uint(idx % 8)
	)
	if byteIdx < c.bitsSize {
		c.data[byteIdx] |= (1 << bitIdx)
	}
}

// ClearBit clears a bit in the bitset at given position
func (c *Chunk) ClearBit(idx int) {
	var (
		byteIdx = idx / 8
		bitIdx  = uint(idx % 8)
	)
	if byteIdx < c.bitsSize {
		c.data[byteIdx] &= ^(1 << bitIdx)
	}
}

// GetBit gets a bit from the bitset at given position
func (c *Chunk) GetBit(idx int) bool {
	var (
		byteIdx = idx / 8
		bitIdx  = uint(idx % 8)
	)
	if byteIdx >= c.bitsSize {
		return false
	}
	return (c.data[byteIdx] & (1 << bitIdx)) != 0
}

// Owns checks if a pointer falls within this chunk's address bounds
func (c *Chunk) Owns(ptr unsafe.Pointer) bool {
	if ptr == nil {
		return false
	}
	return uintptr(ptr) >= c.startAddress && uintptr(ptr) < c.endAddress
}

// TreeIndex calculates the tree index for a given offset in the payload
func (c *Chunk) TreeIndex(offset int) int {
	return (1 << uint(c.order)) + (offset / MIN_BLOCK_SIZE)
}

// OffsetForIndex calculates the payload offset for a given tree leaf index
func (c *Chunk) OffsetForIndex(idx int) int {
	return (idx - (1 << uint(c.order))) * MIN_BLOCK_SIZE
}

// UpdateNode updates a node and propagates changes up to root
func (c *Chunk) UpdateNode(idx int) {
	if idx <= 0 || idx >= (1<<uint(c.order+1)) {
		return
	}

	// For leaf nodes, don't update
	if idx >= (1 << uint(c.order)) {
		return
	}

	var (
		leftFree  = c.GetBit((2 * idx) + 0)
		rightFree = c.GetBit((2 * idx) + 1)
	)

	// Parent is free only if both children are free
	if leftFree && rightFree {
		c.SetBit(idx)
	} else {
		c.ClearBit(idx)
	}

	// Propagate up
	if idx > 0 {
		c.UpdateNode(idx / 2)
	}
}

// Allocate allocates a block of given size from this chunk's payload
// Uses tree-based search with bitset encoding for space efficiency
// Returns a pointer to the allocated block, or nil if no suitable free block exists
func (c *Chunk) Allocate(size int) unsafe.Pointer {
	if size <= 0 || size > c.blockSize {
		return nil
	}
	if size < MIN_BLOCK_SIZE {
		size = MIN_BLOCK_SIZE
	}
	if !c.GetBit(1) {
		return nil
	}
	var (
		idx    = c.FindFreeNode(1, size)
		parent = idx / 2
	)
	if idx < 0 {
		return nil
	}
	c.ClearBit(idx)
	if parent >= 1 {
		c.UpdateNode(parent)
	}
	return unsafe.Add(unsafe.Pointer(unsafe.SliceData(c.data)), c.PayloadOffset()+c.OffsetForIndex(idx))
}

// FindFreeNode recursively searches the tree for a node that can fit the requested size
func (c *Chunk) FindFreeNode(idx int, size int) int {
	maxIdx := 1 << uint(c.order+1)
	if idx < 0 || idx >= maxIdx {
		return -1
	}

	// Check if this node is free
	if !c.GetBit(idx) {
		return -1 // Node is fully allocated
	}

	// If this is a leaf node and it's free, use it
	if idx >= (1 << uint(c.order)) {
		return idx
	}

	// For internal nodes, try left child first (prefer contiguous allocation)
	var (
		left  = (2 * idx) + 0
		right = (2 * idx) + 1
	)

	if left < maxIdx && c.GetBit(left) {
		result := c.FindFreeNode(left, size)
		if result >= 0 {
			return result
		}
	}

	if right < maxIdx && c.GetBit(right) {
		return c.FindFreeNode(right, size)
	}

	return -1
}

// Remove deallocates a pointer within this chunk's payload
// Returns true if the pointer was successfully deallocated, false if it doesn't belong to this chunk
func (c *Chunk) Remove(ptr unsafe.Pointer) bool {
	if !c.Owns(ptr) {
		return false
	}
	offset := int(uintptr(ptr)-c.startAddress) - c.PayloadOffset()
	if offset < 0 || offset >= c.blockSize {
		return false
	}
	var (
		treeIdx = c.TreeIndex(offset)
		maxIdx  = 1 << uint(c.order+1)
	)
	if treeIdx < 0 || treeIdx >= maxIdx {
		return false
	}
	c.SetBit(treeIdx)
	parent := treeIdx / 2
	for parent >= 1 {
		c.UpdateNode(parent)
		parent = parent / 2
	}
	return true
}

// Reset resets all allocations in the chunk
func (c *Chunk) Reset() {
	// Set all bits to 1 (free state)
	for i := 0; i < c.bitsSize; i++ {
		c.data[i] = 0xFF
	}
}

// GetMaxFreeSize returns the maximum allocation size available in this chunk
func (c *Chunk) GetMaxFreeSize() int {
	if c.GetBit(1) {
		return c.blockSize
	}
	return 0
}

// ============================================================================
// BuddyAllocator
// ============================================================================

type BuddyAllocator struct {
	chunks       []*Chunk
	lastChunkIdx int      // Index of last successfully allocated chunk
	res          *res.Res // Used when USE_RES = true
	order        int
	mtx          sync.Mutex
}

func (a *BuddyAllocator) Alloc(size, align uint64) unsafe.Pointer {
	if size == 0 {
		return nil
	}
	if align == 0 {
		align = 8
	}
	align = res.RoundPow2(align)
	blockSize := max(res.RoundPow2(size), align)
	blockSize = max(blockSize, MIN_BLOCK_SIZE)
	a.mtx.Lock()
	defer a.mtx.Unlock()
	if a.lastChunkIdx >= 0 && a.lastChunkIdx < len(a.chunks) {
		chunk := a.chunks[a.lastChunkIdx]
		if (1<<uint(chunk.order)) >= int(size) && chunk.GetMaxFreeSize() > 0 {
			if blockPtr := chunk.Allocate(int(size)); blockPtr != nil {
				return blockPtr
			}
		}
	}
	type chunkEntry struct {
		idx      int
		freeSize int
	}
	entries := make([]chunkEntry, 0, len(a.chunks))
	for i, chunk := range a.chunks {
		if (1 << uint(chunk.order)) >= int(size) {
			if freeSize := chunk.GetMaxFreeSize(); freeSize > 0 {
				entries = append(entries, chunkEntry{i, freeSize})
			}
		}
	}
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[j].freeSize > entries[i].freeSize {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}
	for _, entry := range entries {
		if blockPtr := a.chunks[entry.idx].Allocate(int(size)); blockPtr != nil {
			a.lastChunkIdx = entry.idx
			return blockPtr
		}
	}
	growSize := max(max(int(res.RoundPow2(size*2)), int(blockSize)), MIN_CHUNK_SIZE)
	if growSize > MIN_CHUNK_SIZE {
		if rem := growSize % MIN_CHUNK_SIZE; rem != 0 {
			growSize += MIN_CHUNK_SIZE - rem
		}
	}
	chunk := NewChunk(a.res, growSize)
	a.chunks = append(a.chunks, chunk)
	a.lastChunkIdx = len(a.chunks) - 1
	return chunk.Allocate(int(size))
}

func (a *BuddyAllocator) Remove(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}

	a.mtx.Lock()
	defer a.mtx.Unlock()

	// Find the chunk that owns this pointer and remove from it
	for _, chunk := range a.chunks {
		if chunk.Remove(ptr) {
			return
		}
	}
}
func (a *BuddyAllocator) Reset() {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	a.order = 0
	for _, chunk := range a.chunks {
		chunk.Reset()
		if maxFree := chunk.GetMaxFreeSize(); maxFree >= MIN_BLOCK_SIZE {
			if order := int(res.Log2(uint64(maxFree))); order > a.order {
				a.order = order
			}
		}
	}
}

func (a *BuddyAllocator) Delete() {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	if USE_RES {
		if a.res != nil {
			a.res.Delete()
			a.res = nil
		}
	} else {
		for _, chunk := range a.chunks {
			res.ReleasePages(chunk.data)
		}
	}
	a.chunks = nil
	a.order = 0
}

func (a *BuddyAllocator) Owns(ptr unsafe.Pointer) bool {
	if ptr == nil {
		return false
	}

	a.mtx.Lock()
	defer a.mtx.Unlock()

	// Use chunk's Owns method for address range check
	for _, chunk := range a.chunks {
		if chunk.Owns(ptr) {
			return true
		}
	}
	return false
}

// ============================================================================
// Utility Functions
// ============================================================================

// NewChunk creates a new chunk of given size with bitset metadata at the start
func NewChunk(r *res.Res, size int) *Chunk {
	var data []byte
	if r != nil {
		page := r.New(size)
		data = page.Base()
	} else {
		data = make([]byte, size)
	}

	numLeaves := size / MIN_BLOCK_SIZE

	// Calculate tree depth (order) for the number of leaves
	order := 0
	for (1 << uint(order)) < numLeaves {
		order++
	}

	// Total nodes in tree: 2^(order+1) - 1
	totalNodes := (1 << uint(order+1)) - 1
	// Bitset size in bytes (1 bit per node, rounded up to byte boundary)
	bitsSize := (totalNodes + 7) / 8

	// Adjust size to align bitset to reasonable boundary
	if bitsSize%8 != 0 {
		bitsSize = ((bitsSize / 8) + 1) * 8
	}

	// Check if metadata fits in the data
	if bitsSize >= size {
		// Fallback: reduce order to fit bitset
		order = 0
		for order < 60 {
			totalNodes = (1 << uint(order+1)) - 1
			testBitsSize := (totalNodes + 7) / 8
			if testBitsSize%8 != 0 {
				testBitsSize = ((testBitsSize / 8) + 1) * 8
			}
			testBlockSize := size - testBitsSize
			if testBlockSize >= MIN_BLOCK_SIZE {
				bitsSize = testBitsSize
				break
			}
			order++
		}
	}

	// Initialize bitset in data buffer
	for i := 0; i < bitsSize && i < len(data); i++ {
		data[i] = 0
	}
	blockSize := size - bitsSize
	leafStart := 1 << uint(order)
	for i := leafStart; i < (1 << uint(order+1)); i++ {
		byteIdx := i / 8
		bitIdx := uint(i % 8)
		if byteIdx < bitsSize {
			data[byteIdx] |= (1 << bitIdx)
		}
	}
	for i := (leafStart - 1); i >= 1; i-- {
		left := 2 * i
		right := 2*i + 1
		leftBit := left < (1<<uint(order+1)) && (left/8 < bitsSize) && (data[left/8]&(1<<uint(left%8))) != 0
		rightBit := right < (1<<uint(order+1)) && (right/8 < bitsSize) && (data[right/8]&(1<<uint(right%8))) != 0
		if leftBit && rightBit {
			byteIdx := i / 8
			bitIdx := uint(i % 8)
			if byteIdx < bitsSize {
				data[byteIdx] |= (1 << bitIdx)
			}
		}
	}
	startAddress := uintptr(unsafe.Pointer(unsafe.SliceData(data)))
	return &Chunk{
		data:         data,
		startAddress: startAddress,
		endAddress:   startAddress + uintptr(len(data)),
		blockSize:    blockSize,
		order:        order,
		bitsSize:     bitsSize,
	}
}

func NewBuddyAllocator() *BuddyAllocator {
	// Initialize allocator without pre-allocating chunks
	// First Alloc() call will size the chunk appropriately based on the allocation size
	a := &BuddyAllocator{
		chunks: []*Chunk{},
	}
	if USE_RES {
		a.res = res.NewRes(MIN_CHUNK_SIZE) // res manager still initialized but not used until needed
	}
	return a
}
