export const advancedFeaturesSection = {
    title: 'Advanced Features',
    items: [
        {
            id: 'digital-signatures',
            title: 'Digital Signatures',
            description: 'Add legally-binding digital signatures with X.509 certificates.',
            params: [
                { name: 'enabled', type: 'bool', required: true, description: 'Enable digital signing' },
                { name: 'visible', type: 'bool', required: false, description: 'Show visible signature stamp' },
                { name: 'name', type: 'string', required: false, description: 'Signer name (overrides certificate CN)' },
                { name: 'reason', type: 'string', required: false, description: 'Reason for signing' },
                { name: 'location', type: 'string', required: false, description: 'Geographic location' },
                { name: 'contactInfo', type: 'string', required: false, description: 'Contact information' },
                { name: 'privateKeyPem', type: 'string', required: true, description: 'PEM-encoded private key (RSA or ECDSA)' },
                { name: 'certificatePem', type: 'string', required: true, description: 'PEM-encoded X.509 certificate' },
                { name: 'certificateChain', type: 'array', required: false, description: 'Array of intermediate certificates' }
            ],
            code: {
                json: `{
  "config": {
    "signature": {
      "enabled": true,
      "visible": true,
      "name": "John Doe",
      "reason": "Document Approval",
      "location": "New York, US",
      "contactInfo": "john@example.com",
      "privateKeyPem": "-----BEGIN PRIVATE KEY-----\\n...\\n-----END PRIVATE KEY-----",
      "certificatePem": "-----BEGIN CERTIFICATE-----\\n...\\n-----END CERTIFICATE-----",
      "certificateChain": [
        "-----BEGIN CERTIFICATE-----\\n...intermediate...\\n-----END CERTIFICATE-----",
        "-----BEGIN CERTIFICATE-----\\n...root...\\n-----END CERTIFICATE-----"
      ]
    }
  }
}`
            }
        },
        {
            id: 'pdf-encryption',
            title: 'PDF Encryption',
            description: 'Password-protect documents with granular permission controls.',
            params: [
                { name: 'enabled', type: 'bool', required: true, description: 'Enable encryption' },
                { name: 'ownerPassword', type: 'string', required: true, description: 'Password for full document access' },
                { name: 'userPassword', type: 'string', required: false, description: 'Password to open document (empty = no password to open)' },
                { name: 'allowPrinting', type: 'bool', required: false, description: 'Allow document printing' },
                { name: 'allowCopying', type: 'bool', required: false, description: 'Allow copying text/images' },
                { name: 'allowModifying', type: 'bool', required: false, description: 'Allow content modification' },
                { name: 'allowAnnotations', type: 'bool', required: false, description: 'Allow adding annotations' },
                { name: 'allowFormFilling', type: 'bool', required: false, description: 'Allow filling form fields' }
            ],
            code: {
                json: `{
  "config": {
    "security": {
      "enabled": true,
      "ownerPassword": "admin123",
      "userPassword": "view123",
      "allowPrinting": true,
      "allowCopying": false,
      "allowModifying": false,
      "allowAnnotations": false,
      "allowFormFilling": true
    }
  }
}`
            }
        },
        {
            id: 'bookmarks',
            title: 'Bookmarks & Navigation',
            description: 'Create document outlines with internal navigation links.',
            params: [
                { name: 'title', type: 'string', required: true, description: 'Display text in bookmark panel' },
                { name: 'page', type: 'int', required: false, description: 'Target page number (1-based)' },
                { name: 'dest', type: 'string', required: false, description: 'Named destination (matches cell dest field)' },
                { name: 'y', type: 'float', required: false, description: 'Y position on target page' },
                { name: 'children', type: 'array', required: false, description: 'Nested bookmarks for hierarchy' },
                { name: 'open', type: 'bool', required: false, description: 'Whether children are expanded' }
            ],
            content: `To create internal navigation:

1. Add a destination anchor to a cell using the "dest" field
2. Create a link to that destination using "link": "#destination-name"
3. Optionally add a bookmark for sidebar navigation`,
            code: {
                json: `{
  "bookmarks": [
    {
      "title": "Financial Report",
      "page": 1,
      "children": [
        {"title": "Company Information", "page": 1},
        {"title": "Financial Summary", "dest": "financial-summary"}
      ]
    },
    {
      "title": "Charts & Visualizations",
      "page": 2,
      "dest": "charts-section"
    }
  ]
}`
            }
        },
        {
            id: 'form-fields',
            title: 'Interactive Form Fields',
            description: 'Add fillable form fields to create interactive PDFs.',
            params: [
                { name: 'type', type: 'string', required: true, description: 'checkbox, radio, or text' },
                { name: 'name', type: 'string', required: true, description: 'Field name for data extraction' },
                { name: 'value', type: 'string', required: false, description: 'Export value or default text' },
                { name: 'checked', type: 'bool', required: false, description: 'Initial checked state' },
                { name: 'group_name', type: 'string', required: false, description: 'Radio button group name' },
                { name: 'shape', type: 'string', required: false, description: 'round or square (for radio buttons)' }
            ],
            code: {
                json: `// Text input field
{
  "props": "Helvetica:9:000:left:1:1:1:1",
  "text": "John Doe",
  "form_field": {
    "type": "text",
    "name": "customer_name",
    "value": "John Doe"
  }
}

// Checkbox
{
  "props": "Helvetica:8:000:center:1:1:1:1",
  "form_field": {
    "type": "checkbox",
    "name": "agree_terms",
    "value": "Yes",
    "checked": true
  }
}

// Radio button group
{
  "props": "Helvetica:9:000:center:1:1:1:1",
  "form_field": {
    "type": "radio",
    "name": "payment_method",
    "value": "credit_card",
    "checked": true,
    "group_name": "payment_type",
    "shape": "round"
  }
}`
            }
        },
        {
            id: 'images',
            title: 'Embedded Images',
            description: 'Embed images in cells using base64-encoded data.',
            params: [
                { name: 'imagename', type: 'string', required: false, description: 'Image identifier' },
                { name: 'imagedata', type: 'string', required: true, description: 'Base64-encoded image data' },
                { name: 'width', type: 'float', required: true, description: 'Image width in points' },
                { name: 'height', type: 'float', required: true, description: 'Image height in points' }
            ],
            code: {
                json: `{
  "props": "Helvetica:12:000:center:0:0:0:0",
  "height": 150,
  "link": "https://example.com/charts",
  "image": {
    "imagename": "chart",
    "imagedata": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
    "width": 150,
    "height": 150
  }
}`
            }
        },
        {
            id: 'pdfa-compliance',
            title: 'PDF/A-4 Compliance',
            description: 'Generate archival-standard PDFs for long-term preservation.',
            content: `PDF/A-4 is the archival standard based on PDF 2.0. Enable it with "pdfaCompliant": true.

This automatically:
• Embeds all fonts (via Liberation fonts)
• Adds required XMP metadata
• Follows strict structure requirements
• Ensures long-term document accessibility

Combine with "arlingtonCompatible": true for full PDF 2.0 Arlington Model compliance.`,
            code: {
                json: `{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "pdfTitle": "Archival Document",
    "pdfaCompliant": true,
    "arlingtonCompatible": true,
    "embedFonts": true
  }
}`
            }
        },
        {
            id: 'typst-math',
            title: 'Typst Math Rendering',
            description: 'Render mathematical equations in PDF table cells using Typst syntax.',
            content: `Typst math syntax support allows rendering mathematical expressions in PDF table cells. Enable it with "mathEnabled": true in the cell config.

**How It Works**:
1. Cell values wrapped in \`$...$\` are detected as math expressions
2. The Typst syntax is parsed into an AST
3. The AST is rendered as proper mathematical notation in the PDF

**Supported Syntax**:
• \`$ A = pi r^2 $\` → A = \u03C0r\u00B2
• \`$ frac(a, b) $\` → fraction a/b
• \`$ sqrt(x) $\` → \u221Ax
• \`$ sum_(i=1)^n i $\` → \u03A3 summation
• \`$ vec(a, b, c) $\` → column vector
• \`$ mat(1, 2; 3, 4) $\` → 2\u00D72 matrix

**Sample Data**: [Typst Math Samples](https://github.com/chinmay-sawant/gopdfsuit/tree/master/sampledata/typstsyntax)
**Syntax Reference**: [SYNTAX_REFERENCE.md](https://github.com/chinmay-sawant/gopdfsuit/blob/master/typstsyntax/SYNTAX_REFERENCE.md)`,
            params: [
                { name: 'mathEnabled', type: 'bool', required: true, description: 'Enable Typst math rendering in cells' }
            ],
            code: {
                json: `{
  "config": {
    "mathEnabled": true
  },
  "data": [
    {
      "dataType": "table",
      "rows": [
        {
          "columns": [
            {
              "value": "$ A = pi r^2 $"
            }
          ]
        },
        {
          "columns": [
            {
              "value": "$ frac(a^2 + b^2, c) $"
            }
          ]
        },
        {
          "columns": [
            {
              "value": "$ sum_(i=1)^n i = frac(n(n+1), 2) $"
            }
          ]
        }
      ]
    }
  ]
}`
            }
        }
    ]
};
