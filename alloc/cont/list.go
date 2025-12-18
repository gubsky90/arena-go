package cont

// Node represents a node in the doubly-linked list.
type Node[T any] struct {
	Value T
	Prev  *Node[T]
	Next  *Node[T]
}

// List is a generic doubly-linked list.
type List[T any] struct {
	Head *Node[T]
	Tail *Node[T]
	Len  int
}

// NewList creates a new empty list.
func NewList[T any]() *List[T] {
	return &List[T]{}
}

// PushBack adds an item to the end of the list (tail).
func (l *List[T]) PushBack(value T) {
	temp := &Node[T]{Value: value}

	if l.Tail == nil {
		l.Head = temp
		l.Tail = temp
	} else {
		temp.Prev = l.Tail
		l.Tail.Next = temp
		l.Tail = temp
	}
	l.Len++
}

// PushFront adds an item to the beginning of the list (head).
func (l *List[T]) PushFront(value T) {
	newNode := &Node[T]{Value: value}

	if l.Head == nil {
		l.Head = newNode
		l.Tail = newNode
	} else {
		newNode.Next = l.Head
		l.Head.Prev = newNode
		l.Head = newNode
	}
	l.Len++
}

// PopBack removes and returns the item from the end (tail).
// Returns zero value and false if empty.
func (l *List[T]) PopBack() (T, bool) {
	if l.Tail == nil {
		var zero T
		return zero, false
	}

	node := l.Tail
	value := node.Value

	if node.Prev == nil {
		// Only one node
		l.Head = nil
		l.Tail = nil
	} else {
		l.Tail = node.Prev
		l.Tail.Next = nil
		node.Prev = nil
	}
	l.Len--
	return value, true
}

// PopFront removes and returns the item from the beginning (head).
// Returns zero value and false if empty.
func (l *List[T]) PopFront() (T, bool) {
	if l.Head == nil {
		var zero T
		return zero, false
	}

	node := l.Head
	value := node.Value

	if node.Next == nil {
		// Only one node
		l.Head = nil
		l.Tail = nil
	} else {
		l.Head = node.Next
		l.Head.Prev = nil
		node.Next = nil
	}
	l.Len--
	return value, true
}

// Remove removes a specific node from the list.
// The node must be part of this list.
func (l *List[T]) Remove(node *Node[T]) {
	if node == nil {
		return
	}

	if node.Prev == nil {
		l.Head = node.Next
	} else {
		node.Prev.Next = node.Next
	}

	if node.Next == nil {
		l.Tail = node.Prev
	} else {
		node.Next.Prev = node.Prev
	}

	node.Prev = nil
	node.Next = nil
	l.Len--
}

// Front returns the first node (or nil if empty).
func (l *List[T]) Front() *Node[T] {
	return l.Head
}

// Back returns the last node (or nil if empty).
func (l *List[T]) Back() *Node[T] {
	return l.Tail
}

// Length returns the number of elements.
func (l *List[T]) Length() int {
	return l.Len
}

// IsEmpty reports whether the list is empty.
func (l *List[T]) IsEmpty() bool {
	return l.Len == 0
}

// Reset clears the list without allocating.
func (l *List[T]) Reset() {
	l.Head = nil
	l.Tail = nil
	l.Len = 0
}

// LIterator allows iteration over the list.
type LIterator[T any] struct {
	list    *List[T]
	current *Node[T]
}

// Iter returns an iterator starting from the head of the list.
func (l *List[T]) Iter() *LIterator[T] {
	return &LIterator[T]{
		list:    l,
		current: l.Head,
	}
}

// Next advances the iterator and returns the next value.
// Returns zero value and false when iteration is done.
func (it *LIterator[T]) Next() (T, bool) {
	if it.current == nil {
		var zero T
		return zero, false
	}

	value := it.current.Value
	it.current = it.current.Next
	return value, true
}

// InsertBefore inserts a new element with value v immediately before mark
// and returns the new element. If mark is not in the list, the list is not modified.
func (l *List[T]) InsertBefore(v T, mark *Node[T]) *Node[T] {
	if mark == nil {
		return nil
	}

	temp := &Node[T]{Value: v}

	temp.Prev = mark.Prev
	temp.Next = mark
	mark.Prev = temp

	if temp.Prev == nil {
		l.Head = temp
	} else {
		temp.Prev.Next = temp
	}

	l.Len++
	return temp
}

