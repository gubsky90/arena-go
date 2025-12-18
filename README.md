# arena-go

A high-performance, zero-GC memory allocator library for Go with generic containers. Allocate memory outside the garbage collector using custom allocators (Bump, Slab, Buddy) and work with type-safe generic containers.

## Features

- **Zero-GC Allocations**: Allocate memory outside Go's garbage collector using arena-based memory management
- **Multiple Allocators**: Choose from Bump (fastest), Slab (fixed-size), or Buddy (flexible) allocation strategies
- **Generic Containers**: Type-safe generic containers including Vec (dynamic arrays), Map, SkipList, Pool, and Str
- **Thread-Safe**: All allocators and containers are safe for concurrent use
- **Deterministic Memory Management**: Explicit control over memory lifecycle with Reset() and Delete()

## Installation

```bash
go get github.com/thebagchi/arena-go
```

## Quick Start

### Creating an Arena

```go
package main

import (
    "github.com/thebagchi/arena-go"
    "github.com/thebagchi/arena-go/alloc"
)

func main() {
    // Create an arena with 4KB of memory using Bump allocator
    a := arena.New(alloc.NewBumpAllocator(4096))
    defer a.Delete()
    
    // Use the arena...
}
```

## Allocators

### Bump Allocator
Fastest allocator, best for batch allocations or when the arena is reset frequently.

```go
a := arena.New(alloc.NewBumpAllocator(1024 * 4096)) // 4MB
```

### Slab Allocator
Best for fixed-size objects with high allocation/free turnover. Features a multi-tiered design with 17 size classes (16B to 1MB), exhausted/available slab tracking, and in-object free lists for efficient memory management.

```go
a := arena.New(alloc.NewSlabAllocator())
```

**Architecture highlights:**
- 17 size classes for objects from 16 bytes to 1MB
- Per-size-class bins with exhausted and available slab lists
- In-object free lists for O(1) allocation/deallocation
- Cross-bin free page pool for efficient memory recycling

### Buddy Allocator
Most flexible, good for varied-size allocations with power-of-2 sizes.

```go
a := arena.New(alloc.NewBuddyAllocator(1024 * 4096))
```

## Core Operations

### Allocating Objects

```go
import "github.com/thebagchi/arena-go"

// Allocate a single integer
ptr := arena.Alloc[int](a)
*ptr = 42

// Allocate and initialize a struct
type Person struct {
    Name string
    Age  int
}

person := arena.Ptr(a, Person{Name: "Alice", Age: 30})

// Create a string in the arena
str := a.MakeString("hello world")
```

### Allocating Slices

```go
// Allocate a slice with length 10, capacity 20
slice := arena.MakeSlice[int](a, 10, 20)
slice[0] = 100
```

## Generic Containers

### Vec - Dynamic Array

```go
import "github.com/thebagchi/arena-go/container"

// Create a dynamic array
vec := container.NewVec[int](a)

// Append elements
vec.Append(1, 2, 3)
vec.AppendOne(4)
vec.AppendSlice([]int{5, 6, 7})

// Access elements
fmt.Println(vec.Len())      // 7
fmt.Println(vec.Cap())      // capacity
fmt.Println(vec.Slice())    // []int{1, 2, 3, 4, 5, 6, 7}

// Pop elements
if val, ok := vec.Pop(); ok {
    fmt.Println(val) // 7
}

// Search
idx := vec.IndexOf(3)
fmt.Println(idx) // 2

// Clear
vec.Clear()
fmt.Println(vec.Len()) // 0
```

### List - Generic Doubly-Linked List

```go
import "github.com/thebagchi/arena-go/alloc/cont"

// Create a list (works with any type)
list := cont.NewList[int]()

// Add elements
list.PushBack(1)
list.PushBack(2)
list.PushFront(0)

// Iterate
for it := list.Iter(); ; {
    val, ok := it.Next()
    if !ok {
        break
    }
    fmt.Println(val)
}

// Remove elements
if node := list.Front(); node != nil {
    list.Remove(node)
}

// Insert operations
node := list.Front()
list.InsertAfter(10, node)  // Insert after node
list.InsertBefore(5, node)  // Insert before node

// Move operations
list.MoveToFront(node)      // Move to front
list.MoveToBack(node)       // Move to back
```

### Map - Type-Safe Hash Map

```go
// Create a map
m := container.NewMap[string, int](a)

// Set values
m.Set("alice", 30)
m.Set("bob", 25)

// Get values
if age, found := m.Get("alice"); found {
    fmt.Println(age) // 30
}

// Check existence
if m.Contains("charlie") {
    fmt.Println("Found charlie")
}

// Iterate
m.Range(func(key string, value int) bool {
    fmt.Printf("%s: %d\n", key, value)
    return true
})

// Delete
m.Delete("bob")

// Get length
fmt.Println(m.Len()) // 1
```

