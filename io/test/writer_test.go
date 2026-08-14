package arena_test

import (
	"testing"

	"github.com/gubsky90/arena-go"
	"github.com/gubsky90/arena-go/alloc"
	iopackage "github.com/gubsky90/arena-go/io"
)

func TestWriter(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	defer a.Delete()
	w := iopackage.NewWriter(a)

	// Test Write
	n, err := w.Write([]byte("hello"))
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != 5 {
		t.Errorf("Write: expected 5 bytes written, got %d", n)
	}

	// Test WriteString
	n, err = w.WriteString(" world")
	if err != nil {
		t.Errorf("WriteString failed: %v", err)
	}
	if n != 6 {
		t.Errorf("WriteString: expected 6 bytes written, got %d", n)
	}

	// Test WriteByte
	err = w.WriteByte('!')
	if err != nil {
		t.Errorf("WriteByte failed: %v", err)
	}

	// Test Bytes
	expected := "hello world!"
	actual := string(w.Bytes())
	if actual != expected {
		t.Errorf("Bytes: expected %q, got %q", expected, actual)
	}

	// Test Len
	if w.Len() != 12 {
		t.Errorf("Len: expected 12, got %d", w.Len())
	}

	// Test Cap
	if w.Cap() < w.Len() {
		t.Errorf("Cap: capacity %d should be >= len %d", w.Cap(), w.Len())
	}

	// Test Reset
	w.Reset()
	if w.Len() != 0 {
		t.Errorf("Reset: expected len 0, got %d", w.Len())
	}
	if w.Cap() == 0 {
		t.Errorf("Reset: capacity should remain after reset")
	}

	// Test growth
	w.Reset()
	largeData := make([]byte, 1000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}
	n, err = w.Write(largeData)
	if err != nil {
		t.Errorf("Write large data failed: %v", err)
	}
	if n != 1000 {
		t.Errorf("Write large data: expected 1000 bytes written, got %d", n)
	}
	if w.Len() != 1000 {
		t.Errorf("Write large data: expected len 1000, got %d", w.Len())
	}
	if len(w.Bytes()) != 1000 {
		t.Errorf("Write large data: expected bytes len 1000, got %d", len(w.Bytes()))
	}
}

func TestWriter_MultipleWrites(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	defer a.Delete()

	w := iopackage.NewWriter(a)

	// Multiple sequential writes
	writes := [][]byte{
		[]byte("hello"),
		[]byte(" "),
		[]byte("world"),
		[]byte("!"),
	}

	totalBytes := 0
	for _, data := range writes {
		n, err := w.Write(data)
		if err != nil || n != len(data) {
			t.Errorf("Write failed: expected %d bytes, got %d, err: %v", len(data), n, err)
		}
		totalBytes += n
	}

	if w.Len() != totalBytes || string(w.Bytes()) != "hello world!" {
		t.Errorf("Multiple writes: expected 'hello world!' with len %d, got %q with len %d", totalBytes, string(w.Bytes()), w.Len())
	}
}

func TestWriter_WriteStrings(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	defer a.Delete()

	w := iopackage.NewWriter(a)

	strings := []string{
		"The quick ",
		"brown fox ",
		"jumps over ",
		"the lazy dog",
	}

	expectedLen := 0
	for _, s := range strings {
		n, err := w.WriteString(s)
		if err != nil || n != len(s) {
			t.Errorf("WriteString failed: expected %d bytes, got %d", len(s), n)
		}
		expectedLen += len(s)
	}

	if w.Len() != expectedLen {
		t.Errorf("Expected total len %d, got %d", expectedLen, w.Len())
	}

	if string(w.Bytes()) != "The quick brown fox jumps over the lazy dog" {
		t.Errorf("String mismatch: got %q", string(w.Bytes()))
	}
}

func TestWriter_WriteBytes(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	defer a.Delete()

	w := iopackage.NewWriter(a)

	// Write individual bytes
	for i := 0; i < 10; i++ {
		err := w.WriteByte(byte(65 + i)) // A-J
		if err != nil {
			t.Errorf("WriteByte failed at %d: %v", i, err)
		}
	}

	if w.Len() != 10 {
		t.Errorf("Expected len 10, got %d", w.Len())
	}

	if string(w.Bytes()) != "ABCDEFGHIJ" {
		t.Errorf("Expected 'ABCDEFGHIJ', got %q", string(w.Bytes()))
	}
}

