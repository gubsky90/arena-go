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
│ Buddy Bitmap (Complete Binary Tree)  │  ← bitset metadata
│ e.g., 64KB chunk → ~1KB bitmap       │
├──────────────────────────────────────┤  ← payloadStart = base + bitsSize
│ Buddy blocks (16B, 32B, 64B, ...)    │
│ - Power-of-2 sized and aligned       │
└──────────────────────────────────────┘

• Power-of-2 sized blocks (16B minimum)
• Bitmap: bit = 0 → subtree fully free, bit = 1 → allocated or partially used
• Proper splitting and coalescing
• Size-aware search for large allocations
• Scalable chunk lookup via binary search
• Low metadata overhead, excellent cache locality
================================================================================
*/

const (
	MIN_ORDER        = 4 // 16-byte minimum block
	MIN_BLOCK_SIZE   = 1 << MIN_ORDER
	MAX_LEVEL        = 32 // Max power-of-2 levels for free lists (2^4 to 2^35)
	QWORD_SIZE_BYTES = 8  // Size of uint64 in bytes
	QWORD_SIZE_BITS  = 64 // Number of bits in uint64
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
	maxFreeSize  int     // cached maximum free block size (updated on Allocate/Remove)
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
		qwordIdx    = idx / QWORD_SIZE_BITS
		bitIdx      = uint(idx % QWORD_SIZE_BITS)
		qwordOffset = qwordIdx * QWORD_SIZE_BYTES
	)
	if qwordOffset+QWORD_SIZE_BYTES <= c.bitsSize {
		// Read uint64, set bit, write back
		ptr := unsafe.Pointer(unsafe.SliceData(c.data[qwordOffset:]))
		val := (*uint64)(ptr)
		*val |= (1 << bitIdx)
	}
}

// ClearBit clears a bit in the bitset at given position using uint64 operations
func (c *Chunk) ClearBit(idx int) {
	var (
		qwordIdx    = idx / QWORD_SIZE_BITS
		bitIdx      = uint(idx % QWORD_SIZE_BITS)
		qwordOffset = qwordIdx * QWORD_SIZE_BYTES
	)
	if qwordOffset+QWORD_SIZE_BYTES <= c.bitsSize {
		// Read uint64, clear bit, write back
		ptr := unsafe.Pointer(unsafe.SliceData(c.data[qwordOffset:]))
		val := (*uint64)(ptr)
		*val &= ^(1 << bitIdx)
	}
}

