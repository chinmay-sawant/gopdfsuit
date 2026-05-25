# Pass 1 Implementation Blueprints

Before/after code for Pass 1 optimizations. See [IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md) for status.

---

## P1-01: Zero-Alloc Text Encoding

### Before (`utils.go`)

```go
func formatTextForPDF(resolvedName string, text string, registry *CustomFontRegistry) string {
    if isCustomFontCheck(resolvedName, registry) {
        return EncodeTextForCustomFont(resolvedName, text, registry)
    }
    return "(" + escapePDFString(text) + ")"
}
```

### After

```go
func appendTextForPDF(dst []byte, resolvedName, text string, registry *CustomFontRegistry) []byte {
    if registry.HasFont(resolvedName) {
        return AppendTextForCustomFont(dst, resolvedName, text, registry)
    }
    dst = append(dst, '(')
    dst = appendEscapedPDFString(dst, text)
    return append(dst, ')')
}
```

Call sites change from:

```go
textPosBuf = append(textPosBuf, formatTextForPDF(resolvedName, cell.Text, registry)...)
```

To:

```go
textPosBuf = appendTextForPDF(textPosBuf[:0], resolvedName, cell.Text, registry)
```

---

## P1-02: RuneSet Bitmap

### Before (`registry.go`)

```go
UsedChars map[rune]bool

func (r *CustomFontRegistry) MarkCharsUsed(name string, text string) {
    for _, char := range text {
        font.UsedChars[char] = true
    }
}
```

### After (`runeset.go`)

```go
type RuneSet struct {
    bits   [65536 / 64]uint64
    astral map[rune]struct{}
}

func (s *RuneSet) Add(r rune) { /* bit-set or astral map */ }
func (s *RuneSet) Len() int { /* popcount + astral len */ }
func (s *RuneSet) Range(fn func(rune)) { /* iterate used runes */ }
```

---

## P1-04: Extend Zlib Pool

### Before (`subset.go`)

```go
func CompressFontData(data []byte) ([]byte, error) {
    var buf bytes.Buffer
    w := zlib.NewWriter(&buf)
    // ...
}
```

### After

```go
func CompressFontData(data []byte) ([]byte, error) {
    buf := GetCompressBuffer()
    w := GetZlibWriter(buf)
    defer PutZlibWriter(w)
    defer CompressBufPool.Put(buf)
    // ...
    return append([]byte(nil), buf.Bytes()...), nil
}
```

---

## P1-05: Pre-Grow Page Buffers

### Before (`pagemanager.go`)

```go
pm.ContentStreams = append(pm.ContentStreams, bytes.Buffer{})
```

### After

```go
var b bytes.Buffer
b.Grow(32 * 1024)
pm.ContentStreams = append(pm.ContentStreams, b)
```

---

## P1-07: Respect noLock

### Before (`registry.go`)

```go
func (r *CustomFontRegistry) GenerateSubsets() error {
    r.mu.Lock()
    defer r.mu.Unlock()
    // ...
}
```

### After

```go
func (r *CustomFontRegistry) GenerateSubsets() error {
    if !r.noLock {
        r.mu.Lock()
        defer r.mu.Unlock()
    }
    // ...
}
```
