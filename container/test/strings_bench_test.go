package container_test

import (
	"strings"
	"testing"

	arena "github.com/gubsky90/arena-go"
	"github.com/gubsky90/arena-go/alloc"
	"github.com/gubsky90/arena-go/container"
	"github.com/gubsky90/arena-go/res"
)

var (
	benchStr       = "  hello world from arena allocator  "
	benchLongStr   = strings.Repeat("hello world ", 100)
	benchSubstr    = "world"
	benchSep       = " "
	benchCutset    = " "
	benchOld       = "world"
	benchNew       = "arena"
	benchSplitStr  = "a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,q,r,s,t,u,v,w,x,y,z"
	benchFieldsStr = "hello world from arena allocator with many fields here"
)

// ToBytes/ToString Benchmarks
func BenchmarkString_StdStringToBytes(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = []byte(benchStr)
	}
}

func BenchmarkString_ZeroCopyToBytes(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = res.UnsafeBytes(benchStr)
	}
}

func BenchmarkString_StdBytesToString(b *testing.B) {
	data := []byte(benchStr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = string(data)
	}
}

func BenchmarkString_ZeroCopyToString(b *testing.B) {
	data := []byte(benchStr)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = res.UnsafeString(data)
	}
}

// TrimSpace Benchmarks
func BenchmarkString_StdTrimSpace(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.TrimSpace(benchStr)
	}
}

func BenchmarkString_ZeroCopyTrimSpace(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.TrimSpace(benchStr)
	}
}

// Contains Benchmarks
func BenchmarkString_StdContains(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Contains(benchStr, benchSubstr)
	}
}

func BenchmarkString_ZeroCopyContains(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Contains(benchStr, benchSubstr)
	}
}

// HasPrefix Benchmarks
func BenchmarkString_StdHasPrefix(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.HasPrefix(benchStr, "  hello")
	}
}

func BenchmarkString_ZeroCopyHasPrefix(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.HasPrefix(benchStr, "  hello")
	}
}

// HasSuffix Benchmarks
func BenchmarkString_StdHasSuffix(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.HasSuffix(benchStr, "  ")
	}
}

func BenchmarkString_ZeroCopyHasSuffix(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.HasSuffix(benchStr, "  ")
	}
}

// Index Benchmarks
func BenchmarkString_StdIndex(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Index(benchStr, benchSubstr)
	}
}

func BenchmarkString_ZeroCopyIndex(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Index(benchStr, benchSubstr)
	}
}

// LastIndex Benchmarks
func BenchmarkString_StdLastIndex(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.LastIndex(benchStr, benchSubstr)
	}
}

func BenchmarkString_ZeroCopyLastIndex(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.LastIndex(benchStr, benchSubstr)
	}
}

// Trim Benchmarks
func BenchmarkString_StdTrim(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Trim(benchStr, benchCutset)
	}
}

func BenchmarkString_ZeroCopyTrim(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Trim(benchStr, benchCutset)
	}
}

// TrimLeft Benchmarks
func BenchmarkString_StdTrimLeft(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.TrimLeft(benchStr, benchCutset)
	}
}

func BenchmarkString_ZeroCopyTrimLeft(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.TrimLeft(benchStr, benchCutset)
	}
}

// TrimRight Benchmarks
func BenchmarkString_StdTrimRight(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.TrimRight(benchStr, benchCutset)
	}
}

func BenchmarkString_ZeroCopyTrimRight(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.TrimRight(benchStr, benchCutset)
	}
}

// EqualFold Benchmarks
func BenchmarkString_StdEqualFold(b *testing.B) {
	str1 := "Hello World"
	str2 := "hello world"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.EqualFold(str1, str2)
	}
}

func BenchmarkString_ZeroCopyEqualFold(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	str1 := "Hello World"
	str2 := "hello world"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.EqualFold(str1, str2)
	}
}

// Compare Benchmarks
func BenchmarkString_StdCompare(b *testing.B) {
	str1 := "apple"
	str2 := "banana"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Compare(str1, str2)
	}
}

func BenchmarkString_ZeroCopyCompare(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	str1 := "apple"
	str2 := "banana"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Compare(str1, str2)
	}
}

// ToLower Benchmarks
func BenchmarkString_StdToLower(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.ToLower(benchStr)
	}
}

func BenchmarkString_ZeroCopyToLower(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.ToLower(benchStr)
	}
}

// ToUpper Benchmarks
func BenchmarkString_StdToUpper(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.ToUpper(benchStr)
	}
}

func BenchmarkString_ZeroCopyToUpper(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.ToUpper(benchStr)
	}
}

// Title Benchmarks
func BenchmarkString_StdTitle(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Title(benchStr)
	}
}

