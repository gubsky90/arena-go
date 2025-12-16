package cont

// Stack is a generic LIFO stack.
type Stack[T any] struct {
	items []T
}

// NewStack creates a new empty stack with default capacity 8.
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		items: make([]T, 0, 8),
	}
}

// Push adds an item to the top of the stack.
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

// Pop removes and returns the top item. Returns zero value and false if empty.
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, true
}

// Peek returns the top item without removing it. Returns zero value and false if empty.
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// Len returns the number of items in the stack.
func (s *Stack[T]) Len() int {
	return len(s.items)
}

// IsEmpty reports whether the stack is empty.
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Reset clears the stack (reuses backing array).
func (s *Stack[T]) Reset() {
	s.items = s.items[:0]
}