// InsertAfter inserts a new element with value v immediately after mark
// and returns the new element. If mark is not in the list, the list is not modified.
func (l *List[T]) InsertAfter(v T, mark *Node[T]) *Node[T] {
	if mark == nil {
		return nil
	}

	temp := &Node[T]{Value: v}

	temp.Next = mark.Next
	temp.Prev = mark
	mark.Next = temp

	if temp.Next == nil {
		l.Tail = temp
	} else {
		temp.Next.Prev = temp
	}

	l.Len++
	return temp
}

// MoveToFront moves element e to the front of the list.
// If e is not an element of the list, the list is not modified.
func (l *List[T]) MoveToFront(e *Node[T]) {
	if e == nil || e == l.Head {
		return
	}

	// Remove e from its current position
	if e.Prev != nil {
		e.Prev.Next = e.Next
	}
	if e.Next != nil {
		e.Next.Prev = e.Prev
	}
	if e == l.Tail {
		l.Tail = e.Prev
	}

	// Insert at front
	e.Prev = nil
	e.Next = l.Head
	if l.Head != nil {
		l.Head.Prev = e
	}
	l.Head = e
	if l.Tail == nil {
		l.Tail = e
	}
}

// MoveToBack moves element e to the back of the list.
// If e is not an element of the list, the list is not modified.
func (l *List[T]) MoveToBack(e *Node[T]) {
	if e == nil || e == l.Tail {
		return
	}

	// Remove e from its current position
	if e.Prev != nil {
		e.Prev.Next = e.Next
	}
	if e.Next != nil {
		e.Next.Prev = e.Prev
	}
	if e == l.Head {
		l.Head = e.Next
	}

	// Insert at back
	e.Next = nil
	e.Prev = l.Tail
	if l.Tail != nil {
		l.Tail.Next = e
	}
	l.Tail = e
	if l.Head == nil {
		l.Head = e
	}
}

// MoveBefore moves element e to its new position before mark.
// If e or mark is not an element of the list, or e == mark, the list is not modified.
func (l *List[T]) MoveBefore(e, mark *Node[T]) {
	if e == nil || mark == nil || e == mark {
		return
	}

	// Remove e from its current position
	if e.Prev != nil {
		e.Prev.Next = e.Next
	} else {
		l.Head = e.Next
	}
	if e.Next != nil {
		e.Next.Prev = e.Prev
	} else {
		l.Tail = e.Prev
	}

	// Insert before mark
	e.Prev = mark.Prev
	e.Next = mark
	mark.Prev = e

	if e.Prev == nil {
		l.Head = e
	} else {
		e.Prev.Next = e
	}
}

// MoveAfter moves element e to its new position after mark.
// If e or mark is not an element of the list, or e == mark, the list is not modified.
func (l *List[T]) MoveAfter(e, mark *Node[T]) {
	if e == nil || mark == nil || e == mark {
		return
	}

	// Remove e from its current position
	if e.Prev != nil {
		e.Prev.Next = e.Next
	} else {
		l.Head = e.Next
	}
	if e.Next != nil {
		e.Next.Prev = e.Prev
	} else {
		l.Tail = e.Prev
	}

	// Insert after mark
	e.Next = mark.Next
	e.Prev = mark
	mark.Next = e

	if e.Next == nil {
		l.Tail = e
	} else {
		e.Next.Prev = e
	}
}

// PushBackList inserts a copy of another list at the back of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[T]) PushBackList(other *List[T]) {
	if other == nil || other.Head == nil {
		return
	}

	for node := other.Head; node != nil; node = node.Next {
		l.PushBack(node.Value)
	}
}

// PushFrontList inserts a copy of another list at the front of list l.
// The lists l and other may be the same. They must not be nil.
func (l *List[T]) PushFrontList(other *List[T]) {
	if other == nil || other.Head == nil {
		return
	}

	// Traverse in reverse order to maintain relative order at front
	var nodes []*Node[T]
	for node := other.Head; node != nil; node = node.Next {
		nodes = append(nodes, node)
	}

	for i := len(nodes) - 1; i >= 0; i-- {
		l.PushFront(nodes[i].Value)
	}
}
