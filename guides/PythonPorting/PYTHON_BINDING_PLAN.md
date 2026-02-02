# Plan: Python Bindings for gopdfsuit

## Overview

This document outlines the plan to create Python bindings for the `gopdfsuit` Go package, enabling Python developers to call Go functions directly from Python code.

---

## 1. Approach Options Analysis

### Option A: CGO + ctypes/cffi (Recommended)
**How it works:** Compile Go code to a C shared library (`.so`/`.dll`) using `cgo`, then call it from Python using `ctypes` or `cffi`.

**Pros:**
- Direct function calls with near-native performance
- No network overhead
- Works offline
- Single binary distribution

**Cons:**
- Complex memory management between Go and Python
- Need to handle Go's garbage collector
- Platform-specific builds required

### Option B: gRPC
**How it works:** Run Go as a separate service, Python communicates via gRPC.

**Pros:**
- Clean separation of concerns
- Language-agnostic protocol
- Well-defined interfaces

**Cons:**
- Requires running a separate service
- Network overhead (even for localhost)
- More complex deployment

### Option C: REST API
**How it works:** Already exists in `gopdfsuit` - just call the HTTP endpoints.

**Pros:**
- Already implemented
- Simple to use

**Cons:**
- Network overhead
- Requires server to be running

### Option D: PyBind + CGO (via intermediate C++)
**How it works:** Create C++ wrapper around Go's C exports, then use pybind11.

**Pros:**
- Native Python module

**Cons:**
- Extra complexity layer
- Requires C++ knowledge

---

## 2. Recommended Approach: CGO + Python Wrapper

We'll use **Option A** with a clean Python wrapper package. This provides the best balance of performance, simplicity, and developer experience.

---

## 3. Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                          Python Layer                             │
├──────────────────────────────────────────────────────────────────┤
│  pypdfsuit/                                                       │
│  ├── __init__.py          # Main exports                         │
│  ├── types.py             # Pydantic/dataclass models            │
│  ├── generator.py         # PDF generation                       │
│  ├── merge.py             # PDF merging                          │
│  ├── split.py             # PDF splitting                        │
│  ├── fill.py              # Form filling                         │
│  ├── html.py              # HTML conversion                      │
│  └── _bindings.py         # ctypes bindings to shared lib        │
├──────────────────────────────────────────────────────────────────┤
│                         C Shared Library                          │
├──────────────────────────────────────────────────────────────────┤
│  libgopdfsuit.so (Linux) / libgopdfsuit.dylib (macOS) /          │
│  gopdfsuit.dll (Windows)                                          │
│  - Exported C functions via cgo                                   │
├──────────────────────────────────────────────────────────────────┤
│                           Go Layer                                │
├──────────────────────────────────────────────────────────────────┤
│  pkg/gopdflib/            # Existing Go library                   │
│  internal/pdf/            # Core PDF implementation               │
└──────────────────────────────────────────────────────────────────┘
```

---

## 4. Implementation Tasks

### Phase 1: Create C Exports (Go Side)

#### Task 1.1: Create `cgo_exports.go`
Create a new file that exports Go functions as C functions.

**Location:** `bindings/python/cgo/exports.go`

```go
package main

/*
#include <stdlib.h>

typedef struct {
    char* data;
    int length;
    char* error;
} ByteResult;

typedef struct {
    char** data;
    int* lengths;
    int count;
    char* error;
} ByteArrayResult;
*/
import "C"
import (
    "unsafe"
    "encoding/json"
    "github.com/chinmay-sawant/gopdfsuit/v4/pkg/gopdflib"
)

//export GeneratePDF
func GeneratePDF(jsonTemplate *C.char) C.ByteResult {
    // Implementation
}

//export MergePDFs
func MergePDFs(pdfData **C.char, pdfLengths *C.int, count C.int) C.ByteResult {
    // Implementation
}

//export SplitPDF
func SplitPDF(pdfData *C.char, pdfLength C.int, specJSON *C.char) C.ByteArrayResult {
    // Implementation
}

//export FillPDFWithXFDF
func FillPDFWithXFDF(pdfData *C.char, pdfLen C.int, xfdfData *C.char, xfdfLen C.int) C.ByteResult {
    // Implementation
}

//export ConvertHTMLToPDF
func ConvertHTMLToPDF(requestJSON *C.char) C.ByteResult {
    // Implementation
}

