# 📋 PDF Template Reference

Complete guide to the JSON template format used by GoPdfSuit for generating PDF documents.

---

## Table of Contents

- [Overview](#overview)
- [Template Structure](#template-structure)
- [Config Object](#config-object)
- [Title Object](#title-object)
- [Table Object](#table-object)
- [Cell Object](#cell-object)
- [Props Syntax](#props-syntax)
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
  "table": [ ],       // Array of tables (required)
  "spacer": [ ],      // Array of spacers (optional)
  "image": [ ],       // Array of images (optional)
  "elements": [ ],    // Ordered elements (optional)
  "footer": { }       // Page footer (required)
}
```

---

## Config Object

Controls page layout and appearance.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `page` | string | No | `"A4"` | Page size |
| `pageAlignment` | int | No | `1` | Orientation: `1`=Portrait, `2`=Landscape |
| `pageBorder` | string | No | `"0:0:0:0"` | Border widths: `"left:right:top:bottom"` |
| `watermark` | string | No | - | Diagonal watermark text across all pages |
| `arlingtonCompatible` | bool | No | `false` | Enable PDF 2.0 Arlington Model compliance |

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
    "pageBorder": "1:1:1:1",
    "watermark": "CONFIDENTIAL"
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

### Title with Embedded Table

```json
{
  "title": {
    "props": "font1:18:100:center:0:0:1:0",
    "table": {
      "maxcolumns": 2,
      "columnwidths": [1, 3],
      "rows": [
        {
          "row": [
            { "props": "font1:12:000:center:0:0:0:0", "image": { "imagedata": "base64...", "width": 50, "height": 50 } },
            { "props": "font1:18:100:left:0:0:0:0", "text": "Company Name" }
          ]
        }
      ]
    }
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
  "table": [
    {
      "maxcolumns": 4,
      "columnwidths": [1, 2, 1, 2],
      "bgcolor": "#F5F5F5",
      "rows": [
        {
          "row": [
            { "props": "font1:12:100:left:1:0:1:1", "text": "Name:" },
            { "props": "font1:12:000:left:0:1:1:1", "text": "John Doe" },
            { "props": "font1:12:100:left:1:0:1:1", "text": "DOB:" },
            { "props": "font1:12:000:left:0:1:1:1", "text": "01/15/1990" }
          ]
        }
      ]
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

### Cell Types

```json
// Text cell
{ "props": "font1:12:000:left:1:1:1:1", "text": "Hello World" }

// Checkbox cell
{ "props": "font1:12:000:center:1:1:1:1", "chequebox": true }

// Image cell
{ "props": "font1:12:000:center:1:1:1:1", "image": { "imagedata": "base64...", "width": 100, "height": 50 } }

// Colored cell
{ "props": "font1:12:100:center:1:1:1:1", "text": "Alert!", "bgcolor": "#FF0000", "textcolor": "#FFFFFF" }
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
  "props": "font1:12:000:center:1:1:1:1",
  "form_field": {
    "type": "checkbox",
    "name": "agree_terms",
    "value": "Yes",
    "checked": false
  }
}

// Radio button group
{
  "props": "font1:12:000:center:1:1:1:1",
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
  "props": "font1:12:000:center:0:0:0:0",
  "image": {
    "imagename": "logo",
    "imagedata": "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==",
    "width": 100,
    "height": 50
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

> **Note:** Page numbers ("Page X of Y") are automatically added to the bottom-right corner.

### Example

```json
{
  "footer": {
    "font": "font1:10:000:center",
    "text": "Confidential Document - Do Not Distribute"
  }
}
```

---

## Complete Example

```json
{
  "config": {
    "page": "A4",
    "pageAlignment": 1,
    "pageBorder": "1:1:1:1",
    "watermark": "SAMPLE"
  },
  "title": {
    "props": "font1:20:100:center:0:0:1:0",
    "text": "Patient Registration Form"
  },
  "table": [
    {
      "maxcolumns": 4,
      "columnwidths": [1, 2, 1, 2],
      "rows": [
        {
          "row": [
            { "props": "font1:12:100:left:1:0:1:1", "text": "Full Name:" },
            { "props": "font1:12:000:left:0:1:1:1", "text": "Jane Smith" },
            { "props": "font1:12:100:left:1:0:1:1", "text": "DOB:" },
            { "props": "font1:12:000:left:0:1:1:1", "text": "03/15/1985" }
          ]
        },
        {
          "row": [
            { "props": "font1:12:100:left:1:0:1:1", "text": "Gender:" },
            { "props": "font1:12:000:center:0:0:1:1", "chequebox": false, "text": "Male" },
            { "props": "font1:12:000:left:0:0:1:1", "text": "" },
            { "props": "font1:12:000:center:0:1:1:1", "chequebox": true, "text": "Female" }
          ]
        }
      ]
    },
    {
      "maxcolumns": 2,
      "rows": [
        {
          "row": [
            { "props": "font1:14:100:left:1:1:1:1", "text": "Medical History" },
            { "props": "font1:14:000:left:1:1:1:1", "text": "" }
          ]
        },
        {
          "row": [
            { "props": "font1:12:000:left:1:1:1:1", "text": "Allergies: Penicillin" },
            { "props": "font1:12:000:left:1:1:1:1", "text": "Medications: None" }
          ]
        }
      ]
    }
  ],
  "footer": {
    "font": "font1:10:010:center",
    "text": "Confidential Medical Record"
  }
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
