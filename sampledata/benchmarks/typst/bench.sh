#!/bin/bash
# Typst benchmark runner with PDF/A-2b compliance
# Requires: typst CLI (https://typst.app/)

set -e

START=$(date +%s%N)

# Compile with PDF/A-2b standard (highest available in Typst)
typst compile --pdf-standard a-2b benchmark.typ output_typst.pdf

END=$(date +%s%N)
ELAPSED=$(echo "scale=2; ($END - $START) / 1000000" | bc)

echo "Typst Time: ${ELAPSED} ms"
echo "PDF Standard: PDF/A-2b (ISO 19005-2:2011)"