//export ConvertHTMLToImage
func ConvertHTMLToImage(requestJSON *C.char) C.ByteResult {
    // Implementation
}

//export FreeBytesResult
func FreeBytesResult(result C.ByteResult) {
    // Free allocated memory
}

//export FreeBytesArrayResult
func FreeBytesArrayResult(result C.ByteArrayResult) {
    // Free allocated memory
}

func main() {}
```

#### Task 1.2: Create Build Script
**Location:** `bindings/python/build.sh`

```bash
#!/bin/bash
# Build shared library for current platform

CGO_ENABLED=1 go build -buildmode=c-shared \
    -o libgopdfsuit.so \
    ./bindings/python/cgo/
```

---

### Phase 2: Create Python Package

#### Task 2.1: Package Structure
```
bindings/python/pypdfsuit/
├── pyproject.toml
├── setup.py
├── README.md
├── pypdfsuit/
│   ├── __init__.py
│   ├── types.py
│   ├── _bindings.py
│   ├── generator.py
│   ├── merge.py
│   ├── split.py
│   ├── fill.py
│   ├── html.py
│   └── lib/
│       ├── libgopdfsuit.so      # Linux
│       ├── libgopdfsuit.dylib   # macOS
│       └── gopdfsuit.dll        # Windows
└── tests/
    ├── test_generator.py
    ├── test_merge.py
    ├── test_split.py
    └── test_html.py
```

#### Task 2.2: Types Module (`types.py`)
Define Python equivalents of Go types using Pydantic or dataclasses.

```python
from dataclasses import dataclass, field
from typing import Optional, List, Dict, Any

@dataclass
class Config:
    page: str = "A4"
    page_alignment: int = 1  # 1=Portrait, 2=Landscape
    page_border: str = ""
    watermark: Optional[str] = None
    pdf_title: Optional[str] = None
    security: Optional["SecurityConfig"] = None
    pdfa: Optional["PDFAConfig"] = None
    signature: Optional["SignatureConfig"] = None
    embed_fonts: Optional[bool] = True
    custom_fonts: List["CustomFontConfig"] = field(default_factory=list)

@dataclass
class SecurityConfig:
    enabled: bool = False
    user_password: str = ""
    owner_password: str = ""
    allow_printing: bool = True
    allow_modifying: bool = False
    allow_copying: bool = True
    allow_annotations: bool = False
    # ... more fields

@dataclass
class Cell:
    props: str
    text: str = ""
    checkbox: Optional[bool] = None
    image: Optional["Image"] = None
    bgcolor: Optional[str] = None
    text_color: Optional[str] = None
    link: Optional[str] = None

@dataclass
class Row:
    row: List[Cell]

@dataclass
class Table:
    maxcolumns: int
    rows: List[Row]
    column_widths: Optional[List[float]] = None
    row_heights: Optional[List[float]] = None
    bgcolor: Optional[str] = None
    text_color: Optional[str] = None

@dataclass
class Element:
    type: str  # "table", "spacer", "image"
    table: Optional[Table] = None
    spacer: Optional["Spacer"] = None
    image: Optional["Image"] = None

@dataclass
class PDFTemplate:
    config: Config
    title: "Title"
    elements: List[Element] = field(default_factory=list)
    footer: Optional["Footer"] = None
    bookmarks: List["Bookmark"] = field(default_factory=list)

# ... additional type definitions for all models
```

#### Task 2.3: Bindings Module (`_bindings.py`)
Low-level ctypes interface to the shared library.

```python
import ctypes
from ctypes import c_char_p, c_int, POINTER, Structure
import platform
import os

class ByteResult(Structure):
    _fields_ = [
        ("data", c_char_p),
        ("length", c_int),
        ("error", c_char_p)
    ]

class ByteArrayResult(Structure):
    _fields_ = [
        ("data", POINTER(c_char_p)),
        ("lengths", POINTER(c_int)),
        ("count", c_int),
        ("error", c_char_p)
    ]

def _load_library():
    """Load the appropriate shared library for the current platform."""
    system = platform.system()
    lib_dir = os.path.dirname(__file__)
    
    if system == "Linux":
        lib_path = os.path.join(lib_dir, "lib", "libgopdfsuit.so")
    elif system == "Darwin":
        lib_path = os.path.join(lib_dir, "lib", "libgopdfsuit.dylib")
    elif system == "Windows":
        lib_path = os.path.join(lib_dir, "lib", "gopdfsuit.dll")
    else:
        raise OSError(f"Unsupported platform: {system}")
    
    return ctypes.CDLL(lib_path)

