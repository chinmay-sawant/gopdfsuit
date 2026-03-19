#!/bin/bash
# Comprehensive PDF Library Benchmark Script
# Runs all benchmarks sequentially and saves results

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

run_bench() {
    local name="$1"
    local dir="$2"
    shift 2
    local cmd=("$@")

    echo ">>> Running $name..."
    cd "$SCRIPT_DIR/$dir"
    echo "==========================================" >> "$SCRIPT_DIR/$OUTPUT_FILE"
    echo "$name" >> "$SCRIPT_DIR/$OUTPUT_FILE"
    echo "==========================================" >> "$SCRIPT_DIR/$OUTPUT_FILE"
    
    # Run the command and append both stdout and stderr to the output file
    "${cmd[@]}" | tee -a "$SCRIPT_DIR/$OUTPUT_FILE"
    
    echo "" >> "$SCRIPT_DIR/$OUTPUT_FILE"
    echo ""
}

run_bench "GoPDFLib Data-Table (Go)" "gopdflib" go run . data
run_bench "GoPDFLib Zerodha (Go)" "gopdflib" go run .
run_bench "GoPDFSuit (Go)" "gopdfsuit" go run .
run_bench "PyPDFSuit (Python)" "pypdfsuit" python3 bench.py
run_bench "FPDF2 (Python)" "fpdf" python3 bench.py
run_bench "jsPDF (Node.js)" "jspdf" node bench.js
run_bench "PDFKit (Node.js)" "pdfkit" node bench.js
run_bench "pdf-lib (Node.js)" "pdflib" node bench.js
run_bench "Typst" "typst" bash bench.sh

echo "=========================================="
echo "All benchmarks completed!"
echo "Check $OUTPUT_FILE for the full raw results."
echo "=========================================="
