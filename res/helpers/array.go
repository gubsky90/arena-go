package helpers

type Array[T any] struct {
	items []T
}

func NewArray[T any](capacity int) *Array[T] {
	if capacity < 8 {
		capacity = 8
	}
	return &Array[T]{items: make([]T, 0, capacity)}
}

func (a *Array[T]) Push(item T) {
	a.items = append(a.items, item)
}

func (a *Array[T]) Get(index int) T {
	if index < 0 || index >= len(a.items) {
		var zero T
		return zero
	}
	return a.items[index]
}

// AIterator allows iteration over the array.
type AIterator[T any] struct {
	array *Array[T]
	index int
}

// Iter returns an iterator starting from index 0.
func (a *Array[T]) Iter() *AIterator[T] {
	return &AIterator[T]{
		array: a,
		index: 0,
	}
}

// Next advances the iterator and returns the next value.
// Returns zero value and false when iteration is done.
func (it *AIterator[T]) Next() (T, bool) {
	if it.index >= it.array.Len() {
		var zero T
		return zero, false
	}

	value := it.array.Get(it.index)
	it.index++
	return value, true
}

func (a *Array[T]) Len() int {
	return len(a.items)
}

func (a *Array[T]) Reset() {
	a.items = a.items[:0]
}

// Set sets the value at the given index, growing the array if necessary.
func (a *Array[T]) Set(index int, value T) {
	if index < 0 {
		return
	}
	if index >= len(a.items) {
		// Grow the slice to fit the index
		newItems := make([]T, index+1)
		copy(newItems, a.items)
		a.items = newItems
	}
	a.items[index] = value
}

// Resize resizes the array to the given length, growing with zero values if needed.
func (a *Array[T]) Resize(length int) {
	if length < 0 {
		return
	}
	if length > cap(a.items) {
		newItems := make([]T, length)
		copy(newItems, a.items)
		a.items = newItems
	} else {
		a.items = a.items[:length]
	}
}

// Ptr returns a pointer to the element at the given index.
func (a *Array[T]) Ptr(index int) *T {
	if index < 0 || index >= len(a.items) {
		return nil
	}
	return &a.items[index]
}