_lib = _load_library()

# Configure function signatures
_lib.GeneratePDF.argtypes = [c_char_p]
_lib.GeneratePDF.restype = ByteResult

_lib.MergePDFs.argtypes = [POINTER(c_char_p), POINTER(c_int), c_int]
_lib.MergePDFs.restype = ByteResult

_lib.SplitPDF.argtypes = [c_char_p, c_int, c_char_p]
_lib.SplitPDF.restype = ByteArrayResult

_lib.FillPDFWithXFDF.argtypes = [c_char_p, c_int, c_char_p, c_int]
_lib.FillPDFWithXFDF.restype = ByteResult

_lib.ConvertHTMLToPDF.argtypes = [c_char_p]
_lib.ConvertHTMLToPDF.restype = ByteResult

_lib.ConvertHTMLToImage.argtypes = [c_char_p]
_lib.ConvertHTMLToImage.restype = ByteResult

_lib.FreeBytesResult.argtypes = [ByteResult]
_lib.FreeBytesResult.restype = None

_lib.FreeBytesArrayResult.argtypes = [ByteArrayResult]
_lib.FreeBytesArrayResult.restype = None
```

#### Task 2.4: High-Level API Modules

**generator.py:**
```python
from .types import PDFTemplate
from ._bindings import _lib, ByteResult
import json

def generate_pdf(template: PDFTemplate) -> bytes:
    """
    Generate a PDF from a template.
    
    Args:
        template: PDFTemplate object with configuration and content
        
    Returns:
        bytes: The generated PDF file content
        
    Raises:
        RuntimeError: If PDF generation fails
    """
    template_json = json.dumps(template.to_dict()).encode('utf-8')
    result = _lib.GeneratePDF(template_json)
    
    try:
        if result.error:
            raise RuntimeError(result.error.decode('utf-8'))
        return result.data[:result.length]
    finally:
        _lib.FreeBytesResult(result)
```

**merge.py:**
```python
def merge_pdfs(pdf_files: list[bytes]) -> bytes:
    """
    Merge multiple PDFs into a single document.
    
    Args:
        pdf_files: List of PDF file contents as bytes
        
    Returns:
        bytes: The merged PDF file content
    """
    # Implementation using ctypes
    pass
```

**split.py:**
```python
from dataclasses import dataclass
from typing import Optional, List

@dataclass
class SplitSpec:
    pages: Optional[List[int]] = None
    ranges: Optional[List[tuple[int, int]]] = None
    max_per_file: Optional[int] = None

def split_pdf(pdf_data: bytes, spec: SplitSpec) -> list[bytes]:
    """Split a PDF into multiple parts."""
    pass

def parse_page_spec(spec: str, total_pages: int = 0) -> list[int]:
    """Parse a page specification string like '1-3,5,7-9'."""
    pass
```

**fill.py:**
```python
def fill_pdf_with_xfdf(pdf_data: bytes, xfdf_data: bytes) -> bytes:
    """Fill a PDF form with XFDF data."""
    pass
```

**html.py:**
```python
from dataclasses import dataclass
from typing import Optional, Dict

@dataclass
class HtmlToPDFRequest:
    html: Optional[str] = None
    url: Optional[str] = None
    page_size: str = "A4"
    orientation: str = "Portrait"
    margin_top: str = "10mm"
    margin_right: str = "10mm"
    margin_bottom: str = "10mm"
    margin_left: str = "10mm"

@dataclass
class HtmlToImageRequest:
    html: Optional[str] = None
    url: Optional[str] = None
    format: str = "png"
    width: Optional[int] = None
    height: Optional[int] = None

def convert_html_to_pdf(request: HtmlToPDFRequest) -> bytes:
    """Convert HTML to PDF."""
    pass

def convert_html_to_image(request: HtmlToImageRequest) -> bytes:
    """Convert HTML to image."""
    pass