func TestWriter_ResetAndReuse(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	defer a.Delete()

	w := iopackage.NewWriter(a)

	// First write
	w.Write([]byte("first"))
	if w.Len() != 5 {
		t.Errorf("First write: expected len 5, got %d", w.Len())
	}

	// Reset and second write
	w.Reset()
	if w.Len() != 0 {
		t.Errorf("After reset: expected len 0, got %d", w.Len())
	}

	w.Write([]byte("second"))
	if w.Len() != 6 {
		t.Errorf("Second write: expected len 6, got %d", w.Len())
	}
	if string(w.Bytes()) != "second" {
		t.Errorf("Expected 'second', got %q", string(w.Bytes()))
	}

	// Reset again for third write
	w.Reset()
	w.WriteString("third")
	if string(w.Bytes()) != "third" {
		t.Errorf("Third write: expected 'third', got %q", string(w.Bytes()))
	}
}

func TestWriter_CapacityGrowth(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(10 * 4096))
	defer a.Delete()

	w := iopackage.NewWriter(a)
	initialCap := w.Cap()

	// Write more than initial capacity
	largeData := make([]byte, 100)
	for i := range largeData {
		largeData[i] = 'X'
	}

	w.Write(largeData)
	if w.Cap() <= initialCap {
		t.Errorf("Expected capacity to grow from %d to > %d", initialCap, initialCap)
	}

	if w.Len() != 100 {
		t.Errorf("Expected len 100, got %d", w.Len())
	}
}

func TestWriter_MixedOperations(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	defer a.Delete()

	w := iopackage.NewWriter(a)

	// Mix Write, WriteString, WriteByte
	w.Write([]byte("Start:"))
	w.WriteByte(' ')
	w.WriteString("Hello")
	w.WriteByte(' ')
	w.Write([]byte("World"))
	w.WriteByte('!')

	expected := "Start: Hello World!"
	if string(w.Bytes()) != expected {
		t.Errorf("Mixed operations: expected %q, got %q", expected, string(w.Bytes()))
	}

	if w.Len() != len(expected) {
		t.Errorf("Expected len %d, got %d", len(expected), w.Len())
	}
}

func TestWriter_EmptyWrites(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	defer a.Delete()

	w := iopackage.NewWriter(a)

	// Empty write should not affect length
	n, err := w.Write([]byte{})
	if n != 0 || err != nil {
		t.Errorf("Empty write: expected 0 bytes and no error, got %d and %v", n, err)
	}

	if w.Len() != 0 {
		t.Errorf("Empty write: expected len 0, got %d", w.Len())
	}

	// Write something
	w.Write([]byte("data"))
	if w.Len() != 4 {
		t.Errorf("After data: expected len 4, got %d", w.Len())
	}

	// Another empty write
	w.Write([]byte{})
	if w.Len() != 4 {
		t.Errorf("After second empty write: expected len 4, got %d", w.Len())
	}
}

func TestWriter_BytesSliceBehavior(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	defer a.Delete()

	w := iopackage.NewWriter(a)

	w.Write([]byte("test"))
	bytes1 := w.Bytes()
	if string(bytes1) != "test" {
		t.Errorf("Expected 'test', got %q", string(bytes1))
	}

	// Add more data
	w.Write([]byte("ing"))
	bytes2 := w.Bytes()
	if string(bytes2) != "testing" {
		t.Errorf("Expected 'testing', got %q", string(bytes2))
	}

	// Reset and write new data
	w.Reset()
	w.WriteString("new")
	bytes3 := w.Bytes()
	if string(bytes3) != "new" {
		t.Errorf("Expected 'new', got %q", string(bytes3))
	}
}

func TestWriter_VeryLargeData(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(100 * 4096))
	defer a.Delete()

	w := iopackage.NewWriter(a)

	// Write 10KB of data
	largeData := make([]byte, 10000)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	n, err := w.Write(largeData)
	if n != 10000 || err != nil {
		t.Errorf("Write large: expected 10000 bytes, got %d, err: %v", n, err)
	}

	if w.Len() != 10000 {
		t.Errorf("Expected len 10000, got %d", w.Len())
	}

	// Verify data integrity
	written := w.Bytes()
	for i := 0; i < 10000; i++ {
		if written[i] != largeData[i] {
			t.Errorf("Data mismatch at index %d: expected %d, got %d", i, largeData[i], written[i])
			break
		}
	}
}

func TestWriter_SequentialByteWrites(t *testing.T) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	defer a.Delete()

	w := iopackage.NewWriter(a)

	// Write ASCII printable characters
	for i := 32; i < 127; i++ {
		err := w.WriteByte(byte(i))
		if err != nil {
			t.Errorf("WriteByte %d failed: %v", i, err)
		}
	}

	if w.Len() != 95 { // 127 - 32
		t.Errorf("Expected len 95, got %d", w.Len())
	}

	// Check content length
	if len(w.Bytes()) != 95 {
		t.Errorf("Expected bytes len 95, got %d", len(w.Bytes()))
	}
}
