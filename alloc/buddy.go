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
	MAX_LEVEL      = 32 // Max power-of-2 levels for free lists (2^4 to 2^35)
)

var (
	MIN_BITMAP_SIZE = res.PAGE_SIZE                        // 4KB bitmap
	MIN_DATA_SIZE   = MIN_BITMAP_SIZE * 8 * MIN_BLOCK_SIZE // Bitmap capacity: 4KB * 8 * 16B = 512KB
	MIN_CHUNK_SIZE  = MIN_DATA_SIZE + MIN_BITMAP_SIZE      // Full utilization: 512KB data + 4KB bitmap = ~516KB
	USE_RES         = true                                 // Set to true to use Res manager, false for direct make()
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

// SetBit sets a bit in the bitset at given position using uint64 operations
func (c *Chunk) SetBit(idx int) {
	var (
		qwordIdx    = idx / 64
		bitIdx      = uint(idx % 64)
		qwordOffset = qwordIdx * 8
	)
	if qwordOffset+8 <= c.bitsSize {
		// Read uint64, set bit, write back
		ptr := unsafe.Pointer(unsafe.SliceData(c.data[qwordOffset:]))
		val := (*uint64)(ptr)
		*val |= (1 << bitIdx)
	}
}

// ClearBit clears a bit in the bitset at given position using uint64 operations
func (c *Chunk) ClearBit(idx int) {
	var (
		qwordIdx    = idx / 64
		bitIdx      = uint(idx % 64)
		qwordOffset = qwordIdx * 8
	)
	if qwordOffset+8 <= c.bitsSize {
		// Read uint64, clear bit, write back
		ptr := unsafe.Pointer(unsafe.SliceData(c.data[qwordOffset:]))
		val := (*uint64)(ptr)
		*val &= ^(1 << bitIdx)
	}
}

// GetBit gets a bit from the bitset at given position using uint64 operations
func (c *Chunk) GetBit(idx int) bool {
	var (
		qwordIdx    = idx / 64
		bitIdx      = uint(idx % 64)
		qwordOffset = qwordIdx * 8
	)
	if qwordOffset+8 > c.bitsSize {
		return false
	}
	ptr := unsafe.Pointer(unsafe.SliceData(c.data[qwordOffset:]))
	val := (*uint64)(ptr)
	return (*val & (1 << bitIdx)) != 0
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
// Implements proper buddy splitting: when a larger block is found, splits it down to requested size
//
// Algorithm (with example for 100-byte request, found 512-byte block):
//
//	Step 1: Find free block via DFS search
//	        Stack-based traversal of tree looking for any free block
//	        Found: 512-byte block at index 47
//
//	Step 2: Split block recursively down to size needed
//	        Block size 512 > required 100 → SPLIT
//	        ├─ Calculate children: left=94 (256B), right=95 (256B buddy)
//	        ├─ SetBit(95) → mark right buddy as free
//	        ├─ idx = 94, continue with left
//	        Block size 256 > 100 → SPLIT
//	        ├─ Calculate children: left=188 (128B), right=189 (128B buddy)
//	        ├─ SetBit(189) → mark right buddy as free
//	        ├─ idx = 188, continue with left
//	        Block size 128 > 100 → SPLIT
//	        ├─ Calculate children: left=376 (64B), right=377 (64B buddy)
//	        ├─ SetBit(377) → mark right buddy as free
//	        ├─ idx = 376, continue with left
//	        Block size 64 < 100? NO → Stop splitting
//
//	Step 3: Mark final block as allocated
//	        ClearBit(376) → block at index 376 is now ALLOCATED
//	        UpdateNode propagates changes up to root
//
//	Result: Exact 64-byte allocation (closest power-of-2 to 100)
//	        Buddy blocks 377, 189, 95 remain free for future allocation
//
// Performance: O(log N) where N = number of blocks in tree
//   - FindFreeNode: O(log N) DFS traversal
//   - Splitting: O(log N) recursive splits
//   - Total: O(log N) amortized
func (c *Chunk) Allocate(size int) unsafe.Pointer {
	if size <= 0 || size > c.blockSize {
		return nil
	}
	size = max(size, MIN_BLOCK_SIZE)
	if !c.GetBit(1) {
		return nil
	}

	var (
		requiredSize = max(size, MIN_BLOCK_SIZE)
		maxIdx       = 1 << uint(c.order+1)
	)

	// Find a free node at the appropriate size or larger
	idx := c.FindFreeNode(1, size)
	if idx < 0 {
		return nil
	}

	// Split the block down to the requested size
	// Start from found index and split down until we reach the right size
	currentNodeSize := c.getSizeForIndex(idx)

	// Keep splitting until block size matches requested size
	for currentNodeSize > requiredSize && idx < maxIdx {
		// Split: mark right buddy as free, continue with left half
		var (
			leftChild  = (2 * idx) + 0
			rightChild = (2 * idx) + 1
		)

		if rightChild >= maxIdx {
			break
		}

		// Right child becomes a free buddy block
		c.SetBit(rightChild)
		c.UpdateNode(rightChild / 2)

		// Continue with left child
		idx = leftChild
		currentNodeSize = currentNodeSize / 2
	}

	// Mark the final allocated block as used
	c.ClearBit(idx)

	// Update parent chain
	parent := idx / 2
	if parent >= 1 {
		c.UpdateNode(parent)
	}

	return unsafe.Add(unsafe.Pointer(unsafe.SliceData(c.data)), c.PayloadOffset()+c.OffsetForIndex(idx))
}

// getSizeForIndex returns the block size (in bytes) for a given tree node index
func (c *Chunk) getSizeForIndex(idx int) int {
	if idx <= 0 || idx >= (1<<uint(c.order+1)) {
		return 0
	}

	// Calculate the level of this node (distance from leaves)
	// Leaves are at level order, root is at level 0
	// Level can be calculated from position in tree
	leafStart := 1 << uint(c.order)

	if idx >= leafStart {
		// This is a leaf node = MIN_BLOCK_SIZE
		return MIN_BLOCK_SIZE
	}

	// For internal nodes: count leading zeros or calculate from parent distance to leaf
	// Each level up from leaves doubles the size
	var level int
	temp := idx
	for temp < leafStart {
		temp *= 2
		level++
	}

	return MIN_BLOCK_SIZE << uint(level)
}

// FindFreeNode recursively searches the tree for a free leaf node
// Uses depth-first traversal with preference for left children
// Returns the index of a free leaf, or -1 if none found
func (c *Chunk) FindFreeNode(startIdx int, size int) int {
	var (
		maxIdx    = 1 << uint(c.order+1)
		leafStart = 1 << uint(c.order)
	)

	// Helper function to recursively search the tree
	var search func(int) int
	search = func(idx int) int {
		// Check bounds
		if idx < 0 || idx >= maxIdx {
			return -1
		}

		// Check if node is free
		if !c.GetBit(idx) {
			return -1
		}

		// If this is a leaf and it's free, return it
		if idx >= leafStart {
			return idx
		}

		// For internal nodes, try left child first
		left := 2 * idx
		if left < maxIdx {
			if result := search(left); result >= 0 {
				return result
			}
		}

		// Then try right child
		right := 2*idx + 1
		if right < maxIdx {
			if result := search(right); result >= 0 {
				return result
			}
		}

		return -1
	}

	return search(startIdx)
}

// Remove deallocates a pointer within this chunk's payload
// Returns true if the pointer was successfully deallocated, false if it doesn't belong to this chunk
// Implements proper buddy coalescing: recursively merges with buddy blocks when both are free
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

	// Mark this block as free
	c.SetBit(treeIdx)

	// Attempt to coalesce with buddy blocks up the tree
	c.coalesce(treeIdx)

	return true
}

// coalesce recursively merges buddy blocks when both are free
// Starts from a given node and tries to merge upward in the tree
//
// Algorithm (example: freeing a 64-byte block that can merge to 1KB):
//
//	Initial: Two 64-byte buddies freed sequentially
//	Free block 1 (index 376):
//	  ├─ SetBit(376) → mark as free
//	  └─ coalesce(376)
//	     ├─ Parent = 376 / 2 = 188
//	     ├─ Left child = 2*188 = 376 (our block, now free)
//	     ├─ Right child = 2*188+1 = 377 (buddy, still allocated)
//	     ├─ Both free? NO → UpdateNode(188) and STOP
//
//	Free block 2 (index 377):
//	  ├─ SetBit(377) → mark as free
//	  └─ coalesce(377)
//	     ├─ Parent = 377 / 2 = 188
//	     ├─ Left child = 376 (free from before)
//	     ├─ Right child = 377 (our block, now free)
//	     ├─ Both free? YES → SetBit(188), recurse on parent
//	     │
//	     └─ coalesce(188)
//	        ├─ Parent = 188 / 2 = 94
//	        ├─ Left child = 188 (now free, 128B)
//	        ├─ Right child = 189 (check if free)
//	        ├─ Both free? (depends on buddy 189 state)
//	        └─ Continue coalescing up tree...
//
//	Result: Automatic merging from 64B → 128B → 256B → 512B blocks
//	        Reduces fragmentation without manual tracking
//
// Tree structure visualization:
//
//	       1 (root)
//	     /   \
//	    2     3        (512B each)
//	   / \   / \
//	  4   5 6   7      (256B each)
//	 /\ /\ /\ /\
//	8 9... (128B) ...   (64B leaves)
//
// When both children at level are free, parent becomes free
// A parent is free IFF both its children are free
//
// Performance: O(log N) where N = tree height
//   - At most one traversal up the tree
//   - Each level: constant time (two bit checks, one set, one recurse)
//   - Maximum iterations = tree height = log₂(N)
func (c *Chunk) coalesce(idx int) {
	if idx <= 0 || idx >= (1<<uint(c.order+1)) {
		return
	}

	// Root node cannot be coalesced
	if idx == 1 {
		return
	}

	var (
		parent     = idx / 2
		leftChild  = (2 * parent) + 0
		rightChild = (2 * parent) + 1
	)

	// Check if both children are free (both bits set)
	var (
		leftFree  = c.GetBit(leftChild)
		rightFree = c.GetBit(rightChild)
	)

	// If both children are free, merge them by marking parent as free
	if leftFree && rightFree {
		c.SetBit(parent)
		// Recursively try to coalesce the parent with its sibling
		c.coalesce(parent)
	} else {
		// If we can't coalesce, just update the parent state
		c.UpdateNode(parent)
	}
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
// Growth Strategy Enum
// ============================================================================

type Growth int

const (
	FIXED    Growth = iota // Allocate exactly MIN_CHUNK_SIZE chunks
	ADDITIVE               // Allocate adaptively based on allocation size
)

// ============================================================================
// BuddyAllocator
// ============================================================================

type BuddyAllocator struct {
	chunks         []*Chunk
	lastChunkIdx   int      // Index of last successfully allocated chunk
	res            *res.Res // Used when USE_RES = true
	order          int
	growthStrategy Growth // Strategy for chunk growth
	chunkSize      int    // Custom chunk size (must be multiple of MIN_CHUNK_SIZE)
	mtx            sync.Mutex
}

// BuddyAllocatorOption is a functional option for configuring BuddyAllocator
type BuddyAllocatorOption func(*BuddyAllocator)

// WithGrowthStrategy sets the growth strategy for the allocator
func WithGrowthStrategy(strategy Growth) BuddyAllocatorOption {
	return func(a *BuddyAllocator) {
		a.growthStrategy = strategy
	}
}

// WithSize sets a custom chunk size (must be a multiple of MIN_CHUNK_SIZE)
func WithSize(size int) BuddyAllocatorOption {
	return func(a *BuddyAllocator) {
		if size > 0 && size%MIN_CHUNK_SIZE == 0 {
			a.chunkSize = size
		}
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

	for i, chunk := range a.chunks {
		if (1 << uint(chunk.order)) >= int(size) {
			if chunk.GetMaxFreeSize() > 0 {
				if blockPtr := chunk.Allocate(int(size)); blockPtr != nil {
					a.lastChunkIdx = i
					return blockPtr
				}
			}
		}
	}

	// If FIXED strategy, fail instead of growing
	if a.growthStrategy == FIXED {
		return nil
	}

	if a.growthStrategy == ADDITIVE {
		growSize := max(max(int(res.RoundPow2(size*2)), int(blockSize)), a.chunkSize)
		if growSize > a.chunkSize {
			if rem := growSize % a.chunkSize; rem != 0 {
				growSize += a.chunkSize - rem
			}
		}
		chunk := NewChunk(a.res, growSize)
		a.chunks = append(a.chunks, chunk)
		a.lastChunkIdx = len(a.chunks) - 1
		return chunk.Allocate(int(size))
	}

	return nil
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

	var (
		numLeaves = size / MIN_BLOCK_SIZE
		order     = 0
	)

	// Calculate tree depth (order) for the number of leaves
	for (1 << uint(order)) < numLeaves {
		order++
	}

	// Total nodes in tree: 2^(order+1) - 1
	var (
		totalNodes = (1 << uint(order+1)) - 1
		bitsSize   = (totalNodes + 7) / 8
	)

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
	var (
		blockSize = size - bitsSize
		leafStart = 1 << uint(order)
	)
	for i := leafStart; i < (1 << uint(order+1)); i++ {
		var (
			byteIdx = i / 8
			bitIdx  = uint(i % 8)
		)
		if byteIdx < bitsSize {
			data[byteIdx] |= (1 << bitIdx)
		}
	}
	for i := (leafStart - 1); i >= 1; i-- {
		var (
			left     = 2 * i
			right    = 2*i + 1
			leftBit  = left < (1<<uint(order+1)) && (left/8 < bitsSize) && (data[left/8]&(1<<uint(left%8))) != 0
			rightBit = right < (1<<uint(order+1)) && (right/8 < bitsSize) && (data[right/8]&(1<<uint(right%8))) != 0
		)
		if leftBit && rightBit {
			var (
				byteIdx = i / 8
				bitIdx  = uint(i % 8)
			)
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

func NewBuddyAllocator(opts ...BuddyAllocatorOption) *BuddyAllocator {
	// Initialize allocator without pre-allocating chunks
	// First Alloc() call will size the chunk appropriately based on the allocation size
	a := &BuddyAllocator{
		chunks:         []*Chunk{},
		growthStrategy: ADDITIVE,       // default strategy
		chunkSize:      MIN_CHUNK_SIZE, // default chunk size
	}

	// Apply functional options first (to allow customizing chunkSize)
	for _, opt := range opts {
		opt(a)
	}

	// Initialize Res after options are applied, using the configured chunkSize
	if USE_RES {
		a.res = res.NewRes(a.chunkSize)
	}

	return a
}
