# Benchmark Scripts

This directory contains benchmark scripts for:
- jsPDF (Node.js)
- pdf-lib (Node.js)
- PDFKit (Node.js)
- FPDF (Python)

## Setup

1. Install Node dependencies:
   ```bash
   npm install
   ```

2. Install Python dependencies:
   ```bash
   pip install fpdf2
   ```

3. Generate data (if not exists):
   ```bash
   go run gen_data.go
   ```

## Running Benchmarks

Run individually:
```bash
node bench_jspdf.js
node bench_pdflib.js
node bench_pdfkit.js
python3 bench_fpdf.py
```
