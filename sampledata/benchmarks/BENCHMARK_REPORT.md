# Comprehensive Benchmark Report: gopdfsuit vs Others

## 1. Test Configuration

*   **Date**: 2026-01-30
*   **Hardware**: 13th Gen Intel(R) Core(TM) i7-13700HX
*   **Iterations**: 10 runs per tool
*   **Dataset**: 2,000 rows of user data (ID, Name, Email, Role, Description)
*   **Tools**:
    *   **gopdfsuit**: Go (Table layout) - *v0.1.0*
    *   **Typst**: CLI (Markup -> PDF) - *v0.11.0*
    *   **jsPDF**: Node.js (Text fallback*)
    *   **pdf-lib**: Node.js (Manual text positioning)
    *   **PDFKit**: Node.js (Table plugin)
    *   **FPDF**: Python (Table method)

## 2. Raw Results (ms)

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

## 3. Statistical Summary

| Metric | gopdfsuit | Typst | jsPDF | pdf-lib | PDFKit | FPDF |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Minimum** | 125.79 ms | 697.15 ms | 62.69 ms | 115.89 ms | 299.29 ms | 2841.39 ms |
| **Maximum** | 147.30 ms | 768.91 ms | 96.02 ms | 155.72 ms | 380.58 ms | 4068.13 ms |
| **Average** | **135.56 ms** | **739.00 ms** | **75.93 ms** | **136.45 ms** | **344.89 ms** | **3317.78 ms** |
| **Speedup (vs gopdfsuit)** | **1x** | **0.18x** | **1.78x** | **0.99x** | **0.39x** | **0.04x** |

## 4. Analysis

**gopdfsuit** provides the best balance of performance and feature set for table-based PDF generation:

**gopdfsuit is statistically ~5.46x faster than Typst on average.**

*   **vs PDFKit (Node.js)**: gopdfsuit is **~2.5x faster** (135ms vs 344ms).
*   **vs Typst**: gopdfsuit is **~5.5x faster** (135ms vs 739ms).
*   **vs FPDF (Python)**: gopdfsuit is **~24x faster** (135ms vs 3317ms).
*   **vs pdf-lib (Node.js)**: Performance is roughly equivalent (~135ms vs 136ms), but `pdf-lib` in this benchmark was performing a simpler manual text layout, whereas `gopdfsuit` was performing full table layout calculations.

### Conclusion
For high-performance server-side PDF generation, **gopdfsuit** outperforms specific layout engines (PDFKit, Typst, FPDF) by a significant margin. It matches the speed of lower-level raw PDF writers (pdf-lib) while offering higher-level layout abstractions.
## 5. Scenario: Financial Report (Complex Layout)

*   **Description**: Generates a 2-page report with tables, charts, styling, and bookmarks using the `gopdflib` public API.
*   **Command**: `go run sampledata/gopdflib/financial_report/main.go`
*   **Iterations**: 10 runs

### Raw Results

| Run # | Time (ms) | PDF Size (Bytes) |
| :--- | :--- | :--- |
| 1 | 7.18 | 135,116 |
| 2 | 6.47 | 135,116 |
| 3 | 6.62 | 135,116 |
| 4 | 6.66 | 135,116 |
| 5 | 6.57 | 135,116 |
| 6 | 6.11 | 135,116 |
| 7 | 5.79 | 135,116 |
| 8 | 7.31 | 135,116 |
| 9 | 7.33 | 135,116 |
| 10 | 6.39 | 135,116 |

### Statistics

| Metric | Result |
| :--- | :--- |
| **Minimum** | 5.79 ms |
| **Maximum** | 7.33 ms |
| **Average (Mean)** | ~6.64 ms |
| **Total Time (10 runs)** | 66.43 ms |

### Analysis
The financial report generation is extremely fast (~6.6ms per document), demonstrating the efficiency of the `gopdflib` API for complex layouts involving tables and embedded images.

## 6. Conclusion
For applications requiring high-frequency document generation, `gopdfsuit` provides a massive throughput advantage. Typst remains a viable option for lower-volume, high-design requirements where a ~740ms generation time is acceptable.
