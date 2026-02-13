#!/bin/bash
# Comprehensive PDF Library Benchmark Script
# Runs all benchmarks and collects results

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

OUTPUT_FILE="benchmark_results_raw.txt"

echo "=========================================="
echo "PDF Library Benchmark Suite"
echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
echo "=========================================="
echo ""

# Clear output
> "$OUTPUT_FILE"
echo "# Benchmark Results - $(date '+%Y-%m-%d %H:%M:%S')" >> "$OUTPUT_FILE"
echo "" >> "$OUTPUT_FILE"

# Results arrays
declare -a GOPDFSUIT_TIMES=()

# Run GoPDFSuit benchmark 10 times
echo ">>> Running GoPDFSuit (Go) - 10 iterations..."
echo "## GoPDFSuit (10 runs)" >> "$OUTPUT_FILE"
cd gopdfsuit
for i in {1..10}; do
    TIME=$(go run bench.go 2>&1 | grep "GoPDFSuit Time" | awk '{print $3}')
    GOPDFSUIT_TIMES+=($TIME)
    echo "  Run $i: $TIME ms"
    echo "Run $i: $TIME ms" >> "../$OUTPUT_FILE"
done

# Calculate average using awk (bc may not be available)
GOPDFSUIT_AVG=$(printf '%s\n' "${GOPDFSUIT_TIMES[@]}" | awk '{sum+=$1} END {printf "%.2f", sum/NR}')
GOPDFSUIT_MIN=$(printf '%s\n' "${GOPDFSUIT_TIMES[@]}" | sort -n | head -1)
GOPDFSUIT_MAX=$(printf '%s\n' "${GOPDFSUIT_TIMES[@]}" | sort -n | tail -1)
GOPDFSUIT_SIZE=$(ls -lh output.pdf | awk '{print $5}')

echo "  Average: $GOPDFSUIT_AVG ms (Min: $GOPDFSUIT_MIN, Max: $GOPDFSUIT_MAX)"
echo "Average: $GOPDFSUIT_AVG ms, Min: $GOPDFSUIT_MIN ms, Max: $GOPDFSUIT_MAX ms" >> "../$OUTPUT_FILE"
echo "File Size: $GOPDFSUIT_SIZE" >> "../$OUTPUT_FILE"
echo "" >> "../$OUTPUT_FILE"
echo ""
cd ..

# Run PDFKit benchmark
echo ">>> Running PDFKit (Node.js)..."
echo "## PDFKit" >> "$OUTPUT_FILE"
cd pdfkit
PDFKIT_TIME=$(node bench.js 2>&1 | grep "PDFKit Time" | awk '{print $3}')
PDFKIT_SIZE=$(ls -lh output_pdfkit.pdf | awk '{print $5}')
echo "  Time: $PDFKIT_TIME ms, Size: $PDFKIT_SIZE"
echo "Time: $PDFKIT_TIME ms, Size: $PDFKIT_SIZE" >> "../$OUTPUT_FILE"
echo "" >> "../$OUTPUT_FILE"
echo ""
cd ..

# Run pdf-lib benchmark
echo ">>> Running pdf-lib (Node.js)..."
echo "## pdf-lib" >> "$OUTPUT_FILE"
cd pdflib
PDFLIB_TIME=$(node bench.js 2>&1 | grep "pdf-lib Time" | awk '{print $3}')
PDFLIB_SIZE=$(ls -lh output_pdflib.pdf | awk '{print $5}')
echo "  Time: $PDFLIB_TIME ms, Size: $PDFLIB_SIZE"
echo "Time: $PDFLIB_TIME ms, Size: $PDFLIB_SIZE" >> "../$OUTPUT_FILE"
echo "" >> "../$OUTPUT_FILE"
echo ""
cd ..

# Run jsPDF benchmark
echo ">>> Running jsPDF (Node.js)..."
echo "## jsPDF" >> "$OUTPUT_FILE"
cd jspdf
JSPDF_OUTPUT=$(node bench.js 2>&1)
JSPDF_TIME=$(echo "$JSPDF_OUTPUT" | grep "jsPDF Time" | awk '{print $3}')
JSPDF_SIZE=$(ls -lh output_jspdf.pdf | awk '{print $5}')
JSPDF_NOTE=""
if echo "$JSPDF_OUTPUT" | grep -q "fallback"; then
    JSPDF_NOTE=" (text fallback)"
fi
echo "  Time: $JSPDF_TIME ms$JSPDF_NOTE, Size: $JSPDF_SIZE"
echo "Time: $JSPDF_TIME ms$JSPDF_NOTE, Size: $JSPDF_SIZE" >> "../$OUTPUT_FILE"
echo "" >> "../$OUTPUT_FILE"
echo ""
cd ..