func BenchmarkString_ZeroCopyTitle(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Title(benchStr)
	}
}

// Split Benchmarks
func BenchmarkString_StdSplit(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Split(benchSplitStr, ",")
	}
}

func BenchmarkString_ArenaSplit(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Split(benchSplitStr, ",")
		a.Reset()
	}
}

// Join Benchmarks
func BenchmarkString_StdJoin(b *testing.B) {
	parts := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Join(parts, ",")
	}
}

func BenchmarkString_ArenaJoin(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	parts := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Join(parts, ",")
		a.Reset()
	}
}

// Fields Benchmarks
func BenchmarkString_StdFields(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Fields(benchFieldsStr)
	}
}

func BenchmarkString_ArenaFields(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Fields(benchFieldsStr)
		a.Reset()
	}
}

// Count Benchmarks
func BenchmarkString_StdCount(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Count(benchLongStr, benchSubstr)
	}
}

func BenchmarkString_ZeroCopyCount(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Count(benchLongStr, benchSubstr)
	}
}

// Replace Benchmarks
func BenchmarkString_StdReplace(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Replace(benchLongStr, benchOld, benchNew, -1)
	}
}

func BenchmarkString_ArenaReplace(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Replace(benchLongStr, benchOld, benchNew, -1)
		a.Reset()
	}
}

// ReplaceAll Benchmarks
func BenchmarkString_StdReplaceAll(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.ReplaceAll(benchLongStr, benchOld, benchNew)
	}
}

func BenchmarkString_ArenaReplaceAll(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.ReplaceAll(benchLongStr, benchOld, benchNew)
		a.Reset()
	}
}

// Repeat Benchmarks
func BenchmarkString_StdRepeat(b *testing.B) {
	testStr := "hello"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Repeat(testStr, 10)
	}
}

func BenchmarkString_ArenaRepeat(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	testStr := "hello"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Repeat(testStr, 10)
		a.Reset()
	}
}

// TrimPrefix Benchmarks
func BenchmarkString_StdTrimPrefix(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.TrimPrefix(benchStr, "  hello")
	}
}

func BenchmarkString_ZeroCopyTrimPrefix(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.TrimPrefix(benchStr, "  hello")
	}
}

// TrimSuffix Benchmarks
func BenchmarkString_StdTrimSuffix(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.TrimSuffix(benchStr, "  ")
	}
}

func BenchmarkString_ZeroCopyTrimSuffix(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.TrimSuffix(benchStr, "  ")
	}
}

// Cut Benchmarks
func BenchmarkString_StdCut(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = strings.Cut(benchStr, benchSubstr)
	}
}

func BenchmarkString_ZeroCopyCut(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = str.Cut(benchStr, benchSubstr)
	}
}

// IndexByte Benchmarks
func BenchmarkString_StdIndexByte(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.IndexByte(benchStr, 'w')
	}
}

func BenchmarkString_ZeroCopyIndexByte(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.IndexByte(benchStr, 'w')
	}
}

// ContainsAny Benchmarks
func BenchmarkString_StdContainsAny(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.ContainsAny(benchStr, "xyz")
	}
}

func BenchmarkString_ZeroCopyContainsAny(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.ContainsAny(benchStr, "xyz")
	}
}

// Long String Benchmarks (to test scalability)
func BenchmarkString_StdSplitLong(b *testing.B) {
	longStr := strings.Repeat("word,", 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Split(longStr, ",")
	}
}

func BenchmarkString_ArenaSplitLong(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	longStr := strings.Repeat("word,", 1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Split(longStr, ",")
		a.Reset()
	}
}

// Memory allocation comparison benchmarks
func BenchmarkString_StdSplitAllocs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = strings.Split(benchSplitStr, ",")
	}
}

func BenchmarkString_ArenaSplitAllocs(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = str.Split(benchSplitStr, ",")
		a.Reset()
	}
}

func BenchmarkString_StdFieldsAllocs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = strings.Fields(benchFieldsStr)
	}
}

func BenchmarkString_ArenaFieldsAllocs(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = str.Fields(benchFieldsStr)
		a.Reset()
	}
}

// Lines Benchmarks
func BenchmarkString_Lines(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	linesStr := strings.Repeat("This is a line of text\n", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		count := 0
		for _ = range str.Lines(linesStr) {
			count++
		}
		_ = count
	}
}

func BenchmarkString_StdLines(b *testing.B) {
	linesStr := strings.Repeat("This is a line of text\n", 100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Split(strings.TrimSuffix(linesStr, "\n"), "\n")
	}
}

// Clone Benchmarks
func BenchmarkString_ArenaClone(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.Clone(benchStr)
		a.Reset()
	}
}

