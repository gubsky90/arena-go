package container

import (
	"github.com/thebagchi/arena-go"
)

// Stack is a generic LIFO (Last In First Out) data structure.
type Stack[T any] struct {
	data *Vec[T]
}

// NewStack creates a new empty stack backed by the provided arena.
func NewStack[T any](a *arena.Arena) *Stack[T] {
	return &Stack[T]{
		data: NewVec[T](a),
	}
}

// Push adds an element to the top of the stack.
func (s *Stack[T]) Push(value T) {
	s.data.AppendOne(value)
}

// Pop removes and returns the top element from the stack.
// Returns zero value and false if the stack is empty.
func (s *Stack[T]) Pop() (T, bool) {
	return s.data.Pop()
}

// Peek returns the top element without removing it.
// Returns zero value and false if the stack is empty.
func (s *Stack[T]) Peek() (T, bool) {
	if s.data.Len() == 0 {
		var zero T
		return zero, false
	}
	slice := s.data.Slice()
	return slice[len(slice)-1], true
}

// Len returns the number of elements in the stack.
func (s *Stack[T]) Len() int {
	return s.data.Len()
}

// Cap returns the capacity of the stack.
func (s *Stack[T]) Cap() int {
	return s.data.Cap()
}

// IsEmpty reports whether the stack is empty.
func (s *Stack[T]) IsEmpty() bool {
	return s.data.Len() == 0
}

// Clear removes all elements from the stack.
func (s *Stack[T]) Clear() {
	s.data.Clear()
}
