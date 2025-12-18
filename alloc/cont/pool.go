package cont

import "sync"

// Pool manages a freelist of items
type Pool[T any] struct {
	freelist *List[T]
	mtx      sync.Mutex
}

// NewPool creates a new pool
func NewPool[T any]() *Pool[T] {
	return &Pool[T]{
		freelist: NewList[T](),
	}
}

// Put adds an item to the freelist
func (p *Pool[T]) Put(item T) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if p.freelist == nil {
		p.freelist = NewList[T]()
	}
	p.freelist.PushBack(item)
}

// Get retrieves an item from the freelist, or returns zero value if empty
func (p *Pool[T]) Get() T {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if p.freelist == nil || p.freelist.Length() == 0 {
		var zero T
		return zero
	}
	item, _ := p.freelist.PopFront()
	return item
}

// Length returns the number of items in the freelist
func (p *Pool[T]) Length() int {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if p.freelist == nil {
		return 0
	}
	return p.freelist.Length()
}

// Reset clears the freelist
func (p *Pool[T]) Reset() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.freelist = NewList[T]()
}
