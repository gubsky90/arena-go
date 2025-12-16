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
	r.mtx.Lock()
	defer r.mtx.Unlock()
	aligned := (r.offset + int(align-1)) &^ int(align-1)
	if aligned+int(size) > r.chunks[r.current].size {
		if r.current+1 >= len(r.chunks) {
			sz := max(int(size), r.chunks[0].size)
			r.chunks = append(r.chunks, NewPage(sz))
		}
		r.current = r.current + 1
		r.offset, aligned = 0, 0
	}
	ptr := unsafe.Pointer(unsafe.SliceData(r.chunks[r.current].base[aligned:][:int(size)]))
	r.offset = aligned + int(size)
	return ptr
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
