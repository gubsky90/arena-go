package res

import (
	"sync"
	"unsafe"
)

type Page struct {
	base []byte
	size int
}

func (p *Page) Base() []byte {
	return p.base
}

func NewPage(size int) *Page {
	base := MakePages(size)
	// fmt.Printf("[RES] NewPage: allocated %d bytes at %p\n", size, unsafe.SliceData(base))
	return &Page{base: base, size: len(base)}
}

type Res struct {
	chunks  []*Page
	current int
	offset  int
	mtx     sync.Mutex
}

func NewRes(size int) *Res {
	initialSize := max(size, PAGE_SIZE)
	return &Res{
		chunks:  []*Page{NewPage(initialSize)},
		current: 0,
		offset:  0,
	}
}

func (r *Res) Reset() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	r.offset, r.current = 0, 0
}

func (r *Res) Alloc(size, align uint64) unsafe.Pointer {
	// fmt.Println("Size: ", size, "Align", align)
	if align < 1 {
		align = 1
	}
	r.mtx.Lock()
	defer r.mtx.Unlock()
	for {
		var (
			chunk   = r.chunks[r.current]
			base    = unsafe.Pointer(unsafe.SliceData(chunk.base))
			current = uintptr(base) + uintptr(r.offset)
			aligned = (current + uintptr(align-1)) &^ uintptr(align-1)
			end     = aligned + uintptr(size)
		)
		if int(end-uintptr(base)) <= chunk.size {
			r.offset = int(end - uintptr(base))
			ptr := unsafe.Add(base, int(aligned-uintptr(base)))
			return ptr
		}
		if r.current+1 >= len(r.chunks) {
			r.chunks = append(r.chunks, NewPage(max(int(size+align-1), r.chunks[0].size)))
		}
		r.current++
		r.offset = 0
	}
}

func (r *Res) Delete() {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	for _, p := range r.chunks {
		ReleasePages(p.base)
	}
	r.chunks = nil
	r.current, r.offset = 0, 0
}

func (r *Res) Owns(ptr unsafe.Pointer) bool {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	addr := uintptr(ptr)
	for _, p := range r.chunks {
		var (
			start = uintptr(unsafe.Pointer(unsafe.SliceData(p.base)))
			end   = start + uintptr(len(p.base))
		)
		if addr >= start && addr < end {
			return true
		}
	}
	return false
}

func (r *Res) Chunks() []*Page {
	return r.chunks
}

func (r *Res) Current() int {
	return r.current
}

func (r *Res) Size() int {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	size := 0
	for _, c := range r.chunks {
		size = size + len(c.base)
	}
	return size
}

// New creates a new chunk of the given size and adds it to the resource manager
func (r *Res) New(size int) *Page {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	page := NewPage(size)
	r.chunks = append(r.chunks, page)
	return page
}

// FindPage returns the Page that contains the given pointer and true,
// or an empty Page and false if the pointer is not owned by any page.
func (r *Res) FindPage(ptr unsafe.Pointer) (*Page, bool) {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	addr := uintptr(ptr)
	for _, p := range r.chunks {
		start := uintptr(unsafe.Pointer(unsafe.SliceData(p.base)))
		end := start + uintptr(len(p.base))
		if addr >= start && addr < end {
			return p, true
		}
	}
	return nil, false
}
