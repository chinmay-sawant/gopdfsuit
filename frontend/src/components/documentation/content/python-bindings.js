export const pythonBindingsSection = {
    title: 'Python Bindings',
    items: [
        {
            id: 'python-bindings-intro',
            title: 'Native Python Support',
            description: 'Native Python bindings for GoPdfSuit using CGO and ctypes.',
            content: `Use the power of GoPdfSuit directly in your Python applications without running a separate server.
            
The **pypdfsuit** package provides high-performance bindings to the Go core using a **CGO + ctypes** architecture.

**Key Features:**
• **Zero Interaction Overhead**: Calls Go functions directly in-process via shared library.
• **Type-Safe API**: Uses Python dataclasses and Pydantic models.
• **Full Feature Set**: PDF Generation, Merging, Splitting, Form Filling, and HTML Conversion.
`,
            features: [
                { title: 'Native Performance', description: 'Direct CGO calls without network overhead', icon: 'Zap' },
                { title: 'Type Safety', description: 'Python dataclasses matching Go structs', icon: 'CheckCircle' },
                { title: 'No Server Required', description: 'Embed directly in your Python apps', icon: 'Box' },
                { title: 'Cross-Platform', description: 'Supports Linux, macOS, and Windows', icon: 'Globe' }
            ]
        },
        {
            id: 'python-installation',
            title: 'Installation',
            description: 'Install the pypdfsuit package.',
            content: `You can install the package directly from the repository.

Prerequisites:
• **Go 1.21+** (for building the shared library)
• **Python 3.8+**
• **GCC/Clang** (for CGO compilation)`,
            code: {
                bash: `# Clone the repository
git clone https://github.com/chinmay-sawant/gopdfsuit.git
cd gopdfsuit

# Build and install the Python package
cd bindings/python
./build.sh   # Compiles the Go shared library
pip install .`
            }
        },
        {
            id: 'python-generate-pdf',
            title: 'Generate PDF',
            description: 'Generate PDFs from Python using Type-Safe Templates.',
            content: `Create PDFs using the \`PDFTemplate\` class. The structure mirrors the JSON template format exactly.`,
            code: {
                python: `from pypdfsuit import generate_pdf, PDFTemplate, Config, Title, Element, Table, Row, Cell

# Create a template
template = PDFTemplate(
    config=Config(
        page="A4", 
        page_alignment=1, # Portrait
        pdf_title="My Python PDF"
    ),
    title=Title(
        props="Helvetica:24:100:center:0:0:0:0",
        text="Hello via Python!"
    ),
    elements=[
        Element(
            type="table",
            table=Table(
                maxcolumns=2,
                column_widths=[1.0, 1.0],
                rows=[
                    Row(row=[
                        Cell(props="Helvetica:12:100:left:1:1:1:1", text="Language"),
                        Cell(props="Helvetica:12:000:left:1:1:1:1", text="Python + Go"),
                    ]),
                    Row(row=[
                        Cell(props="Helvetica:12:100:left:1:1:1:1", text="Performance"),
                        Cell(props="Helvetica:12:000:left:1:1:1:1", text="Fast 🚀"),
                    ])
                ]
            )
        )
    ]
)

# Generate bytes
pdf_bytes = generate_pdf(template)

# Save to file
with open("python_output.pdf", "wb") as f:
    f.write(pdf_bytes)`
            }
        },
        {
            id: 'python-merge',
            title: 'Merge PDFs',
            description: 'Combine multiple PDF documents programmatically.',
            code: {
                python: `from pypdfsuit import merge_pdfs

# Read input PDFs
with open("doc1.pdf", "rb") as f:
    pdf1 = f.read()

with open("doc2.pdf", "rb") as f:
    pdf2 = f.read()

# Merge them
merged_bytes = merge_pdfs([pdf1, pdf2])

with open("merged.pdf", "wb") as f:
    f.write(merged_bytes)`
            }
        },
        {
            id: 'python-html-to-pdf',
            title: 'HTML to PDF',
            description: 'Convert HTML content or URLs to PDF.',
            code: {
                python: `from pypdfsuit import convert_html_to_pdf, HtmlToPDFRequest

# Convert raw HTML
req = HtmlToPDFRequest(
    html="<html><body><h1>Generated from Python</h1></body></html>",
    page_size="A4"
)
output = convert_html_to_pdf(req)

# Convert URL
url_req = HtmlToPDFRequest(
    url="https://example.com",
    page_size="Letter"
)
url_output = convert_html_to_pdf(req)`
            }
        },
        {
            id: 'python-redact',
            title: 'Redact PDF',
            description: 'Apply visual or structural redactions securely from Python.',
            code: {
                python: `from pypdfsuit import apply_redactions_advanced

# Read input PDF
with open("financial_report.pdf", "rb") as f:
    pdf_bytes = f.read()

# Apply redactions via text search
out = apply_redactions_advanced(
    pdf_bytes,
    {
        "mode": "visual_allowed",
        "textSearch": [{"text": "SECTION"}, {"text": "Total"}],
    }
)

if out:
    with open("redacted.pdf", "wb") as f:
        f.write(out)`
            }
        }
    ]
};

