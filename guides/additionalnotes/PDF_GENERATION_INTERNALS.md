# PDF Generation Internals: How gopdfsuit Handles Assets

This guide explains the internal architecture of the `gopdfsuit` PDF generation engine, specifically detailing how different asset types (SVGs, borders, fonts, and images) are processed and stored in the generated PDF files.

The implementation follows professional best practices for high-performance, native PDF generation.

## 1. SVGs (Vector Graphics)

**How it works:**
The engine does **NOT** store the raw `.svg` file, nor does it rasterize the image into pixels. Instead, it parses the SVG XML and translates it into native PDF vector commands.

- **Implementation Location:** `internal/pdf/svg.go`
- **Process:**
  1.  Parses the SVG paths (`d` attributes, `<rect>`, `<circle>`, etc.).
  2.  Translates them into PDF drawing operators:
      - `m` (move to)
      - `l` (line to)
      - `c` (bezier curve)
      - `re` (rectangle)
  3.  Writes these commands directly to the PDF content stream.

**Benefits:**
- **Infinite Scalability:** Graphics remain crisp at any zoom level.
- **Small File Size:** Only mathematical path instructions are stored, avoiding large raster bitmaps.
- **Native Rendering:** Works on all PDF viewers without needing special SVG support.

## 2. Table Borders & Drawing Operations

**How it works:**
Borders and shapes are drawn procedurally using vector commands within the page stream.

- **Implementation Location:** `internal/pdf/draw.go`
- **Process:**
  - The engine calculates coordinates for each cell or element.
  - It writes vector drawing commands (`m`, `l`, `S` for stroke, `f` for fill) directly into the page stream.
  
**Benefits:**
- **Precision:** Borders are mathematically defined lines.
- **Efficiency:** Takes up negligible space (just a few bytes of instructions).
- **Quality:** Lines appear perfectly sharp regardless of resolution.

## 3. Fonts

**How it works:**
The engine ensures document portability and text searchability by handling fonts natively.

- **Implementation Location:** `internal/pdf/fontregistry.go` and `internal/pdf/ttf.go`
- **Process:**
  - **Standard Fonts (Helvetica, Times, Courier):** Only references are stored, relying on the viewer's standard fonts.
  - **Custom Fonts:** The **actual binary font file** (TTF/OTF) is embedded directly into the PDF object stream.
  
**Benefits:**
- **Portability:** Documents look identical on all devices, even if the user doesn't have the specific font installed.
- **Searchable/Selectable:** Text is stored as meaningful character codes, not pixels, allowing users to search and copy text.

## 4. Images (Raster Graphics)

**How it works:**
The engine optimizes image storage based on the file format to balance performance and quality.

- **Implementation Location:** `internal/pdf/image.go`
- **Process:**
  - **JPEGs:** The raw bytes are detected and passed **directly** into the PDF stream (~`DCTDecode`).
    - *Why?* Prevents re-compression artifacts and is extremely fast (zero processing overhead).
  - **PNGs:**
    - Decoded to raw RGB/RGBA.
    - Verified for transparency.
    - Compressed using `FlateDecode` (Zip).
    - Deduplicated using hashing (if the same image appears multiple times, it's stored once and referenced multiple times).

**Benefits:**
- **High Performance:** minimal processing for JPEGs.
- **Optimized Storage:** Shared resources for duplicate images.
- **Quality Preservation:** No generational loss for JPEGs; efficient lossless compression for PNGs.

---

**Summary:**
`gopdfsuit` is a **native PDF generator**. It builds documents from the ground up using the PDF specification's primitives (vectors, text objects, streams), rather than taking screenshots or rasterizing HTML. This results in professional-grade documents that are searchable, scalable, and highly portable.
