package container

import (
	"github.com/thebagchi/arena-go"
)

// Queue is a generic FIFO (First In First Out) data structure.
type Queue[T any] struct {
	data *Vec[T]
	head int // index of the first element
}

// NewQueue creates a new empty queue backed by the provided arena.
func NewQueue[T any](a *arena.Arena) *Queue[T] {
	return &Queue[T]{
		data: NewVec[T](a),
		head: 0,
	}
}

// Enqueue adds an element to the back of the queue.
func (q *Queue[T]) Enqueue(value T) {
	q.data.AppendOne(value)
}

// Dequeue removes and returns the front element from the queue.
// Returns zero value and false if the queue is empty.
func (q *Queue[T]) Dequeue() (T, bool) {
	if q.head >= q.data.Len() {
		var zero T
		return zero, false
	}

	slice := q.data.Slice()
	value := slice[q.head]
	q.head++

	// Compact the queue if head has moved too far
	if q.head > q.data.Len()/2 && q.head > 16 {
		// Shift remaining elements to the front
		remaining := slice[q.head:]
		q.data.Clear()
		q.data.AppendSlice(remaining)
		q.head = 0
	}

	return value, true
}

// Peek returns the front element without removing it.
// Returns zero value and false if the queue is empty.
func (q *Queue[T]) Peek() (T, bool) {
	if q.head >= q.data.Len() {
		var zero T
		return zero, false
	}
	return q.data.Slice()[q.head], true
}

// Len returns the number of elements in the queue.
func (q *Queue[T]) Len() int {
	return q.data.Len() - q.head
}

// Cap returns the capacity of the underlying storage.
func (q *Queue[T]) Cap() int {
	return q.data.Cap()
}

// IsEmpty reports whether the queue is empty.
func (q *Queue[T]) IsEmpty() bool {
	return q.head >= q.data.Len()
}

// Clear removes all elements from the queue.
func (q *Queue[T]) Clear() {
	q.data.Clear()
	q.head = 0
}
