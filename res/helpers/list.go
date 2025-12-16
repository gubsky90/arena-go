package helpers

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
