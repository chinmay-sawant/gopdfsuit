# This is a sample Python script to generate a PDF from a JSON template using the GoPDFSuit API.
# It demonstrates how to fill a template with user data and generate a PDF.

import logging
import json
import re
from pathlib import Path
import time
from gopdf import PdfClient

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(levelname)s: %(message)s')

def fill_template(template, data):
    """
    Recursively traverse the template (dict/list/str) and replace {key} placeholders
    with values from the data dictionary.
    """
    if isinstance(template, dict):
        return {k: fill_template(v, data) for k, v in template.items()}
    elif isinstance(template, list):
        return [fill_template(i, data) for i in template]
    elif isinstance(template, str):
        # Replace occurrences of {key} with data[key]
        # We use a regex to find all {identifier} patterns
        def replace_match(match):
            key = match.group(1)
            # Return the value from data if found, otherwise keep the placeholder
            # Convert non-string values to string
            return str(data.get(key, match.group(0)))
        
        return re.sub(r'\{(\w+)\}', replace_match, template)
    else:
        return template

def main():
    # Configuration
    BASE_DIR = Path(__file__).parent
    TEMPLATE_FILE = BASE_DIR / "financial_template.json"
    OUTPUT_FILE = BASE_DIR / "financial_report.pdf"

    if not TEMPLATE_FILE.exists():
        logging.error(f"Template file {TEMPLATE_FILE} not found.")
        return

    # 1. Define User Input (Data to fill)
    user_input = {
        "company_name": "TechCorp Industries Inc.",
        "company_url": "https://techcorp.example.com",
        "report_period": "Q4 2025",
        "address_line_1": "123 Business Ave, ",
        "address_line_2": "Suite 456, City, State 12345",
        "fiscal_year": "2025",
        "total_revenue": "$2,450,000",
        "cogs": "$1,225,000",
        "gross_profit": "$1,225,000",
        "operating_expenses": "$750,000",
        "rnd_expenses": "$200,000",
        "marketing_sales": "$150,000",
        "admin_expenses": "$100,000",
        "depreciation": "$50,000",
        "interest_expense": "$25,000",
        "taxes": "$75,000",
        "net_income": "$125,000",
        "eps": "$2.50",
        "total_assets": "$5,000,000",
        "total_liabilities": "$2,500,000",
        "shareholders_equity": "$2,500,000",
        "footer_text": "TECHCORP INDUSTRIES INC. | FINANCIAL REPORT Q4 2025 | CONFIDENTIAL",
        "footer_link": "https://example.com/legal",
        "figure_1_caption": "Figure 1: Quarterly revenue comparison by region",
        "figure_2_caption": "Figure 2: Breakdown of operating expenses"
    }

    # 2. Load Template
    logging.info(f"Loading template from {TEMPLATE_FILE}...")
    try:
        with open(TEMPLATE_FILE, "r") as f:
            template_data = json.load(f)
    except Exception as e:
        logging.error(f"Error loading template: {e}")
        return

    # 3. Fill Template
    logging.info("Filling template with user data...")
    filled_data = fill_template(template_data, user_input)

    # 3.5 Transform Structure (if needed)
    # The library template uses a "table" list and "elements" with indices (referenced style).
    # The PdfRequest model expects "elements" to contain the table definitions directly (embedded style).
    if "table" in filled_data and isinstance(filled_data["table"], list):
        logging.info("Transforming 'table' list to embedded 'elements'...")
        tables = filled_data.pop("table")
        
        # If 'elements' exists and has indices, map them. Otherwise just map all tables in order.
        new_elements = []
        if "elements" in filled_data:
            for el in filled_data["elements"]:
                if el.get("type") == "table" and "index" in el:
                    idx = el["index"]
                    if 0 <= idx < len(tables):
                        new_elements.append({
                            "type": "table",
                            "table": tables[idx]
                        })
        else:
            # Fallback: just add all tables in order if no elements map exists
            for tbl in tables:
                new_elements.append({
                    "type": "table",
                    "table": tbl
                })
        
        filled_data["elements"] = new_elements

    # 4. Generate PDF
    client = PdfClient()
    logging.info(f"Sending request to generate PDF...")
    pdf_content = client.generate_pdf(filled_data)
    
    if pdf_content:
        logging.info(f"Success! Saving PDF to {OUTPUT_FILE}...")
        with open(OUTPUT_FILE, "wb") as f:
            f.write(pdf_content)
        logging.info("Done.")
    else:
        logging.error("Failed to generate PDF.")

if __name__ == "__main__":
    now = time.time()
    main()
    end = time.time()
    execution_time_ms = (end - now) * 1000
    logging.info(f"Execution time: {execution_time_ms:.2f} milliseconds")
