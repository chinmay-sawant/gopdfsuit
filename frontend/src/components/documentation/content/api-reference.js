export const apiReferenceSection = {
  title: 'API Reference',
  items: [
    {
      id: 'generate-pdf',
      title: 'Generate PDF',
      method: 'POST',
      endpoint: '/api/v1/generate/template-pdf',
      description: 'Generate a PDF document from a JSON template. Returns the PDF as binary data.',
      params: [
        { name: 'config', type: 'object', required: true, description: 'Page configuration (size, orientation, security, signature)' },
        { name: 'title', type: 'object', required: true, description: 'Document title section' },
        { name: 'elements', type: 'array', required: false, description: 'Ordered array of tables, spacers, images' },
        { name: 'table', type: 'array', required: false, description: 'Legacy: array of tables' },
        { name: 'footer', type: 'object', required: true, description: 'Page footer configuration' },
        { name: 'bookmarks', type: 'array', required: false, description: 'Document outline/navigation bookmarks' }
      ],
      code: {
        curl: `curl -X POST "http://localhost:8080/api/v1/generate/template-pdf" \\
  -H "Content-Type: application/json" \\
  -d '{
    "config": {
      "page": "A4",
      "pageAlignment": 1,
      "pdfTitle": "Report"
    },
    "title": {
      "props": "Helvetica:24:100:center:0:0:0:0",
      "text": "FINANCIAL REPORT",
      "bgcolor": "#154360",
      "textcolor": "#FFFFFF"
    },
    "elements": [
      {
        "type": "table",
        "table": {
          "maxcolumns": 2,
          "columnwidths": [1, 2],
          "rows": [{"row": [
            {"props": "Helvetica:10:100:left:1:1:1:1", "text": "Company:"},
            {"props": "Helvetica:10:000:left:1:1:1:1", "text": "TechCorp Inc."}
          ]}]
        }
      }
    ],
    "footer": {"font": "Helvetica:8:000:center", "text": "Confidential"}
  }' --output document.pdf`,
        node: `const response = await fetch('http://localhost:8080/api/v1/generate/template-pdf', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    config: { page: 'A4', pageAlignment: 1, pdfTitle: 'Report' },
    title: {
      props: 'Helvetica:24:100:center:0:0:0:0',
      text: 'FINANCIAL REPORT',
      bgcolor: '#154360',
      textcolor: '#FFFFFF'
    },
    elements: [{
      type: 'table',
      table: {
        maxcolumns: 2,
        columnwidths: [1, 2],
        rows: [{row: [
          {props: 'Helvetica:10:100:left:1:1:1:1', text: 'Company:'},
          {props: 'Helvetica:10:000:left:1:1:1:1', text: 'TechCorp Inc.'}
        ]}]
      }
    }],
    footer: {font: 'Helvetica:8:000:center', text: 'Confidential'}
  })
});

const pdfBlob = await response.blob();`,
        python: `from gopdf import PdfClient

client = PdfClient(api_url="http://localhost:8080/api/v1/generate/template-pdf")

# Define request data
payload = {
    "config": {
        "page": "A4",
        "pageAlignment": 1,
        "pdfTitle": "Financial Report"
    },
    "title": {
        "props": "Helvetica:24:100:center:0:0:0:0",
        "text": "FINANCIAL REPORT",
        "bgcolor": "#154360",
        "textcolor": "#FFFFFF"
    },
    "elements": [
        {
            "type": "table",
            "table": {
                "maxcolumns": 2,
                "columnwidths": [1, 2],
                "rows": [{
                    "row": [
                        {"props": "Helvetica:10:100:left:1:1:1:1", "text": "Company:"},
                        {"props": "Helvetica:10:000:left:1:1:1:1", "text": "TechCorp Inc."}
                    ]
                }]
            }
        }
    ],
    "footer": {
        "font": "Helvetica:8:000:center", 
        "text": "Confidential"
    }
}

# Generate PDF
pdf_bytes = client.generate_pdf(payload)

if pdf_bytes:
    with open("report.pdf", "wb") as f:
        f.write(pdf_bytes)
    print("PDF saved as report.pdf")`
      }
    },
    {
      id: 'merge-pdfs',
      title: 'Merge PDFs',
      method: 'POST',
      endpoint: '/api/v1/merge',
      description: 'Combine multiple PDF files into a single document. Send PDFs as multipart form data.',
      params: [
        { name: 'pdf', type: 'file', required: true, description: 'PDF files to merge (multiple). Files are merged in order received.' }
      ],
      code: {
        curl: `curl -X POST "http://localhost:8080/api/v1/merge" \\
  -F "pdf=@file1.pdf" \\
  -F "pdf=@file2.pdf" \\
  -F "pdf=@file3.pdf" \\
  --output merged.pdf`,
        node: `const formData = new FormData();
formData.append('pdf', file1);
formData.append('pdf', file2);
formData.append('pdf', file3);

const response = await fetch('http://localhost:8080/api/v1/merge', {
  method: 'POST',
  body: formData
});

const mergedPdf = await response.blob();`
      }
    },
    {
      id: 'fill-form',
      title: 'Fill PDF Form',
      method: 'POST',
      endpoint: '/api/v1/fill',
      description: 'Fill AcroForm fields in a PDF using XFDF data. Returns the filled PDF.',
      params: [
        { name: 'pdf', type: 'file', required: true, description: 'The PDF file with form fields to fill' },
        { name: 'xfdf', type: 'file', required: true, description: 'XFDF file containing field values' }
      ],
      code: {
        curl: `curl -X POST "http://localhost:8080/api/v1/fill" \\
  -F "pdf=@form.pdf" \\
  -F "xfdf=@data.xfdf" \\
  --output filled.pdf`,
        node: `const formData = new FormData();
formData.append('pdf', pdfFile);
formData.append('xfdf', xfdfFile);

const response = await fetch('http://localhost:8080/api/v1/fill', {
  method: 'POST',
  body: formData
});

const filledPdf = await response.blob();`
      }
    },
    {
      id: 'html-to-pdf',
      title: 'HTML to PDF',
      method: 'POST',
      endpoint: '/api/v1/htmltopdf',
      description: 'Convert HTML content or a URL to PDF using headless Chrome.',
      params: [
        { name: 'html', type: 'string', required: false, description: 'Raw HTML content to convert' },
        { name: 'url', type: 'string', required: false, description: 'URL of the page to convert (if html not provided)' },
        { name: 'page_size', type: 'string', default: 'A4', description: 'Page size: A4, Letter, Legal, A3, A5' },
        { name: 'orientation', type: 'string', default: 'Portrait', description: 'Portrait or Landscape' }
      ],
      code: {
        curl: `curl -X POST "http://localhost:8080/api/v1/htmltopdf" \\
  -H "Content-Type: application/json" \\
  -d '{
    "html": "<h1>Hello World</h1><p>This is a PDF.</p>",
    "page_size": "A4",
    "orientation": "Portrait"
  }' --output output.pdf`,
        node: `const response = await fetch('http://localhost:8080/api/v1/htmltopdf', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    html: '<h1>Hello World</h1><p>This is a PDF.</p>',
    page_size: 'A4',
    orientation: 'Portrait'
  })
});

const pdfBlob = await response.blob();`
      }
    },
    {
      id: 'html-to-image',
      title: 'HTML to Image',
      method: 'POST',
      endpoint: '/api/v1/htmltoimage',
      description: 'Convert HTML content or a URL to PNG, JPG, or SVG using headless Chrome.',
      params: [
        { name: 'html', type: 'string', required: false, description: 'Raw HTML content to convert' },
        { name: 'url', type: 'string', required: false, description: 'URL of the page to convert' },
        { name: 'format', type: 'string', default: 'png', description: 'Output format: png, jpg, svg' },
        { name: 'width', type: 'int', default: '1920', description: 'Viewport width in pixels' },
        { name: 'height', type: 'int', default: '1080', description: 'Viewport height in pixels' }
      ],
      code: {
        curl: `curl -X POST "http://localhost:8080/api/v1/htmltoimage" \\
  -H "Content-Type: application/json" \\
  -d '{
    "url": "https://example.com",
    "format": "png",
    "width": 1920,
    "height": 1080
  }' --output screenshot.png`,
        node: `const response = await fetch('http://localhost:8080/api/v1/htmltoimage', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    url: 'https://example.com',
    format: 'png',
    width: 1920,
    height: 1080
  })
});

const imageBlob = await response.blob();`
      }
    },
    {
      id: 'get-template',
      title: 'Get Template Data',
      method: 'GET',
      endpoint: '/api/v1/template-data',
      description: 'Retrieve a saved template JSON file from the server.',
      params: [
        { name: 'file', type: 'string', required: true, description: 'Filename of the template (e.g., temp_multiplepage.json)' }
      ],
      code: {
        curl: `curl "http://localhost:8080/api/v1/template-data?file=temp_multiplepage.json"`,
        node: `const response = await fetch('http://localhost:8080/api/v1/template-data?file=temp_multiplepage.json');
const template = await response.json();`
      }
    },
    {
      id: 'redact-pdf',
      title: 'Redact PDF',
      method: 'POST',
      endpoint: '/api/v1/redact/apply',
      description: 'Apply redactions to a PDF document using explicit regions or text search. Send as multipart form data.',
      params: [
        { name: 'pdf', type: 'file', required: true, description: 'The PDF file to redact' },
        { name: 'blocks', type: 'string', required: false, description: 'JSON array of RedactionRect objects' },
        { name: 'textSearch', type: 'string', required: false, description: 'JSON array of TextQuery objects (e.g., [{"text": "Secret"}])' },
        { name: 'mode', type: 'string', required: false, description: 'Redaction mode: "secure_required" or "visual_allowed"' }
      ],
      code: {
        curl: `curl -X POST "http://localhost:8080/api/v1/redact/apply" \\
  -F "pdf=@Epsteinfiles.pdf" \\
  -F "blocks=[]" \\
  -F "textSearch=[{\\"text\\":\\"donald\\"},{\\"text\\":\\"Jeffrey epstein\\"}]" \\
  -F "mode=secure_required" \\
  --output redacted.pdf`,
        node: `const formData = new FormData();
formData.append('pdf', pdfFile);
formData.append('blocks', '[]');
formData.append('textSearch', '[{"text":"donald"}, {"text":"Jeffrey epstein"}]');
formData.append('mode', 'secure_required');

const response = await fetch('http://localhost:8080/api/v1/redact/apply', {
  method: 'POST',
  body: formData
});

const redactedPdf = await response.blob();
// Check response headers for X-Redaction-Report
const reportMsg = response.headers.get('X-Redaction-Report');`
      }
    }
  ]
};

