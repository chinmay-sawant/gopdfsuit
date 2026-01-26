"""
In a real-world scenario, you would typically:
1. Create a dynamic JSON template and store it locally with placeholders (e.g., using {} or Jinja2 syntax).
2. Collect user input (such as username, DOB, etc.).
3. Fill the JSON template by replacing these placeholders with the actual user data.
4. Directly call the `generate_pdf` method from the `PdfClient`, sending the populated JSON data 
   to the API and receiving the generated PDF back.
"""
import logging
from pathlib import Path
from gopdf import PdfClient

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(levelname)s: %(message)s')

def main():
    # Configuration
    BASE_DIR = Path(__file__).parent
    JSON_SOURCE = BASE_DIR / "../editor/financial_digitalsignature_python.json"
    OUTPUT_FILE = BASE_DIR / "output.pdf"

    if not JSON_SOURCE.exists():
        logging.error(f"Source file {JSON_SOURCE} not found.")
        return

    client = PdfClient()
    
    logging.info(f"Generating PDF from {JSON_SOURCE}...")
    
    pdf_content = client.generate_from_file(JSON_SOURCE)
    
    if pdf_content:
        logging.info(f"Success! Saving PDF to {OUTPUT_FILE}...")
        with open(OUTPUT_FILE, "wb") as f:
            f.write(pdf_content)
        logging.info("Done.")
    else:
        logging.error("Failed to generate PDF.")

if __name__ == "__main__":
    main()
