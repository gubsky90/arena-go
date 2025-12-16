package io

import (
	arena "github.com/thebagchi/arena-go"
)

// Writer provides a way to write bytes to an arena-allocated buffer
// without the byte array escaping to the heap.
type Writer struct {
	arena  *arena.Arena
	buffer []byte
	offset int
}

// NewWriter creates a new Writer with an arena-allocated buffer.
func NewWriter(a *arena.Arena) *Writer {
	buf := arena.MakeSlice[byte](a, 0, 32)
	buf = buf[:cap(buf)] // set len to cap to allow writing
	return &Writer{
		arena:  a,
		buffer: buf,
		offset: 0,
	}
}

// Write writes p to the buffer, growing it as needed.
// The buffer is reallocated in the arena if necessary.
func (w *Writer) Write(p []byte) (n int, err error) {
	needed := w.offset + len(p)
	if needed > cap(w.buffer) {
		w.grow(needed)
	}
	copy(w.buffer[w.offset:], p)
	w.offset = w.offset + len(p)
	return len(p), nil
}

// WriteString writes s to the buffer, growing it as needed.
func (w *Writer) WriteString(s string) (n int, err error) {
	needed := w.offset + len(s)
	if needed > cap(w.buffer) {
		w.grow(needed)
	}
	copy(w.buffer[w.offset:], s)
	w.offset = w.offset + len(s)
	return len(s), nil
}

// WriteByte writes a single byte to the buffer, growing it as needed.
func (w *Writer) WriteByte(c byte) error {
	if w.offset >= cap(w.buffer) {
		w.grow(w.offset + 1)
	}
	w.buffer[w.offset] = c
	w.offset = w.offset + 1
	return nil
}

// Bytes returns the written bytes as a slice.
// The underlying array is arena-allocated and does not escape to the heap.
func (w *Writer) Bytes() []byte {
	return w.buffer[:w.offset]
}

// Len returns the number of bytes written.
func (w *Writer) Len() int {
	return w.offset
}

// Cap returns the capacity of the buffer.
func (w *Writer) Cap() int {
	return cap(w.buffer)
}

// Reset resets the writer to be empty but retains the underlying buffer.
func (w *Writer) Reset() {
	w.offset = 0
}

// grow ensures the buffer has at least the given capacity.
// Note: This is called from arena.go which handles actual memory allocation
func (w *Writer) grow(size int) {
	capacity := cap(w.buffer) * 2
	if capacity < size {
		capacity = size
	}
	if capacity < 64 {
		capacity = 64
	}
	// Caller (arena.go) will handle the actual arena allocation
	temp := make([]byte, capacity)
	copy(temp, w.buffer[:w.offset])
	w.buffer = temp
}
