# 📋 PDF Template Reference

Complete guide to the JSON template format used by GoPdfSuit for generating PDF documents.

---

## Table of Contents

- [Overview](#overview)
- [Template Structure](#template-structure)
- [Config Object](#config-object)
- [Digital Signatures](#digital-signatures)
- [Security & Encryption](#security--encryption)
- [Title Object](#title-object)
- [Table Object](#table-object)
- [Cell Object](#cell-object)
- [Props Syntax](#props-syntax)
- [Bookmarks & Navigation](#bookmarks--navigation)
- [Form Fields](#form-fields)
- [Images](#images)
- [Footer Object](#footer-object)
- [Complete Example](#complete-example)
- [API Usage](#api-usage)

---

## Overview

GoPdfSuit uses JSON templates to define PDF document structure. Templates are processed by the `/api/v1/generate/template-pdf` endpoint and rendered into PDF documents with automatic page breaks, styling, and form elements.

---

## Template Structure

```json
{
  "config": { },      // Page configuration (required)
  "title": { },       // Document title (required)
  "table": [ ],       // Array of tables (legacy, use elements)
  "spacer": [ ],      // Array of spacers (legacy, use elements)
  "image": [ ],       // Array of images (legacy, use elements)
  "elements": [ ],    // Ordered elements - tables, spacers, images (recommended)
  "footer": { },      // Page footer (required)
  "bookmarks": [ ]    // Document outline/navigation (optional)
}
```

---

## Config Object

Controls page layout, appearance, and security features.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `page` | string | No | `"A4"` | Page size |
| `pageAlignment` | int | No | `1` | Orientation: `1`=Portrait, `2`=Landscape |
| `pageBorder` | string | No | `"0:0:0:0"` | Border widths: `"left:right:top:bottom"` |
| `watermark` | string | No | - | Diagonal watermark text across all pages |
| `pdfTitle` | string | No | - | Document title for PDF metadata |
| `pdfaCompliant` | bool | No | `false` | Enable PDF/A-4 compliance (embeds fonts via Liberation) |
| `arlingtonCompatible` | bool | No | `false` | Enable PDF 2.0 Arlington Model compliance |
| `embedFonts` | bool | No | `true` | Embed fonts for document portability |
| `signature` | object | No | - | Digital signature settings (see [Digital Signatures](#digital-signatures)) |
| `security` | object | No | - | Encryption settings (see [Security & Encryption](#security--encryption)) |

### Page Sizes

| Size | Dimensions (inches) | Dimensions (points) |
|------|---------------------|---------------------|
| `A3` | 11.69 × 16.54 | 842 × 1191 |
| `A4` | 8.27 × 11.69 | 595 × 842 |
| `A5` | 5.83 × 8.27 | 420 × 595 |
| `LETTER` | 8.5 × 11 | 612 × 792 |
| `LEGAL` | 8.5 × 14 | 612 × 1008 |

### Example

```json
{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "pageBorder": "0:0:0:0",
    "pdfTitle": "Financial Report Q4 2025",
    "pdfaCompliant": true,
    "watermark": "CONFIDENTIAL"
  }
}
```

---

## Digital Signatures

Add legally-binding digital signatures with X.509 certificates.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `enabled` | bool | Yes | Enable digital signing |
| `visible` | bool | No | Show visible signature stamp on document |
| `name` | string | No | Signer name (overrides certificate CN) |
| `reason` | string | No | Reason for signing |
| `location` | string | No | Geographic location of signing |
| `contactInfo` | string | No | Contact information |
| `privateKeyPem` | string | Yes | PEM-encoded private key (RSA or ECDSA) |
| `certificatePem` | string | Yes | PEM-encoded X.509 certificate |
| `certificateChain` | array | No | Array of PEM-encoded intermediate certificates |
| `page` | int | No | Page number for visible signature (1-based, default: 1) |
| `x` | float | No | X position for visible signature |
| `y` | float | No | Y position for visible signature |
| `width` | float | No | Width of visible signature (default: 200) |
| `height` | float | No | Height of visible signature (default: 50) |

### Example

```json
{
  "config": {
    "signature": {
      "enabled": true,
      "visible": true,
      "name": "John Doe",
      "reason": "Document Approval",
      "location": "New York, US",
      "contactInfo": "john@example.com",
      "privateKeyPem": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
      "certificatePem": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
      "certificateChain": [
        "-----BEGIN CERTIFICATE-----\n...intermediate...\n-----END CERTIFICATE-----",
        "-----BEGIN CERTIFICATE-----\n...root...\n-----END CERTIFICATE-----"
      ]
    }
  }
}
```

---

## Security & Encryption

Password-protect documents with granular permission controls.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `enabled` | bool | Yes | Enable encryption |
| `ownerPassword` | string | Yes | Password for full document access |
| `userPassword` | string | No | Password to open document (empty = no password to open) |
| `allowPrinting` | bool | No | Allow document printing |
| `allowModifying` | bool | No | Allow content modification |
| `allowCopying` | bool | No | Allow copying text/images |
| `allowAnnotations` | bool | No | Allow adding annotations |
| `allowFormFilling` | bool | No | Allow filling form fields |
| `allowAccessibility` | bool | No | Allow accessibility features |
| `allowAssembly` | bool | No | Allow document assembly |
| `allowHighQualityPrint` | bool | No | Allow high quality printing |

### Example

```json
{
  "config": {
    "security": {
      "enabled": true,
      "ownerPassword": "admin123",
      "userPassword": "view123",
      "allowPrinting": true,
      "allowCopying": false,
      "allowModifying": false,
      "allowFormFilling": true
    }
  }
}
```

---

## Title Object

Defines the document header/title section.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `props` | string | Yes | Styling properties (see [Props Syntax](#props-syntax)) |
| `text` | string | Yes* | Title text (*ignored if `table` is provided) |
| `table` | object | No | Embedded table for complex layouts (e.g., logo + text) |
| `bgcolor` | string | No | Background color (hex: `"#RRGGBB"`) |
| `textcolor` | string | No | Text color (hex: `"#RRGGBB"`) |
| `link` | string | No | External URL hyperlink |

### Title with Embedded Table

```json
{
  "title": {
    "props": "Helvetica:18:100:center:0:0:1:0",
    "table": {
      "maxcolumns": 2,
      "columnwidths": [1, 3],
      "rows": [
        {
          "row": [
            { "props": "Helvetica:12:000:center:0:0:0:0", "image": { "imagedata": "base64...", "width": 50, "height": 50 } },
            { "props": "Helvetica:18:100:left:0:0:0:0", "text": "Company Name" }
          ]
        }
      ]
    }
  }
}
```

### Title with Colors and Link

```json
{
  "title": {
    "props": "Helvetica:24:100:center:0:0:0:0",
    "text": "FINANCIAL REPORT",
    "bgcolor": "#154360",
    "textcolor": "#FFFFFF",
    "link": "https://example.com/report"
  }
}
```

---

## Table Object

Tables are the primary content containers.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `maxcolumns` | int | Yes | Number of columns |
| `rows` | array | Yes | Array of row objects |
| `columnwidths` | array | No | Relative column width weights (e.g., `[2,1,1]` = 50%, 25%, 25%) |
| `rowheights` | array | No | Row heights in points (default: 25) |
| `bgcolor` | string | No | Default background color for all cells |
| `textcolor` | string | No | Default text color for all cells |

### Example

```json
{
  "elements": [
    {
      "type": "table",
      "table": {
        "maxcolumns": 4,
        "columnwidths": [1.2, 2, 1.2, 2],
        "rows": [
          {
            "row": [
              { "props": "Helvetica:10:100:left:1:0:1:1", "text": "Company Name:", "bgcolor": "#F4F6F7" },
              { "props": "Helvetica:10:000:left:0:0:1:1", "text": "TechCorp Inc.", "link": "https://techcorp.example.com" },
              { "props": "Helvetica:10:100:left:0:0:1:1", "text": "Report Period:" },
              { "props": "Helvetica:10:000:left:0:1:1:1", "text": "Q4 2025" }
            ]
          }
        ]
      }
    }
  ]
}
```

---

## Cell Object

Individual cells within table rows.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `props` | string | Yes | Styling properties |
| `text` | string | No | Cell text content |
| `chequebox` | bool | No | Render checkbox (`true`=checked, `false`=unchecked) |
| `image` | object | No | Embedded image |
| `form_field` | object | No | Interactive form field |
| `width` | float | No | Cell width weight |
| `height` | float | No | Cell height in points |
| `bgcolor` | string | No | Cell background color (overrides table) |
| `textcolor` | string | No | Cell text color (overrides table) |
| `link` | string | No | Hyperlink - external URL or internal `#destination` |
| `dest` | string | No | Named destination anchor for internal links |

### Cell Types

```json
// Text cell
{ "props": "Helvetica:12:000:left:1:1:1:1", "text": "Hello World" }

// Checkbox cell
{ "props": "Helvetica:12:000:center:1:1:1:1", "chequebox": true }

// Image cell
{ "props": "Helvetica:12:000:center:1:1:1:1", "image": { "imagedata": "base64...", "width": 100, "height": 50 } }

// Colored cell
{ "props": "Helvetica:12:100:center:1:1:1:1", "text": "Alert!", "bgcolor": "#FF0000", "textcolor": "#FFFFFF" }

// External link cell
{ "props": "Helvetica:10:000:left:1:1:1:1", "text": "Visit Website", "link": "https://example.com", "textcolor": "#0000FF" }

// Internal link cell (jumps to destination)
{ "props": "Helvetica:10:000:left:1:1:1:1", "text": "Go to Summary", "link": "#financial-summary", "textcolor": "#0000FF" }

// Destination anchor cell (target for internal links)
{ "props": "Helvetica:12:100:left:1:1:1:1", "text": "FINANCIAL SUMMARY", "dest": "financial-summary", "bgcolor": "#21618C", "textcolor": "#FFFFFF" }
```

---

## Props Syntax

The `props` string defines text styling and cell borders.

### Format

```
"fontname:fontsize:style:alignment:left:right:top:bottom"
```

### Components

| Position | Name | Values | Description |
|----------|------|--------|-------------|
| 1 | `fontname` | `font1`, `font2`, etc. | Font identifier |
| 2 | `fontsize` | Integer | Font size in points |
| 3 | `style` | 3-digit code | Bold/Italic/Underline flags |
| 4 | `alignment` | `left`, `center`, `right` | Text alignment |
| 5 | `left` | `0` or `1` | Left border (0=none, 1=draw) |
| 6 | `right` | `0` or `1` | Right border |
| 7 | `top` | `0` or `1` | Top border |
| 8 | `bottom` | `0` or `1` | Bottom border |

### Style Code (3-digit)

| Code | Meaning |
|------|---------|
| `000` | Normal text |
| `100` | **Bold** |
| `010` | *Italic* |
| `001` | Underlined |
| `110` | **Bold + Italic** |
| `101` | **Bold + Underlined** |
| `011` | *Italic + Underlined* |
| `111` | **Bold + Italic + Underlined** |

### Examples

```
"font1:12:000:left:1:1:1:1"   → Normal, all borders
"font1:14:100:center:0:0:1:0" → Bold, centered, top border only
"font1:10:111:right:0:0:0:0"  → Bold+Italic+Underline, right-aligned, no borders
```

---

## Form Fields

Interactive form fields for fillable PDFs.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `type` | string | Yes | `"checkbox"`, `"radio"`, or `"text"` |
| `name` | string | Yes | Field name for data extraction |
| `value` | string | No | Export value (checkbox/radio) or default text |
| `checked` | bool | No | Initial checked state |
| `group_name` | string | No | Radio button group name |
| `shape` | string | No | `"round"` or `"square"` (for radio buttons) |

### Examples

```json
// Checkbox field
{
  "props": "Helvetica:12:000:center:1:1:1:1",
  "form_field": {
    "type": "checkbox",
    "name": "agree_terms",
    "value": "Yes",
    "checked": false
  }
}

// Radio button group
{
  "props": "Helvetica:12:000:center:1:1:1:1",
  "form_field": {
    "type": "radio",
    "name": "gender",
    "value": "male",
    "group_name": "gender_group",
    "shape": "round"
  }
}
```

---

## Bookmarks & Navigation

Create document outlines and internal navigation links.

### Bookmark Object

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `title` | string | Yes | Display text in bookmark panel |
| `page` | int | No | Target page number (1-based) |
| `dest` | string | No | Named destination (matches cell `dest` field) |
| `y` | float | No | Y position on target page (from top) |
| `children` | array | No | Nested bookmarks for hierarchical structure |
| `open` | bool | No | Whether children are expanded by default |

### Example

```json
{
  "bookmarks": [
    {
      "title": "Financial Report",
      "page": 1,
      "children": [
        { "title": "Company Information", "page": 1 },
        { "title": "Financial Summary", "dest": "financial-summary" }
      ]
    },
    {
      "title": "Charts & Visualizations",
      "page": 2,
      "dest": "charts-section"
    }
  ]
}
```

### Creating Navigation Links

1. **Add destination anchor** to a cell that will be the target:
```json
{ "props": "Helvetica:12:100:left:1:1:1:1", "text": "SECTION B: FINANCIAL SUMMARY", "dest": "financial-summary" }
```

2. **Create link** to jump to that destination:
```json
{ "props": "Helvetica:10:000:left:0:0:0:1", "text": "Go to Financial Summary", "link": "#financial-summary", "textcolor": "#0000FF" }
```

3. **Add bookmark** (optional) for sidebar navigation:
```json
{ "title": "Financial Summary", "dest": "financial-summary" }
```

---

## Images

Embed images in cells or as standalone elements.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `imagename` | string | No | Image identifier |
| `imagedata` | string | Yes | Base64-encoded image data |
| `width` | float | Yes | Image width in points |
| `height` | float | Yes | Image height in points |

### Example

```json
{
  "props": "Helvetica:12:000:center:0:0:0:0",
  "image": {
    "imagename": "logo",
    "imagedata": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
    "width": 100,
    "height": 50
  }
}
```

### Image with Link

```json
{
  "props": "Helvetica:9:000:center:1:1:0:1",
  "height": 200,
  "link": "https://example.com/charts/bar",
  "image": {
    "imagename": "bar_chart",
    "imagedata": "base64data...",
    "width": 200,
    "height": 200
  }
}
```

---

## Footer Object

Appears at the bottom of every page.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `font` | string | Yes | Font props: `"fontname:fontsize:style:alignment"` |
| `text` | string | Yes | Footer text |
| `link` | string | No | External URL hyperlink |

> **Note:** Page numbers ("Page X of Y") are automatically added to the bottom-right corner.

### Example

```json
{
  "footer": {
    "font": "Helvetica:8:000:center",
    "text": "TECHCORP INDUSTRIES INC. | FINANCIAL REPORT Q4 2025 | CONFIDENTIAL",
    "link": "https://example.com/legal"
  }
}
```

---

## Complete Example

A comprehensive financial report with digital signature, bookmarks, and internal navigation:

```json
{
  "config": {
    "pageBorder": "0:0:0:0",
    "page": "A4",
    "pageAlignment": 1,
    "pdfTitle": "Financial Report Q4 2025",
    "pdfaCompliant": true,
    "arlingtonCompatible": true,
    "embedFonts": true,
    "signature": {
      "enabled": true,
      "visible": true,
      "name": "John Doe",
      "reason": "Document Approval",
      "location": "US",
      "contactInfo": "john@example.com",
      "privateKeyPem": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
      "certificatePem": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----"
    }
  },
  "title": {
    "props": "Helvetica:24:100:center:0:0:0:0",
    "text": "FINANCIAL REPORT",
    "bgcolor": "#154360",
    "textcolor": "#FFFFFF",
    "link": "https://example.com/report"
  },
  "elements": [
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "columnwidths": [1],
        "rows": [
          {
            "row": [
              { "props": "Helvetica:12:100:left:1:1:1:1", "text": "SECTION A: COMPANY INFORMATION", "bgcolor": "#21618C", "textcolor": "#FFFFFF" }
            ]
          }
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 4,
        "columnwidths": [1.2, 2, 1.2, 2],
        "rows": [
          {
            "row": [
              { "props": "Helvetica:10:100:left:1:0:0:1", "text": "Company Name:", "bgcolor": "#F4F6F7" },
              { "props": "Helvetica:10:000:left:0:0:0:1", "text": "TechCorp Industries Inc.", "link": "https://techcorp.example.com", "bgcolor": "#F4F6F7" },
              { "props": "Helvetica:10:100:left:0:0:0:1", "text": "Report Period:", "bgcolor": "#F4F6F7" },
              { "props": "Helvetica:10:000:left:0:1:0:1", "text": "Q4 2025", "bgcolor": "#F4F6F7" }
            ]
          },
          {
            "row": [
              { "props": "Helvetica:12:000:left:1:0:0:1", "text": "" },
              { "props": "Helvetica:10:000:left:0:0:0:1", "text": "Go to Financial Summary", "textcolor": "#0000FF", "link": "#financial-summary" },
              { "props": "Helvetica:10:000:left:0:0:0:1", "text": "Go to Charts", "textcolor": "#0000FF", "link": "#charts-section" },
              { "props": "Helvetica:12:000:left:0:1:0:1", "text": "" }
            ]
          }
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "columnwidths": [1],
        "rows": [
          {
            "row": [
              { "props": "Helvetica:12:100:left:1:1:1:1", "text": "SECTION B: FINANCIAL SUMMARY", "bgcolor": "#21618C", "textcolor": "#FFFFFF", "dest": "financial-summary" }
            ]
          }
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 2,
        "columnwidths": [2, 1],
        "rows": [
          {
            "row": [
              { "props": "Helvetica:10:000:left:1:0:0:1", "text": "Total Revenue" },
              { "props": "Helvetica:10:000:right:0:1:0:1", "text": "$2,450,000" }
            ]
          },
          {
            "row": [
              { "props": "Helvetica:10:100:left:1:0:0:1", "text": "Gross Profit", "bgcolor": "#D4E6F1" },
              { "props": "Helvetica:10:100:right:0:1:0:1", "text": "$1,225,000", "bgcolor": "#D4E6F1" }
            ]
          },
          {
            "row": [
              { "props": "Helvetica:11:100:left:1:0:1:1", "text": "Net Income", "bgcolor": "#A9CCE3" },
              { "props": "Helvetica:11:100:right:0:1:1:1", "text": "$125,000", "bgcolor": "#A9CCE3" }
            ]
          }
        ]
      }
    },
    {
      "type": "table",
      "table": {
        "maxcolumns": 1,
        "columnwidths": [1],
        "rows": [
          {
            "row": [
              { "props": "Helvetica:12:100:left:1:1:1:1", "text": "SECTION C: CHARTS", "bgcolor": "#21618C", "textcolor": "#FFFFFF", "dest": "charts-section" }
            ]
          }
        ]
      }
    }
  ],
  "footer": {
    "font": "Helvetica:8:000:center",
    "text": "TECHCORP INDUSTRIES INC. | FINANCIAL REPORT Q4 2025 | CONFIDENTIAL",
    "link": "https://example.com/legal"
  },
  "bookmarks": [
    {
      "title": "Financial Report",
      "page": 1,
      "children": [
        { "title": "Company Information", "page": 1 },
        { "title": "Financial Summary", "dest": "financial-summary" }
      ]
    },
    {
      "title": "Charts",
      "page": 2,
      "dest": "charts-section"
    }
  ]
}
```

---

## API Usage

### Generate PDF

```bash
curl -X POST "http://localhost:8080/api/v1/generate/template-pdf" \
  -H "Content-Type: application/json" \
  -d @template.json \
  --output document.pdf
```

### Load Template Data

```bash
curl "http://localhost:8080/api/v1/template-data?file=temp_multiplepage.json"
```

### Response

- **Content-Type:** `application/pdf`
- **Filename:** `template-pdf-<timestamp>.pdf`

---

## Automatic Features

| Feature | Description |
|---------|-------------|
| **Page Breaks** | Content automatically flows to new pages when needed |
| **Page Numbering** | "Page X of Y" added to bottom-right of each page |
| **Border Preservation** | Page borders drawn on every page |
| **Watermarks** | Diagonal watermark rendered on all pages |
| **Height Tracking** | System monitors content height for optimal pagination |
