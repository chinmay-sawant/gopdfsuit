#!/bin/bash

# Configuration
OUTPUT_FILE="benchmark_results_raw.txt" 
DATA_FILE="data.json"

# Clear output
> "$OUTPUT_FILE"

echo "Starting Benchmarks..."

# 1. gopdfsuit & Typst (via Go Benchmark)
echo "--- gopdfsuit & Typst (Go Test) ---" >> "$OUTPUT_FILE"
cd ../..
# This runs both BenchmarkGoPdfSuit and BenchmarkTypst
# Note: Ensure Typst binary is available for BenchmarkTypst to working
go test -bench=. -benchmem -run=^$ -count=10 -v ./internal/pdf >> "sampledata/benchmarks/$OUTPUT_FILE" 2>&1
cd sampledata/benchmarks

# 2. jsPDF
echo "--- jsPDF ---" >> "$OUTPUT_FILE"
cd jspdf
for i in {1..10}
do
   node bench.js >> "../$OUTPUT_FILE"
done
cd ..

# 3. pdf-lib
echo "--- pdf-lib ---" >> "$OUTPUT_FILE"
cd pdflib
for i in {1..10}
do
   node bench.js >> "../$OUTPUT_FILE"
done
cd ..

# 4. PDFKit
echo "--- PDFKit ---" >> "$OUTPUT_FILE"
cd pdfkit
for i in {1..10}
do
   node bench.js >> "../$OUTPUT_FILE"
done
cd ..

# 5. FPDF
echo "--- FPDF ---" >> "$OUTPUT_FILE"
cd fpdf
for i in {1..10}
do
   python3 bench.py >> "../$OUTPUT_FILE"
done
cd ..

echo "Benchmarks Complete. Results saved to $OUTPUT_FILE"