```

---

### Phase 3: Memory Management

#### Task 3.1: Implement Safe Memory Handling
Go's garbage collector runs independently. We need to ensure:

1. **Pin Go memory** during C calls using `runtime.Pinner`
2. **Copy data** to C-allocated memory for return values
3. **Provide explicit free functions** for Python to call

```go
//export GeneratePDF
func GeneratePDF(jsonTemplate *C.char) C.ByteResult {
    var result C.ByteResult
    
    goTemplate := C.GoString(jsonTemplate)
    var template gopdflib.PDFTemplate
    if err := json.Unmarshal([]byte(goTemplate), &template); err != nil {
        result.error = C.CString(err.Error())
        return result
    }
    
    pdfBytes, err := gopdflib.GeneratePDF(template)
    if err != nil {
        result.error = C.CString(err.Error())
        return result
    }
    
    // Allocate C memory and copy data
    result.data = (*C.char)(C.malloc(C.size_t(len(pdfBytes))))
    C.memcpy(unsafe.Pointer(result.data), unsafe.Pointer(&pdfBytes[0]), C.size_t(len(pdfBytes)))
    result.length = C.int(len(pdfBytes))
    
    return result
}

//export FreeBytesResult
func FreeBytesResult(result C.ByteResult) {
    if result.data != nil {
        C.free(unsafe.Pointer(result.data))
    }
    if result.error != nil {
        C.free(unsafe.Pointer(result.error))
    }
}
```

---

### Phase 4: Build & Distribution

#### Task 4.1: Multi-Platform Build
Create GitHub Actions workflow to build for all platforms:

```yaml
# .github/workflows/build-python.yml
name: Build Python Bindings

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        
    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          
      - name: Build shared library
        run: |
          cd bindings/python
          ./build.sh  # Platform-specific build
          
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: lib-${{ matrix.os }}
          path: bindings/python/pypdfsuit/pypdfsuit/lib/
```

#### Task 4.2: Python Package Distribution
**pyproject.toml:**
```toml
[build-system]
requires = ["setuptools>=61.0", "wheel"]
build-backend = "setuptools.build_meta"

[project]
name = "pypdfsuit"
version = "1.0.0"
description = "Python bindings for gopdfsuit - PDF generation, merging, splitting, and more"
readme = "README.md"
requires-python = ">=3.8"
license = {text = "MIT"}
authors = [
    {name = "Chinmay Sawant", email = "chinmay@example.com"}
]
keywords = ["pdf", "generation", "merge", "split", "html"]
classifiers = [
    "Development Status :: 4 - Beta",
    "Intended Audience :: Developers",
    "License :: OSI Approved :: MIT License",
    "Programming Language :: Python :: 3",
    "Programming Language :: Python :: 3.8",
    "Programming Language :: Python :: 3.9",
    "Programming Language :: Python :: 3.10",
    "Programming Language :: Python :: 3.11",
    "Programming Language :: Python :: 3.12",
]

[project.urls]
Homepage = "https://github.com/chinmay-sawant/gopdfsuit"
Documentation = "https://github.com/chinmay-sawant/gopdfsuit/tree/main/bindings/python"

[tool.setuptools.packages.find]
where = ["."]

[tool.setuptools.package-data]
pypdfsuit = ["lib/*"]
```

---

### Phase 5: Testing

#### Task 5.1: Unit Tests
```python
# tests/test_generator.py
import pytest
from pypdfsuit import generate_pdf, PDFTemplate, Config, Title

def test_basic_generation():
    template = PDFTemplate(
        config=Config(page="A4", page_alignment=1),
        title=Title(props="Helvetica:18:100:center:0:0:0:0", text="Test"),
        elements=[]
    )
    pdf_bytes = generate_pdf(template)
    assert pdf_bytes.startswith(b'%PDF-')

def test_merge():
    from pypdfsuit import merge_pdfs
    # Create two simple PDFs
    pdf1 = generate_pdf(...)
    pdf2 = generate_pdf(...)
    merged = merge_pdfs([pdf1, pdf2])
    assert merged.startswith(b'%PDF-')
