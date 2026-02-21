
import os
import sys
import json
from pathlib import Path

# Add bindings to path
bindings_path = Path(__file__).parent.parent.parent / "bindings" / "python"
sys.path.append(str(bindings_path))

from pypdfsuit.redact import get_page_info, extract_text_positions, apply_redactions

def test_redaction():
    project_root = Path(__file__).parent.parent.parent
    pdf_path = project_root / "sampledata/financialreport/financial_report.pdf"
    output_path = project_root / "redacted_test.pdf"

    print(f"Testing with PDF: {pdf_path}")
    
    if not pdf_path.exists():
        print(f"Error: {pdf_path} does not exist")
        sys.exit(1)

    with open(pdf_path, "rb") as f:
        pdf_bytes = f.read()

    # 1. Test GetPageInfo
    print("\n--- Testing GetPageInfo ---")
    try:
        info = get_page_info(pdf_bytes)
        print(f"Total Pages: {info['totalPages']}")
        print(f"Page 1 Dimensions: {info['pages'][0]}")
    except Exception as e:
        print(f"GetPageInfo failed: {e}")
        sys.exit(1)

    # 2. Test ExtractTextPositions
    print("\n--- Testing ExtractTextPositions (Page 1) ---")
    try:
        positions = extract_text_positions(pdf_bytes, 1)
        print(f"Found {len(positions)} text fragments")
        for i, pos in enumerate(positions[:5]):
            print(f"  [{i}] {pos['text']} at ({pos['x']:.2f}, {pos['y']:.2f})")
    except Exception as e:
        print(f"ExtractTextPositions failed: {e}")
        # Don't exit, try redaction anyway

    # 3. Test ApplyRedactions
    print("\n--- Testing ApplyRedactions ---")
    redactions = [
        {
            "pageNum": 1,
            "x": 100,
            "y": 500,
            "width": 200,
            "height": 50
        }
    ]
    try:
        redacted_bytes = apply_redactions(pdf_bytes, redactions)
        with open(output_path, "wb") as f:
            f.write(redacted_bytes)
        print(f"Redacted PDF saved to: {output_path}")
        print(f"Original size: {len(pdf_bytes)}, Redacted size: {len(redacted_bytes)}")
    except Exception as e:
        print(f"ApplyRedactions failed: {e}")
        sys.exit(1)

if __name__ == "__main__":
    test_redaction()