### SkipList - Ordered Map

```go
// Create a skip list (ordered by key)
sl := container.NewSkipList[int, string](a)

// Insert elements
sl.Insert(10, "ten")
sl.Insert(5, "five")
sl.Insert(15, "fifteen")

// Search
if val, found := sl.Search(5); found {
    fmt.Println(val) // "five"
}

// Iterate in order
sl.Range(func(key int, value string) bool {
    fmt.Printf("%d: %s\n", key, value)
    return true
})

// Delete
sl.Delete(5)
```

### Pool - Object Pool

```go
// Create an object pool for efficient allocation/deallocation
pool := container.NewPool[Person](a)

// Allocate from pool
p := pool.Alloc()
p.Name = "Alice"
p.Age = 30

// Return to pool for reuse
pool.Free(p)

// Next allocation will reuse the same memory (zeroed)
p2 := pool.Alloc()
fmt.Println(p2.Name) // "" (zeroed)
fmt.Println(p2.Age)  // 0 (zeroed)

// Check pool stats
fmt.Println(pool.Len()) // number in free list
fmt.Println(pool.Cap()) // total capacity
```

### Str - String Utilities

```go
// Create a string builder
str := container.NewStr(a)

// Build strings
str.WriteString("Hello, ")
str.WriteString("World!")

// Get result
result := str.String()
fmt.Println(result)

// Or as bytes
bytes := str.Bytes()
```

## I/O Operations

### Writer

```go
import arenaio "github.com/thebagchi/arena-go/io"

// Create a writer backed by arena memory
writer := arenaio.NewWriter(a)

// Write data
writer.Write([]byte("hello"))
writer.WriteString(" world")
writer.WriteByte('!')

// Get result
fmt.Println(string(writer.Bytes())) // "hello world!"

// Check stats
fmt.Println(writer.Len()) // bytes written
fmt.Println(writer.Cap()) // capacity

// Reset for reuse
writer.Reset()
```

### Reader

```go
// Create a reader
data := []byte("hello world")
reader := arenaio.NewReader(a, data)

// Read data
buf := make([]byte, 5)
n, err := reader.Read(buf)
fmt.Println(string(buf[:n])) // "hello"

// Check position
fmt.Println(reader.Len())  // bytes remaining
fmt.Println(reader.Size()) // total size

// Reset to beginning
reader.Reset()
```

## Memory Management

### Resetting an Arena

Clears allocations but retains underlying memory pages:

```go
a.Reset()
```

### Deleting an Arena

Frees all memory back to the OS:

```go
a.Delete()
```

### Checking Memory Ownership

```go
// Check if a pointer was allocated by this arena
if arena.OwnsPtr(a, ptr) {
    fmt.Println("Pointer belongs to this arena")
}

// Check if a slice was allocated by this arena
if arena.OwnsSlice(a, slice) {
    fmt.Println("Slice belongs to this arena")
}
```

## Performance Tips

1. **Choose the right allocator**: Use Bump for batch allocations, Slab for fixed-size objects, Buddy for mixed sizes
2. **Use pools for frequent allocations**: Object pools reduce allocation overhead for frequently created/destroyed objects
3. **Pre-allocate containers**: Specify capacity when creating containers if you know the size
4. **Reset vs Delete**: Use Reset() for batch operations, Delete() only when completely done
5. **Arena per operation**: Create separate arenas for independent operations and delete them when done

## Thread Safety

All allocators and containers are thread-safe:

```go
var wg sync.WaitGroup
a := arena.New(alloc.NewBumpAllocator(1024 * 4096))
defer a.Delete()

vec := container.NewVec[int](a)

for i := 0; i < 10; i++ {
    wg.Add(1)
    go func(n int) {
        defer wg.Done()
        vec.Append(n)
    }(i)
}

wg.Wait()
fmt.Println(vec.Len()) // 10
```

## Examples

See the `example/main.go` file for more comprehensive examples covering all features.

Run the example:

```bash
go run ./example/main.go
```

## Package Structure

- `arena.go` - Core arena functionality and allocator interface
- `object.go` - Basic object allocation utilities
- `io/` - I/O operations (Reader, Writer)
- `container/` - Generic containers (Vec, Map, SkipList, Pool, Str)
- `alloc/` - Allocator implementations
  - `bump.go` - Bump allocator (linear allocation)
  - `slab.go` - Slab allocator (fixed-size object pools with 17 size classes)
  - `buddy.go` - Buddy allocator (power-of-2 flexible allocation)
  - `cont/` - Internal data structures (List, Cont)
- `res/` - Resource management (page allocation)

## License

See LICENSE file for details.