// GetBit gets a bit from the bitset at given position using uint64 operations
func (c *Chunk) GetBit(idx int) bool {
	var (
		qwordIdx    = idx / QWORD_SIZE_BITS
		bitIdx      = uint(idx % QWORD_SIZE_BITS)
		qwordOffset = qwordIdx * QWORD_SIZE_BYTES
	)
	if qwordOffset+QWORD_SIZE_BYTES > c.bitsSize {
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

// OffsetForIndex calculates the payload offset for a given tree node index
func (c *Chunk) OffsetForIndex(idx int) int {
	if idx <= 0 {
		return 0
	}
	// Find the leftmost leaf under this node to determine its offset
	for idx < (1 << uint(c.order)) {
		idx = idx << 1
	}
	return (idx - (1 << uint(c.order))) * MIN_BLOCK_SIZE
}

// UpdateNode updates a node and propagates changes up to root iteratively
func (c *Chunk) UpdateNode(idx int) {
	for idx > 0 {
		if idx >= (1 << uint(c.order+1)) {
			return
		}

		// For leaf nodes, don't update, just move to parent
		if idx >= (1 << uint(c.order)) {
			idx = idx / 2
			continue
		}

		var (
			leftFree  = !c.GetBit((2 * idx) + 0)
			rightFree = !c.GetBit((2 * idx) + 1)
		)

		// Parent is free only if both children are free
		if leftFree && rightFree {
			c.ClearBit(idx)
		} else {
			c.SetBit(idx)
		}

		// Move to parent
		idx = idx / 2
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
	if c.maxFreeSize < size {
		return nil
	}

	var (
		requiredSize = max(size, MIN_BLOCK_SIZE)
	)

	// Find a free node at the appropriate size or larger
	idx := c.FindFreeNode(1, requiredSize)
	if idx < 0 {
		return nil
	}

	// Split the block down to the requested size
	// Start from found index and split down until we reach the right size
	idx, maxFreeBuddySize := c.split(idx, requiredSize)

	// Mark the final allocated block as used
	c.SetBit(idx)

	// Update parent chain
	c.UpdateNode(idx / 2)

	// Update cached max free size: track largest buddy from split or existing max
	if maxFreeBuddySize > 0 {
		if maxFreeBuddySize > c.maxFreeSize {
			c.maxFreeSize = maxFreeBuddySize
		}
	} else {
		// No split occurred, meaning we allocated a block of size requiredSize.
		// If this was the only block of size c.maxFreeSize, we should decrease it.
		if requiredSize == c.maxFreeSize {
			// Conservative estimate: next largest could be requiredSize / 2
			c.maxFreeSize = requiredSize / 2
		}
	}

	return unsafe.Add(unsafe.Pointer(unsafe.SliceData(c.data)), c.PayloadOffset()+c.OffsetForIndex(idx))
}

// getSizeForIndex returns the block size (in bytes) for a given tree node index
func (c *Chunk) getSizeForIndex(idx int) int {
	if idx <= 0 || idx >= (1<<uint(c.order+1)) {
		return 0
	}

	// Calculate block size based on depth from root
	// Count how many times to divide idx by 2 until reaching root (idx=1)
	// This gives depth from root: at each level down, size halves
	var (
		depthFromRoot = 0
		temp          = idx
	)
	for temp > 1 {
		temp = temp / 2
		depthFromRoot = depthFromRoot + 1
	}

	// Size = blockSize >> depthFromRoot
	// Root (depth=0): blockSize >> 0 = blockSize
	// Children of root (depth=1): blockSize >> 1 = blockSize/2
	// Leaves (depth=order): blockSize >> order = MIN_BLOCK_SIZE
	return c.blockSize >> uint(depthFromRoot)
}

// split recursively divides a block until it matches the required size
// Returns the final node index and the size of the largest free buddy created
func (c *Chunk) split(idx int, requiredSize int) (int, int) {
	var (
		maxIdx           = 1 << uint(c.order+1)
		currentNodeSize  = c.getSizeForIndex(idx)
		maxFreeBuddySize = 0
	)

	// If we are allocating the entire block without splitting,
	// the largest free buddy created is 0.
	// However, we could track currentNodeSize/2 as a conservative estimate
	// if we wanted to ensure maxFreeSize is updated even when no split occurs.
	// For now, we track actual free buddies created.

	for currentNodeSize > requiredSize && idx < maxIdx {
		var (
			leftChild  = (2 * idx) + 0
			rightChild = (2 * idx) + 1
		)

		if rightChild >= maxIdx {
			break
		}

		// Right child becomes a free buddy block
		rightBuddySize := currentNodeSize / 2
		c.ClearBit(rightChild)

		// Update parent of split node
		c.UpdateNode(idx / 2)

		// Track largest buddy created
		if rightBuddySize > maxFreeBuddySize {
			maxFreeBuddySize = rightBuddySize
		}

		// Continue with left child
		idx = leftChild
		currentNodeSize = currentNodeSize / 2
	}

	return idx, maxFreeBuddySize
}

// FindFreeNode iteratively searches the tree for a free node that is at least the requested size
// Uses depth-first traversal with preference for left children via stack
// Returns the index of a suitable free node, or -1 if none found
func (c *Chunk) FindFreeNode(startIdx int, requiredSize int) int {
	var (
		maxIdx = 1 << uint(c.order+1)
	)

	// Use a stack for iterative DFS: process left children before right
	stack := []int{startIdx}

	for len(stack) > 0 {
		// Pop from stack
		idx := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Check bounds
		if idx <= 0 || idx >= maxIdx {
			continue
		}

		// Check size of this node
		nodeSize := c.getSizeForIndex(idx)
		if nodeSize < requiredSize {
			continue
		}

		// If this node is completely free (bit is 0), return it.
		// split() will handle splitting it down to requiredSize if needed.
		if !c.GetBit(idx) {
			return idx
		}

		// If we are here, GetBit(idx) is true, meaning the node is partially or fully allocated.
		// If it's exactly the size we need, we can't use it.
		if nodeSize == requiredSize {
			continue
		}

		// If it's larger than requiredSize, descend to find a free sub-block.
		// BUT ONLY if it's not fully allocated as a single block.
		// A node is fully allocated if its bit is 1 but its children's bits are 0.
		var (
			left  = (2 * idx) + 0
			right = (2 * idx) + 1
		)
		if right < maxIdx {
			if c.GetBit(left) || c.GetBit(right) {
				stack = append(stack, right)
				stack = append(stack, left)
			}
		}
	}

	return -1
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

	// Find the actual allocated block.
	// It's either this leaf or one of its ancestors.
	// The allocated block is the first node we find going up that has its bit set
	// AND (is a leaf OR has both children bits clear).
	curr := treeIdx
	for curr > 0 {
		if c.GetBit(curr) {
			// Check if this is the allocated block.
			// If it's a leaf, it's the block.
			if curr >= (1 << uint(c.order)) {
				break
			}
			// If it's an internal node, it's the block if its children bits are 0.
			var (
				left  = (2 * curr) + 0
				right = (2 * curr) + 1
			)
			if !c.GetBit(left) && !c.GetBit(right) {
				break
			}
		}
		curr = curr / 2
	}

	if curr <= 0 || !c.GetBit(curr) {
		return false
	}

	// Mark this block as free
	c.ClearBit(curr)

	// Attempt to coalesce with buddy blocks up the tree
	// Returns the size of the largest free block created
	coalescedSize := c.coalesce(curr)

	// Update maxFreeSize if coalesced block is larger
	if coalescedSize > c.maxFreeSize {
		c.maxFreeSize = coalescedSize
	}

	return true
}

// coalesce recursively merges buddy blocks when both are free iteratively
// Starts from a given node and tries to merge upward in the tree
// Returns the size of the largest free block created by coalescing
//
// Algorithm (example: freeing a 64-byte block that can merge to 1KB):
//
//	Initial: Two 64-byte buddies freed sequentially
//	Free block 1 (index 376):
//	  ├─ ClearBit(376) → mark as free
//	  └─ coalesce(376)
//	     ├─ Parent = 376 / 2 = 188
//	     ├─ Left child = 2*188 = 376 (our block, now free)
//	     ├─ Right child = 2*188+1 = 377 (buddy, still allocated)
//	     ├─ Both free? NO → UpdateNode(188) and return size(376)
//
//	Free block 2 (index 377):
//	  ├─ ClearBit(377) → mark as free
//	  └─ coalesce(377)
//	     ├─ Parent = 377 / 2 = 188
//	     ├─ Left child = 376 (free from before)
//	     ├─ Right child = 377 (our block, now free)
//	     ├─ Both free? YES → ClearBit(188), repeat for parent
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
func (c *Chunk) coalesce(idx int) int {
	for idx > 1 {
		var (
			parent     = idx / 2
			leftChild  = (2 * parent) + 0
			rightChild = (2 * parent) + 1
		)

		// Check if both children are free (both bits clear)
		if !c.GetBit(leftChild) && !c.GetBit(rightChild) {
			c.ClearBit(parent)
			idx = parent
		} else {
			// If we can't coalesce, just update the parent state and stop
			c.UpdateNode(parent)
			return c.getSizeForIndex(idx)
		}
	}
	return c.getSizeForIndex(1)
}

// Reset resets all allocations in the chunk
func (c *Chunk) Reset() {
	// Set all bits to 0 (free state)
	for i := 0; i < c.bitsSize; i++ {
		c.data[i] = 0x00
	}
	// Entire chunk is free after reset
	c.maxFreeSize = c.blockSize
}

// GetMaxFreeSize returns the cached maximum allocation size available in this chunk.
// The value is updated during Allocate() and Remove() operations for O(1) lookup.
func (c *Chunk) GetMaxFreeSize() int {
	return c.maxFreeSize
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
	growthStrategy Growth   // Strategy for chunk growth
	chunkSize      int      // Custom chunk size (must be multiple of MIN_CHUNK_SIZE)
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

// WithSize sets a custom data size (rounded to nearest multiple of MIN_DATA_SIZE)
func WithSize(size int) BuddyAllocatorOption {
	return func(a *BuddyAllocator) {
		var (
			bitmapSize = MIN_BITMAP_SIZE
			dataSize   = MIN_DATA_SIZE
		)
		if size > 0 {
			// Round up to nearest multiple of MIN_DATA_SIZE
			if rem := size % MIN_DATA_SIZE; rem != 0 {
				size = size + (MIN_DATA_SIZE - rem)
			}
			dataSize = size
			bitmapSize = (dataSize / MIN_DATA_SIZE) * MIN_BITMAP_SIZE
		}
		a.chunkSize = dataSize + bitmapSize
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
	blockSize := max(max(res.RoundPow2(size), align), MIN_BLOCK_SIZE)

	a.mtx.Lock()
	defer a.mtx.Unlock()

	// Try last chunk first (cache locality)
	if a.lastChunkIdx >= 0 && a.lastChunkIdx < len(a.chunks) {
		chunk := a.chunks[a.lastChunkIdx]
		if chunk.GetMaxFreeSize() >= int(size) {
			if blockPtr := chunk.Allocate(int(size)); blockPtr != nil {
				return blockPtr
			}
		}
	}

	// Find best-fit chunk (smallest suitable chunk with available space)
	// Use single-pass linear search - O(n) is better than sorting O(n log n) for typical cases
	var (
		bestIdx  int = -1
		bestSize int = int(^uint(0) >> 1) // max int
	)

	for i, chunk := range a.chunks {
		chunkCapacity := 1 << uint(chunk.order)
		if chunk.GetMaxFreeSize() >= int(size) {
			// Best-fit: prefer smallest suitable chunk
			if chunkCapacity < bestSize {
				bestIdx = i
				bestSize = chunkCapacity
			}
		}
	}

	if bestIdx >= 0 {
		if blockPtr := a.chunks[bestIdx].Allocate(int(size)); blockPtr != nil {
			a.lastChunkIdx = bestIdx
			return blockPtr
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
				growSize = growSize + a.chunkSize - rem
			}
		}
		chunk := NewChunk(a.res, growSize)

		// Insert chunk in sorted order by startAddress for binary search efficiency
		insertIdx := 0
		for insertIdx < len(a.chunks) && a.chunks[insertIdx].startAddress < chunk.startAddress {
			insertIdx = insertIdx + 1
		}

		// Insert at position insertIdx
		a.chunks = append(a.chunks[:insertIdx], append([]*Chunk{chunk}, a.chunks[insertIdx:]...)...)
		a.lastChunkIdx = insertIdx

		return chunk.Allocate(int(size))
	}

	return nil
}

// findChunkByAddress uses binary search to find chunk owning a pointer
// Chunks are sorted by startAddress
func (a *BuddyAllocator) findChunkByAddress(ptr uintptr) *Chunk {
	// Binary search for the chunk
	var (
		left  = 0
		right = len(a.chunks) - 1
	)

	for left <= right {
		var (
			mid   = (left + right) / 2
			chunk = a.chunks[mid]
		)

		if ptr < chunk.startAddress {
			right = mid - 1
		} else if ptr >= chunk.endAddress {
			left = mid + 1
		} else {
			// Found: ptr is in range [startAddress, endAddress)
			return chunk
		}
	}

	return nil
}

func (a *BuddyAllocator) Remove(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}

	a.mtx.Lock()
	defer a.mtx.Unlock()

	// Binary search to find the chunk that owns this pointer
	chunk := a.findChunkByAddress(uintptr(ptr))
	if chunk != nil {
		chunk.Remove(ptr)
	}
}
func (a *BuddyAllocator) Reset() {
	a.mtx.Lock()
	defer a.mtx.Unlock()
	for _, chunk := range a.chunks {
		chunk.Reset()
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
}

func (a *BuddyAllocator) Owns(ptr unsafe.Pointer) bool {
	if ptr == nil {
		return false
	}

	a.mtx.Lock()
	defer a.mtx.Unlock()

	// Binary search to find the chunk that owns this pointer
	chunk := a.findChunkByAddress(uintptr(ptr))
	return chunk != nil
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

	// Find the largest power-of-2 block size that fits in the given size
	// along with its required metadata bits.
	// Each 16B block needs 1 bit at the leaf level, plus internal node bits.
	// Total bits = 2 * numLeaves - 1
	order := 0
	for {
		var (
			numLeaves    = 1 << uint(order)
			totalNodes   = (1 << uint(order+1)) - 1
			testBitsSize = (totalNodes + 7) / QWORD_SIZE_BITS * QWORD_SIZE_BYTES
		)
		if (numLeaves*MIN_BLOCK_SIZE)+testBitsSize > size {
			order--
			break
		}
		order++
	}

	var (
		numLeaves  = 1 << uint(order)
		totalNodes = (1 << uint(order+1)) - 1
		bitsSize   = (totalNodes + 7) / QWORD_SIZE_BITS * QWORD_SIZE_BYTES
		blockSize  = numLeaves * MIN_BLOCK_SIZE
	)

	// Initialize bitset in data buffer (all 0 = all free)
	for i := 0; i < bitsSize && i < len(data); i++ {
		data[i] = 0
	}

	startAddress := uintptr(unsafe.Pointer(unsafe.SliceData(data)))
	chunk := &Chunk{
		data:         data,
		startAddress: startAddress,
		endAddress:   startAddress + uintptr(len(data)),
		blockSize:    blockSize,
		order:        order,
		bitsSize:     bitsSize,
		maxFreeSize:  blockSize, // entire chunk is initially free
	}

	// Reset to ensure all bits are properly initialized to free state
	chunk.Reset()
	return chunk
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
