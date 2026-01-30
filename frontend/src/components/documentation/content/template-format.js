export const templateFormatSection = {
    title: 'Template Format',
    items: [
        {
            id: 'template-structure',
            title: 'Template Structure',
            description: 'Overview of the JSON template structure used for PDF generation.',
            content: `A PDF template consists of several top-level objects that define the document layout and content.

The recommended approach uses the "elements" array for ordered content. The legacy approach uses separate "table", "spacer", and "image" arrays.`,
            code: {
                json: `{
  "config": { },      // Page configuration (required)
  "title": { },       // Document title (required)
  "elements": [ ],    // Ordered elements - tables, spacers, images (recommended)
  "table": [ ],       // Legacy: array of tables
  "spacer": [ ],      // Legacy: array of spacers
  "image": [ ],       // Legacy: array of images
  "footer": { },      // Page footer (required)
  "bookmarks": [ ]    // Document outline/navigation (optional)
}`
            }
        },
        {
            id: 'config-object',
            title: 'Config Object',
            description: 'Page layout, appearance, and security settings.',
            params: [
                { name: 'page', type: 'string', default: 'A4', description: 'Page size: A4, A3, A5, LETTER, LEGAL' },
                { name: 'pageAlignment', type: 'int', default: '1', description: '1 = Portrait, 2 = Landscape' },
                { name: 'pageBorder', type: 'string', default: '0:0:0:0', description: 'Border widths: left:right:top:bottom' },
                { name: 'watermark', type: 'string', required: false, description: 'Diagonal watermark text across all pages' },
                { name: 'pdfTitle', type: 'string', required: false, description: 'Document title for PDF metadata' },
                { name: 'pdfaCompliant', type: 'bool', default: 'false', description: 'Enable PDF/A-4 compliance' },
                { name: 'arlingtonCompatible', type: 'bool', default: 'false', description: 'Enable PDF 2.0 Arlington Model' },
                { name: 'embedFonts', type: 'bool', default: 'true', description: 'Embed fonts for portability' },
                { name: 'signature', type: 'object', required: false, description: 'Digital signature settings' },
                { name: 'security', type: 'object', required: false, description: 'Encryption/password settings' }
            ],
            code: {
                json: `{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "pageBorder": "0:0:0:0",
    "pdfTitle": "Financial Report Q4 2025",
    "pdfaCompliant": true,
    "arlingtonCompatible": true,
    "embedFonts": true,
    "watermark": "CONFIDENTIAL"
  }
}`
            }
        },
        {
            id: 'page-sizes',
            title: 'Page Sizes',
            description: 'Supported page sizes and their dimensions.',
            content: `The following page sizes are available:

| Size   | Dimensions (inches) | Dimensions (points) |
|--------|---------------------|---------------------|
| A3     | 11.69 × 16.54       | 842 × 1191          |
| A4     | 8.27 × 11.69        | 595 × 842           |
| A5     | 5.83 × 8.27         | 420 × 595           |
| LETTER | 8.5 × 11            | 612 × 792           |
| LEGAL  | 8.5 × 14            | 612 × 1008          |

Use pageAlignment: 1 for Portrait or pageAlignment: 2 for Landscape orientation.`
        },
        {
            id: 'props-syntax',
            title: 'Props Syntax',
            description: 'The props string defines text styling and cell borders.',
            content: `Format: "fontname:fontsize:style:alignment:left:right:top:bottom"

Components:
• fontname - Font identifier (Helvetica, Times-Roman, Courier)
• fontsize - Font size in points
• style - 3-digit code for Bold/Italic/Underline
• alignment - Text alignment: left, center, right
• left/right/top/bottom - Border flags (0=none, 1=draw)

Style Codes:
• 000 = Normal
• 100 = Bold
• 010 = Italic
• 001 = Underline
• 110 = Bold + Italic
• 101 = Bold + Underline
• 011 = Italic + Underline
• 111 = Bold + Italic + Underline`,
            code: {
                json: `// Normal, all borders
"Helvetica:12:000:left:1:1:1:1"

// Bold, centered, top border only
"Helvetica:14:100:center:0:0:1:0"

// Bold+Italic+Underline, right-aligned, no borders
"Helvetica:10:111:right:0:0:0:0"`
            }
        },
        {
            id: 'title-object',
            title: 'Title Object',
            description: 'Defines the document header/title section.',
            params: [
                { name: 'props', type: 'string', required: true, description: 'Styling properties (see Props Syntax)' },
                { name: 'text', type: 'string', required: false, description: 'Title text (ignored if table is provided)' },
                { name: 'table', type: 'object', required: false, description: 'Embedded table for complex layouts' },
                { name: 'bgcolor', type: 'string', required: false, description: 'Background color (hex: #RRGGBB)' },
                { name: 'textcolor', type: 'string', required: false, description: 'Text color (hex: #RRGGBB)' },
                { name: 'link', type: 'string', required: false, description: 'External URL hyperlink' }
            ],
            code: {
                json: `{
  "title": {
    "props": "Helvetica:24:100:center:0:0:0:0",
    "text": "FINANCIAL REPORT",
    "bgcolor": "#154360",
    "textcolor": "#FFFFFF",
    "link": "https://example.com/report"
  }
}`
            }
        },
        {
            id: 'table-object',
            title: 'Table Object',
            description: 'Tables are the primary content containers in PDF templates.',
            params: [
                { name: 'maxcolumns', type: 'int', required: true, description: 'Number of columns' },
                { name: 'rows', type: 'array', required: true, description: 'Array of row objects' },
                { name: 'columnwidths', type: 'array', required: false, description: 'Relative width weights (e.g., [2,1,1] = 50%, 25%, 25%)' },
                { name: 'rowheights', type: 'array', required: false, description: 'Row heights in points (default: 25)' },
                { name: 'bgcolor', type: 'string', required: false, description: 'Default background color for all cells' },
                { name: 'textcolor', type: 'string', required: false, description: 'Default text color for all cells' }
            ],
            code: {
                json: `{
  "type": "table",
  "table": {
    "maxcolumns": 4,
    "columnwidths": [1.2, 2, 1.2, 2],
    "rows": [
      {
        "row": [
          {"props": "Helvetica:10:100:left:1:0:1:1", "text": "Company:", "bgcolor": "#F4F6F7"},
          {"props": "Helvetica:10:000:left:0:0:1:1", "text": "TechCorp Inc."},
          {"props": "Helvetica:10:100:left:0:0:1:1", "text": "Period:"},
          {"props": "Helvetica:10:000:left:0:1:1:1", "text": "Q4 2025"}
        ]
      }
    ]
  }
}`
            }
        },
        {
            id: 'cell-object',
            title: 'Cell Object',
            description: 'Individual cells within table rows support text, checkboxes, images, and form fields.',
            params: [
                { name: 'props', type: 'string', required: true, description: 'Styling properties' },
                { name: 'text', type: 'string', required: false, description: 'Cell text content' },
                { name: 'chequebox', type: 'bool', required: false, description: 'Render checkbox (true=checked)' },
                { name: 'image', type: 'object', required: false, description: 'Embedded image' },
                { name: 'form_field', type: 'object', required: false, description: 'Interactive form field' },
                { name: 'bgcolor', type: 'string', required: false, description: 'Cell background color' },
                { name: 'textcolor', type: 'string', required: false, description: 'Cell text color' },
                { name: 'link', type: 'string', required: false, description: 'URL or internal #destination' },
                { name: 'dest', type: 'string', required: false, description: 'Named destination anchor' },
                { name: 'height', type: 'float', required: false, description: 'Cell height in points' }
            ],
            code: {
                json: `// Text cell with borders
{"props": "Helvetica:12:000:left:1:1:1:1", "text": "Hello World"}

// Checkbox cell
{"props": "Helvetica:12:000:center:1:1:1:1", "chequebox": true}

// Colored cell
{"props": "Helvetica:12:100:center:1:1:1:1", "text": "Alert!", "bgcolor": "#FF0000", "textcolor": "#FFFFFF"}

// External link
{"props": "Helvetica:10:000:left:1:1:1:1", "text": "Visit Website", "link": "https://example.com", "textcolor": "#0000FF"}

// Internal link (jumps to destination)
{"props": "Helvetica:10:000:left:1:1:1:1", "text": "Go to Summary", "link": "#financial-summary", "textcolor": "#0000FF"}

// Destination anchor
{"props": "Helvetica:12:100:left:1:1:1:1", "text": "FINANCIAL SUMMARY", "dest": "financial-summary", "bgcolor": "#21618C", "textcolor": "#FFFFFF"}`
            }
        },
        {
            id: 'footer-object',
            title: 'Footer Object',
            description: 'Footer appears at the bottom of every page. Page numbers are added automatically.',
            params: [
                { name: 'font', type: 'string', required: true, description: 'Font props: "fontname:fontsize:style:alignment"' },
                { name: 'text', type: 'string', required: true, description: 'Footer text' },
                { name: 'link', type: 'string', required: false, description: 'External URL hyperlink' }
            ],
            code: {
                json: `{
  "footer": {
    "font": "Helvetica:8:000:center",
    "text": "TECHCORP INC. | FINANCIAL REPORT Q4 2025 | CONFIDENTIAL",
    "link": "https://example.com/legal"
  }
}`
            }
        }
    ]
};