// FieldsFunc Benchmarks
func BenchmarkString_StdFieldsFunc(b *testing.B) {
	isSpace := func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.FieldsFunc(benchFieldsStr, isSpace)
	}
}

func BenchmarkString_ArenaFieldsFunc(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	isSpace := func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.FieldsFunc(benchFieldsStr, isSpace)
		a.Reset()
	}
}

// ContainsFunc Benchmarks
func BenchmarkString_ContainsFunc(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	isDigit := func(r rune) bool { return r >= '0' && r <= '9' }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.ContainsFunc(benchStr, isDigit)
	}
}

// IndexFunc Benchmarks
func BenchmarkString_IndexFunc(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	isSpace := func(r rune) bool { return r == ' ' }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.IndexFunc(benchStr, isSpace)
	}
}

// LastIndexFunc Benchmarks
func BenchmarkString_LastIndexFunc(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	isSpace := func(r rune) bool { return r == ' ' }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.LastIndexFunc(benchStr, isSpace)
	}
}

// MapString Benchmarks
func BenchmarkString_StdMap(b *testing.B) {
	toUpper := func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 32
		}
		return r
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.Map(toUpper, benchStr)
	}
}

func BenchmarkString_ArenaMapASCII(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	toUpper := func(c byte) int {
		if c >= 'a' && c <= 'z' {
			return int(c - 32)
		}
		return int(c)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.MapASCII(toUpper, benchStr)
	}
}

func BenchmarkString_ArenaMapUTF8(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	toUpper := func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 32
		}
		return r
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.MapUTF8(toUpper, benchStr)
	}
}

func BenchmarkString_ArenaMapString(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	toUpper := func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 32
		}
		return r
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.MapString(toUpper, benchStr)
	}
}

// ToTitle Benchmarks
func BenchmarkString_StdToTitle(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.ToTitle(benchStr)
	}
}

func BenchmarkString_ArenaToTitle(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.ToTitle(benchStr)
		a.Reset()
	}
}

// ToValidUTF8 Benchmarks
func BenchmarkString_StdToValidUTF8(b *testing.B) {
	invalidStr := "hello\xffworld\xfe"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.ToValidUTF8(invalidStr, "?")
	}
}

func BenchmarkString_ArenaToValidUTF8(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	invalidStr := "hello\xffworld\xfe"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.ToValidUTF8(invalidStr, "?")
		a.Reset()
	}
}

// TrimFunc Benchmarks
func BenchmarkString_StdTrimFunc(b *testing.B) {
	isSpace := func(r rune) bool { return r == ' ' || r == '\t' }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.TrimFunc(benchStr, isSpace)
	}
}

func BenchmarkString_TrimFunc(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	isSpace := func(r rune) bool { return r == ' ' || r == '\t' }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.TrimFunc(benchStr, isSpace)
	}
}

// TrimLeftFunc Benchmarks
func BenchmarkString_StdTrimLeftFunc(b *testing.B) {
	isSpace := func(r rune) bool { return r == ' ' || r == '\t' }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.TrimLeftFunc(benchStr, isSpace)
	}
}

func BenchmarkString_TrimLeftFunc(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	isSpace := func(r rune) bool { return r == ' ' || r == '\t' }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.TrimLeftFunc(benchStr, isSpace)
	}
}

// TrimRightFunc Benchmarks
func BenchmarkString_StdTrimRightFunc(b *testing.B) {
	isSpace := func(r rune) bool { return r == ' ' || r == '\t' }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = strings.TrimRightFunc(benchStr, isSpace)
	}
}

func BenchmarkString_TrimRightFunc(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	isSpace := func(r rune) bool { return r == ' ' || r == '\t' }
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = str.TrimRightFunc(benchStr, isSpace)
	}
}

// Allocation comparison benchmarks for new functions
func BenchmarkString_ArenaCloneAllocs(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = str.Clone(benchStr)
		a.Reset()
	}
}

func BenchmarkString_ArenaFieldsFuncAllocs(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	isSpace := func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' }
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = str.FieldsFunc(benchFieldsStr, isSpace)
		a.Reset()
	}
}

func BenchmarkString_ArenaMapStringAllocs(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	toUpper := func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 32
		}
		return r
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = str.MapString(toUpper, benchStr)
	}
}

func BenchmarkString_ArenaToTitleAllocs(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = str.ToTitle(benchStr)
		a.Reset()
	}
}

func BenchmarkString_ArenaToValidUTF8Allocs(b *testing.B) {
	a := arena.New(alloc.NewBumpAllocator(1 * 4096))
	str := container.NewStr(a)
	defer a.Delete()
	invalidStr := "hello\xffworld\xfe"
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = str.ToValidUTF8(invalidStr, "?")
		a.Reset()
	}
}
