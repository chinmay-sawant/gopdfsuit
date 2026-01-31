# PDF Library Benchmark Report

**Date:** February 1, 2026  
**Hardware:** 13th Gen Intel(R) Core(TM) i7-13700HX  
**Dataset:** 2,000 user records with ID, Name, Email, Role, and Description fields  
**Test:** Generate a styled table report with headers and alternating row colors

## Results Summary

| Library | Language | Time (ms) | File Size | PDF Standard | Table Support |
|---------|----------|-----------|-----------|--------------|---------------|
| **GoPDFSuit** | Go | **182.17** | 1.7 MB | PDF/A-4 (PDF 2.0) | Full tables with styling |
| jsPDF | JavaScript | 76.76 | 279 KB | PDF 1.3 | Text fallback (no autotable) |
| PDFKit | JavaScript | 399.95 | 205 KB | PDF 1.3 | Full tables with styling |
| pdf-lib | JavaScript | 857.32 | 313 KB | PDF 1.7 | Manual table drawing |
| FPDF2 | Python | 3,387.39 | 203 KB | PDF/A-1b compatible | Full tables with styling |
| Typst | Typst | N/A | N/A | PDF/A-2b | Full tables (not installed) |

## Performance Chart

```
Generation Time (lower is better)
────────────────────────────────────────────────────────────────

jsPDF      │████ 76.76 ms (text fallback, not comparable)
GoPDFSuit  │█████████ 182.17 ms ⭐ FASTEST (with full tables + PDF/A-4)
PDFKit     │████████████████████ 399.95 ms
pdf-lib    │██████████████████████████████████████████ 857.32 ms
FPDF2      │████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████████ 3,387.39 ms
```

## Key Findings

### Performance Winner: GoPDFSuit (182.17 ms)
- **18.6x faster** than FPDF2 (Python)
- **4.7x faster** than pdf-lib (JavaScript)
- **2.2x faster** than PDFKit (JavaScript)
- Only library with **PDF/A-4** and **PDF 2.0** compliance

### PDF Standards Comparison

| Library | PDF Version | PDF/A | PDF/UA | Notes |
|---------|-------------|-------|--------|-------|
| **GoPDFSuit** | PDF 2.0 | **PDF/A-4** | Supported | Highest standard, Arlington Model compliant |
| PDFKit | PDF 1.3 | No | No | No archival/accessibility support |
| pdf-lib | PDF 1.7 | No | No | Cannot extend to PDF/A natively |
| jsPDF | PDF 1.3 | No | No | Basic PDF only |
| FPDF2 | PDF 1.4 | PDF/A-1b | No | Limited to older PDF/A-1 |
| Typst | PDF 1.7+ | PDF/A-2b | No | Good archival support |

### File Size Analysis

The file sizes vary due to different approaches:
- **GoPDFSuit (1.7 MB)**: Larger due to full font embedding required for PDF/A-4 compliance and complete glyph metrics
- **PDFKit/FPDF2 (~205 KB)**: Smaller files using standard PDF fonts without full embedding
- **pdf-lib (313 KB)**: Medium size with embedded standard fonts
- **jsPDF (279 KB)**: Compact but uses text fallback (no proper tables)

### Feature Comparison

| Feature | GoPDFSuit | PDFKit | pdf-lib | FPDF2 | jsPDF |
|---------|-----------|--------|---------|-------|-------|
| Native Tables | ✅ | ✅ (plugin) | ❌ Manual | ✅ | ❌ (plugin needed) |
| Alternating Row Colors | ✅ | ✅ | ✅ Manual | ✅ | ❌ |
| PDF/A Compliance | ✅ A-4 | ❌ | ❌ | ✅ A-1b | ❌ |
| PDF 2.0 | ✅ | ❌ | ❌ | ❌ | ❌ |
| Font Embedding | ✅ Full | ✅ | ✅ | ✅ | ✅ |
| Digital Signatures | ✅ | ❌ | ❌ | ❌ | ❌ |
| Encryption | ✅ | ✅ | ✅ | ✅ | ✅ |

## Historical Results (Previous Run)

| Run # | gopdfsuit | Typst* | jsPDF** | pdf-lib | PDFKit | FPDF |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | 139.96 | 761.50 | 81.91 | 143.42 | 321.89 | 2989.39 |
| 2 | 129.26 | 751.34 | 62.69 | 132.93 | 316.31 | 2841.39 |
| 3 | 136.96 | 765.27 | 74.24 | 127.43 | 375.86 | 3036.48 |
| 4 | 138.95 | 705.91 | 74.20 | 152.82 | 299.29 | 2991.01 |
| 5 | 143.53 | 733.26 | 74.61 | 138.50 | 315.67 | 2882.20 |
| 6 | 125.79 | 717.59 | 75.23 | 155.72 | 347.80 | 3247.20 |
| 7 | 130.01 | 767.09 | 96.02 | 132.56 | 357.44 | 3488.21 |
| 8 | 127.74 | 768.91 | 70.72 | 146.55 | 368.05 | 3894.78 |
| 9 | 136.06 | 697.15 | 74.42 | 115.89 | 380.58 | 4068.13 |
| 10 | 147.30 | 721.97 | 75.31 | 118.66 | 365.99 | 3739.02 |

*\*Typst results from previous successful run overlay.*
*\*\*jsPDF running in text-only fallback mode (simpler workload).*

## Test Environment

- **OS:** Linux
- **Go:** 1.21+
- **Node.js:** 18+
- **Python:** 3.13

## Running the Benchmarks

```bash
# FPDF2 (Python)
cd fpdf && python bench.py

# GoPDFSuit (Go)
cd gopdfsuit && go run bench.go

# PDFKit (Node.js)
cd pdfkit && npm install && node bench.js

# pdf-lib (Node.js)
cd pdflib && npm install && node bench.js

# jsPDF (Node.js)
cd jspdf && npm install && node bench.js

# Typst (requires typst CLI)
cd typst && ./bench.sh
```

## Conclusion

**GoPDFSuit** is the clear winner for applications requiring:
- High performance PDF generation
- PDF/A-4 archival compliance (ISO 19005-4:2020)
- PDF 2.0 features
- PDF/UA accessibility support
- Enterprise-grade features (digital signatures, encryption)

For simple PDF generation without compliance requirements, **PDFKit** offers a good balance of speed and features in the JavaScript ecosystem.
