# PDF Library Benchmark Report

**Date:** February 1, 2026  
**Hardware:** 13th Gen Intel(R) Core(TM) i7-13700HX  
**Dataset:** 2,000 user records with ID, Name, Email, Role, and Description fields  
**Test:** Generate a styled table report with headers and alternating row colors

## Results Summary

| Library | Language | Time (ms) | File Size | PDF Standard | Table Support |
|---------|----------|-----------|-----------|--------------|---------------|
| **GoPDFSuit** | Go | **126.99** | 1.7 MB | PDF/A-4 (PDF 2.0) | Full tables with styling + wrap |
| jsPDF | JavaScript | 56.20 | 279 KB | PDF 1.3 | Text fallback (no autotable) |
| PDFKit | JavaScript | 305.71 | 205 KB | PDF 1.3 | Full tables with styling |
| Typst | Typst | 774.58 | 221 KB | PDF/A-2b | Full tables with styling |
| pdf-lib | JavaScript | 674.16 | 313 KB | PDF 1.7 | Manual table drawing |
| FPDF2 | Python | 2,834.42 | 203 KB | PDF/A-1b compatible | Full tables with styling |

## GoPDFSuit Statistical Summary (10 runs)

| Metric | Value |
| :--- | :--- |
| **Minimum** | **122.97 ms** |
| **Maximum** | 131.39 ms |
| **Average (Mean)** | **126.99 ms** |
| **Throughput** | ~7.9 docs/sec |

*\*jsPDF running in text-only fallback mode (simpler workload).*

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