# Run FPDF2 benchmark
echo ">>> Running FPDF2 (Python)..."
echo "## FPDF2" >> "$OUTPUT_FILE"
cd fpdf
FPDF_TIME=$(python bench.py 2>&1 | grep "FPDF Time" | awk '{print $3}')
FPDF_SIZE=$(ls -lh output_fpdf.pdf | awk '{print $5}')
echo "  Time: $FPDF_TIME ms, Size: $FPDF_SIZE"
echo "Time: $FPDF_TIME ms, Size: $FPDF_SIZE" >> "../$OUTPUT_FILE"
echo "" >> "../$OUTPUT_FILE"
echo ""
cd ..

# Run Typst benchmark (if available)
echo ">>> Running Typst..."
echo "## Typst" >> "$OUTPUT_FILE"
cd typst
TYPST_BIN="../typst-x86_64-unknown-linux-musl/typst"
if [ -x "$TYPST_BIN" ]; then
    START=$(date +%s%N)
    $TYPST_BIN compile --pdf-standard a-2b benchmark.typ output_typst.pdf 2>/dev/null
    END=$(date +%s%N)
    TYPST_TIME=$(awk "BEGIN {printf \"%.2f\", ($END - $START) / 1000000}")
    TYPST_SIZE=$(ls -lh output_typst.pdf | awk '{print $5}')
    echo "  Time: $TYPST_TIME ms, Size: $TYPST_SIZE"
    echo "Time: $TYPST_TIME ms, Size: $TYPST_SIZE" >> "../$OUTPUT_FILE"
else
    TYPST_TIME="N/A"
    TYPST_SIZE="N/A"
    echo "  Typst not installed"
    echo "Not installed" >> "../$OUTPUT_FILE"
fi
echo "" >> "../$OUTPUT_FILE"
echo ""
cd ..

# Print summary
echo "=========================================="
echo "BENCHMARK RESULTS SUMMARY"
echo "=========================================="
echo ""
printf "| %-12s | %-12s | %-10s | %-10s | %-20s |\n" "Library" "Language" "Time (ms)" "File Size" "PDF Standard"
printf "|--------------|--------------|------------|------------|----------------------|\n"
printf "| %-12s | %-12s | %-10s | %-10s | %-20s |\n" "GoPDFSuit" "Go" "$GOPDFSUIT_AVG" "$GOPDFSUIT_SIZE" "PDF/A-4 (PDF 2.0)"
printf "| %-12s | %-12s | %-10s | %-10s | %-20s |\n" "jsPDF" "JavaScript" "$JSPDF_TIME" "$JSPDF_SIZE" "PDF 1.3$JSPDF_NOTE"
printf "| %-12s | %-12s | %-10s | %-10s | %-20s |\n" "PDFKit" "JavaScript" "$PDFKIT_TIME" "$PDFKIT_SIZE" "PDF 1.3"
printf "| %-12s | %-12s | %-10s | %-10s | %-20s |\n" "Typst" "Typst" "$TYPST_TIME" "$TYPST_SIZE" "PDF/A-2b"
printf "| %-12s | %-12s | %-10s | %-10s | %-20s |\n" "pdf-lib" "JavaScript" "$PDFLIB_TIME" "$PDFLIB_SIZE" "PDF 1.7"
printf "| %-12s | %-12s | %-10s | %-10s | %-20s |\n" "FPDF2" "Python" "$FPDF_TIME" "$FPDF_SIZE" "PDF/A-1b"
echo ""
echo "GoPDFSuit 10-run stats: Min=$GOPDFSUIT_MIN ms, Max=$GOPDFSUIT_MAX ms, Avg=$GOPDFSUIT_AVG ms"
echo ""

# Export summary
echo "" >> "$OUTPUT_FILE"
echo "## Summary" >> "$OUTPUT_FILE"
echo "GoPDFSuit: $GOPDFSUIT_AVG ms (avg of 10), $GOPDFSUIT_SIZE, PDF/A-4" >> "$OUTPUT_FILE"
echo "jsPDF: $JSPDF_TIME ms, $JSPDF_SIZE, PDF 1.3$JSPDF_NOTE" >> "$OUTPUT_FILE"
echo "PDFKit: $PDFKIT_TIME ms, $PDFKIT_SIZE, PDF 1.3" >> "$OUTPUT_FILE"
echo "Typst: $TYPST_TIME ms, $TYPST_SIZE, PDF/A-2b" >> "$OUTPUT_FILE"
echo "pdf-lib: $PDFLIB_TIME ms, $PDFLIB_SIZE, PDF 1.7" >> "$OUTPUT_FILE"
echo "FPDF2: $FPDF_TIME ms, $FPDF_SIZE, PDF/A-1b" >> "$OUTPUT_FILE"

echo "Results saved to $OUTPUT_FILE"