```

#### Task 5.2: Integration Tests
Test against sample data in `sampledata/` directory.

---

## 5. Directory Structure Summary

```
gopdfsuit/
├── bindings/
│   └── python/
│       ├── cgo/
│       │   ├── exports.go      # CGO exports
│       │   └── main.go         # Main for shared lib
│       ├── pypdfsuit/
│       │   ├── __init__.py
│       │   ├── types.py
│       │   ├── _bindings.py
│       │   ├── generator.py
│       │   ├── merge.py
│       │   ├── split.py
│       │   ├── fill.py
│       │   ├── html.py
│       │   └── lib/            # Compiled shared libraries
│       ├── tests/
│       ├── pyproject.toml
│       ├── setup.py
│       └── README.md
│       └── build.sh
├── pkg/gopdflib/               # Existing Go library
├── internal/                   # Existing internal packages
└── ...
```

---

## 6. API Mapping

| Go Function | Python Function |
|-------------|-----------------|
| `gopdflib.GeneratePDF(template)` | `pypdfsuit.generate_pdf(template)` |
| `gopdflib.GetAvailableFonts()` | `pypdfsuit.get_available_fonts()` |
| `gopdflib.GetFontRegistry()` | `pypdfsuit.get_font_registry()` |
| `gopdflib.MergePDFs(files)` | `pypdfsuit.merge_pdfs(files)` |
| `gopdflib.SplitPDF(file, spec)` | `pypdfsuit.split_pdf(file, spec)` |
| `gopdflib.ParsePageSpec(spec, total)` | `pypdfsuit.parse_page_spec(spec, total)` |
| `gopdflib.FillPDFWithXFDF(pdf, xfdf)` | `pypdfsuit.fill_pdf_with_xfdf(pdf, xfdf)` |
| `gopdflib.ConvertHTMLToPDF(req)` | `pypdfsuit.convert_html_to_pdf(req)` |
| `gopdflib.ConvertHTMLToImage(req)` | `pypdfsuit.convert_html_to_image(req)` |

---

## 7. Timeline Estimate

| Phase | Task | Effort |
|-------|------|--------|
| 1 | CGO Exports | 2-3 days |
| 2 | Python Package | 3-4 days |
| 3 | Memory Management | 1-2 days |
| 4 | Build & Distribution | 1-2 days |
| 5 | Testing | 2-3 days |
| **Total** | | **9-14 days** |

---

## 8. Usage Examples (Post-Implementation)

### Basic PDF Generation
```python
from pypdfsuit import generate_pdf, PDFTemplate, Config, Title, Element, Table, Row, Cell

# Create a simple PDF
template = PDFTemplate(
    config=Config(page="A4", page_alignment=1),
    title=Title(
        props="Helvetica:24:100:center:0:0:0:0",
        text="My Document"
    ),
    elements=[
        Element(
            type="table",
            table=Table(
                maxcolumns=2,
                column_widths=[1.0, 1.0],
                rows=[
                    Row(row=[
                        Cell(props="Helvetica:12:100:left:1:1:1:1", text="Name"),
                        Cell(props="Helvetica:12:000:left:1:1:1:1", text="John Doe"),
                    ])
                ]
            )
        )
    ]
)

pdf_bytes = generate_pdf(template)
with open("output.pdf", "wb") as f:
    f.write(pdf_bytes)
```

### Merge PDFs
```python
from pypdfsuit import merge_pdfs

with open("doc1.pdf", "rb") as f1, open("doc2.pdf", "rb") as f2:
    merged = merge_pdfs([f1.read(), f2.read()])

with open("merged.pdf", "wb") as f:
    f.write(merged)
```

### HTML to PDF
```python
from pypdfsuit import convert_html_to_pdf, HtmlToPDFRequest

request = HtmlToPDFRequest(
    html="<html><body><h1>Hello World</h1></body></html>",
    page_size="A4",
    orientation="Portrait"
)
pdf_bytes = convert_html_to_pdf(request)
```

---

## 9. Alternative: gopy (Automated Approach)

If manual CGO is too complex, consider using **gopy** which auto-generates Python bindings:

```bash
go install github.com/go-python/gopy@latest
gopy build -output=pypdfsuit -vm=python3 github.com/chinmay-sawant/gopdfsuit/v4/pkg/gopdflib
```

**Pros:** Less manual work
**Cons:** Less control, may not handle all types properly

---

## 10. Decision Points

1. **Package name:** `pypdfsuit` or `gopdfsuit-python`?
2. **Minimum Python version:** 3.8+ recommended
3. **Type hints:** Use Pydantic models or Python dataclasses?
4. **Distribution:** PyPI only or also conda-forge?
5. **Thread safety:** Document Go's GC behavior or add Python-side locking?
