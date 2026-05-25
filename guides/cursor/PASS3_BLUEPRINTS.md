# Pass 3 Implementation Blueprints

Before/after code for Pass 3 optimizations.

---

## P3-01: Allocation-Free WrapText

### Before

```go
words := strings.Fields(text)
testLine += " " + word  // string concat per iteration
runes := []rune(word)
lines = append(lines, string(runes[start:end]))
```

### After

```go
type WrapState struct {
    lines   [][]byte
    buf     []byte
    wordBuf []byte
}

lines := WrapTextInto(&wrapState, text, fontName, fontSize, maxWidth, registry)
// Reuse wrapState across rows in drawTable
```

---

## P3-02: ExtraObjects as []byte

### Before

```go
ExtraObjects map[int]string
pdfBuffer.WriteString(content)
```

### After

```go
ExtraObjects map[int][]byte
pdfBuffer.Write(content)
bytes.Contains(content, []byte("/Subtype /Widget"))
```

---

## P3-03: Redact Parser Unification

### Before

```go
objMap map[string][]byte  // keys "3 0"
re := regexp.MustCompile(`(?s)(\d+) (\d+) obj.*?endobj`)
```

### After

```go
objMap map[int][]byte
objGen map[int]int
boundaries := merge.FindObjectBoundaries(pdfBytes)
```

---

## P3-05: Typed StructKid

### Before

```go
Kids []interface{}
append(elem.Kids, mcid)
kid.(*StructElem)
```

### After

```go
Kids []StructKid
append(elem.Kids, StructKid{MCID: mcid})
kid.Elem != nil
```
