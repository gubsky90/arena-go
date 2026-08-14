package io

import (
	"io"

	arena "github.com/gubsky90/arena-go"
)

// Reader provides a way to read bytes from a buffer
// associated with an arena.
type Reader struct {
	arena  *arena.Arena
	buffer []byte
	offset int
}

// NewReader creates a new Reader with an arena and buffer.
func NewReader(a *arena.Arena, data []byte) *Reader {
	return &Reader{
		arena:  a,
		buffer: data,
		offset: 0,
	}
}

// Read reads up to len(p) bytes into p. It returns the number of bytes
// read (0 <= n <= len(p)) and any error encountered.
func (r *Reader) Read(p []byte) (n int, err error) {
	if r.offset >= len(r.buffer) {
		return 0, io.EOF
	}
	n = copy(p, r.buffer[r.offset:])
	r.offset = r.offset + n
	return n, nil
}

// Len returns the number of bytes remaining to be read.
func (r *Reader) Len() int {
	return len(r.buffer) - r.offset
}

// Size returns the original length of the buffer.
func (r *Reader) Size() int {
	return len(r.buffer)
}

// Reset resets the reader to the beginning of the buffer.
func (r *Reader) Reset() {
	r.offset = 0
}
