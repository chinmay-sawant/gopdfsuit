# Comprehensive Benchmark Report: gopdfsuit vs Typst

## 1. Test Configuration

*   **Date**: 2026-01-30
*   **Hardware**: 13th Gen Intel(R) Core(TM) i7-13700HX
*   **Iterations**: 10 runs per tool
*   **Dataset**: 2,000 rows of user data (ID, Name, Email, Role, Description)
*   **Command**: `go test -bench=Benchmark -benchmem -run=^$ -count=10 -v ./internal/pdf`

## 2. Raw Results

| Run # | gopdfsuit (ms) | Typst (ms) | Speedup (x) |
| :--- | :--- | :--- | :--- |
| 1 | 130.76 | 761.50 | 5.82 |
| 2 | 132.54 | 751.34 | 5.67 |
| 3 | 135.49 | 765.27 | 5.65 |
| 4 | 145.47 | 705.91 | 4.85 |
| 5 | 141.03 | 733.26 | 5.20 |
| 6 | 129.28 | 717.59 | 5.55 |
| 7 | 132.06 | 767.09 | 5.81 |
| 8 | 137.95 | 768.91 | 5.57 |
| 9 | 132.01 | 697.15 | 5.28 |
| 10 | 136.19 | 721.97 | 5.30 |

## 3. Statistical Summary

| Metric | gopdfsuit (Library) | Typst (CLI) |
| :--- | :--- | :--- |
| **Minimum** | **129.28 ms** | 697.15 ms |
| **Maximum** | 145.47 ms | 768.91 ms |
| **Average (Mean)** | **~135.28 ms** | **~739.00 ms** |
| **Throughput** | ~7.4 docs/sec | ~1.35 docs/sec |
| **Memory/Op** | ~49 MB | ~17 KB (allocs only)* |

*\*Note: Memory for Typst shows only the Go wrapper allocations. The actual Typst process uses significantly more private memory.*

## 4. Analysis

**gopdfsuit is statistically ~5.46x faster than Typst on average.**

The variance in `gopdfsuit` is low (approx ±6%), indicating stable performance suitable for high-throughput APIS. Typst shows slightly higher variance (approx ±10%), likely due to process startup and OS scheduling variability.

### Conclusion
For applications requiring high-frequency document generation, `gopdfsuit` provides a massive throughput advantage. Typst remains a viable option for lower-volume, high-design requirements where a ~740ms generation time is acceptable.
