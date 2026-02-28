export const sampleDataSection = {
    title: 'Sample Data',
    items: [
        {
            id: 'sampledata-overview',
            title: 'Sample Templates',
            description: 'Ready-to-use JSON templates and examples available in the repository.',
            content: `Complete sample templates and data files are available at [sampledata](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata).

**Available Sample Data Folders**:

• [acroform](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/acroform) - AcroForm templates with interactive form fields
• [benchmarks](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/benchmarks) - Performance benchmark data and reports
• [editor](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/editor) - Templates designed for the visual editor
• [filler](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/filler) - XFDF data files for form filling
• [financialreport](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/financialreport) - Financial report templates
• [gopdflib](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/gopdflib) - Go library usage examples
• [htmltoimg](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/htmltoimg) - HTML to image conversion samples
• [htmltopdf](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/htmltopdf) - HTML to PDF conversion samples
• [legalcontract](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/legalcontract) - Legal document templates
• [librarybook](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/librarybook) - Library management forms
• [merge](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/merge) - PDF merging examples
• [oldata](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/oldata) - Old Data
• [python](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/python) - Python client examples and templates
• [samplecode](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/samplecode) - Code snippets and usage examples
• [split](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/split) - PDF splitting samples
• [svg](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/svg) - SVG vector graphics examples
• [typstsyntax](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/typstsyntax) - Typst math syntax samples and rendered PDFs`
        },
        {
            id: 'sampledata-gopdflib',
            title: 'gopdflib Examples',
            description: 'Go library examples for programmatic PDF generation.',
            content: `**gopdflib Usage**:
The gopdflib examples demonstrate how to use the Go library directly:

• [financial_report](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/gopdflib/financial_report) - Complete financial report with digital signatures, benchmarking, and performance metrics.
• [load_from_json](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/gopdflib/load_from_json) - Loading templates from JSON files and generating PDFs.
• [text_wrapping](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/gopdflib/text_wrapping) - Demonstrating auto text wrapping with dynamic row heights.

Run examples:
go run sampledata/gopdflib/financial_report/main.go
go run sampledata/gopdflib/text_wrapping/main.go`
        },
        {
            id: 'sampledata-python',
            title: 'Python Examples',
            description: 'Python client examples for API-based PDF generation.',
            content: `**Python Client Usage**:
The Python examples demonstrate using the GoPdfSuit API from Python:

sampledata/python/
• main.py - Main example with template filling
• gopdf/ - Python client package
  - client.py - PdfClient class for API calls
  - models.py - Pydantic models for requests
• financial_template.json - Sample financial template
• JsonFileExample.py - Loading JSON templates
• requirements.txt - Python dependencies

Setup and run:
pip install -r requirements.txt
python main.py`,
            code: {
                bash: `# Setup Python environment
cd sampledata/python
python -m venv venv
source venv/bin/activate  # On Windows: venv\\Scripts\\activate
pip install -r requirements.txt

# Start the GoPdfSuit server first
# Then run the Python example
python main.py`,
                python: `# Quick usage example
from gopdf import PdfClient

client = PdfClient(api_url="http://localhost:8080/api/v1/generate/template-pdf")

# Generate from dict
pdf_bytes = client.generate_pdf({
    "config": {"page": "A4", "pageAlignment": 1},
    "title": {"props": "Helvetica:18:100:center:0:0:0:0", "text": "Hello World"},
    "footer": {"font": "Helvetica:8:000:center", "text": "Page"}
})

# Or generate from JSON file
pdf_bytes = client.generate_from_file("template.json")

# Save the PDF
with open("output.pdf", "wb") as f:
    f.write(pdf_bytes)`
            }
        },
        {
            id: 'sampledata-editor',
            title: 'Editor Templates',
            description: 'Pre-built templates for the visual drag-and-drop editor.',
            content: `**Editor Templates Overview**:
The editor templates are designed to work with the /editor visual interface:

sampledata/editor/
• financial_digitalsignature.json - Financial report with digital signature
• financial_encrypted.json - Password-protected financial report
• service_agreement.json - Professional services contract
• library_book_receiving.json - Library book intake form
• us_hospital_encounter.json - Medical encounter form

These templates demonstrate:
• Complex multi-column layouts
• Digital signature configuration
• PDF encryption settings
• Form fields (text, checkbox, radio)
• Styled headers and sections
• Internal navigation with bookmarks`,
            code: {
                bash: `# Load a template in the editor
# 1. Start the server
go run ./cmd/gopdfsuit

# 2. Open http://localhost:8080/editor
# 3. Click "Load Template" and select a JSON file

# Or use the API to get template data
curl "http://localhost:8080/api/v1/template-data?file=financial_digitalsignature.json"`
            }
        }
    ]
};
