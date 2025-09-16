# 📄 GoPdfSuit

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-Web%20Framework-00ADD8?style=flat)](https://gin-gonic.com/)
[![Docker](https://img.shields.io/badge/Docker-Container-2496ED?style=flat&logo=docker)](https://hub.docker.com/)
[![gochromedp](https://img.shields.io/badge/gochromedp-1.0+-00ADD8?style=flat)](https://github.com/chinmay-sawant/gochromedp)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> 🚀 A powerful Go web service that generates template-based PDF documents on-the-fly with **multi-page support**, **custom page sizes**, **automatic page breaks**, **PDF merge capabilities**, and **HTML to PDF/Image conversion**.

## 📖 Overview

GoPdfSuit is a flexible web service built with Go and the Gin framework. It features a custom template-based PDF generator that creates professional documents from JSON templates, supporting **multiple page sizes**, **automatic page breaks**, **PDF merging**, **form filling**, tables, borders, checkboxes, **font styling (bold, italic, underline)**, and custom layouts without external dependencies.

**New Features:** HTML to PDF and Image conversion using gochromedp with web interfaces and REST APIs.

## 🔧 Requirements

- **Go** `1.20+` (project currently targets Go 1.23)
- **gochromedp** `1.0+` (for HTML to PDF/Image conversion)
- **Dependencies**: Automatically managed via Go modules

## ⚡ Quick Start

### 1️⃣ Install System Dependencies

**For HTML to PDF/Image conversion:**
```bash
# Ubuntu/Debian
sudo apt-get install chromium-browser

# macOS (with Homebrew)
brew install chromium

# Windows - Download Chrome from: https://www.google.com/chrome/
```

**Note:** gochromedp uses Chromium headless for HTML conversion, so Chrome/Chromium must be installed.

### 2️⃣ Clone the Repository
```bash
git clone https://github.com/chinmay-sawant/gopdfsuit.git
cd gopdfsuit
```

### 3️⃣ Install Dependencies
```bash
go mod download

# Build frontend (React/Vite)
cd frontend
npm install
npm run build
cd ..
```

### 4️⃣ Run the Server
```bash
# From repository root
go run ./cmd/gopdfsuit
```

### 5️⃣ Server Running
```
🌐 Server listening on: http://localhost:8080
```

## 🐳 Docker Deployment

GoPdfSuit can be easily deployed using Docker for containerized environments.

### Prerequisites
- Docker installed on your system

### Makefile Variables

You can customize the Docker build using environment variables:

```bash
# Set custom version
export VERSION=1.0.0

# Set custom Docker Hub username
export DOCKERUSERNAME=yourusername

# Build with custom settings
make docker
```

Default values:
- `VERSION=1.0.0`
- `DOCKERUSERNAME=chinmaysawant`

### Available Makefile Targets

- `make docker` - Build and run Docker container
- `make dockertag` - Tag and push to Docker Hub
- `make build` - Build Go application
- `make test` - Run tests
- `make clean` - Clean build artifacts
- `make run` - Run application locally
- `make fmt` - Format Go code
- `make vet` - Run Go vet
- `make mod` - Tidy Go modules
- `make pull` - Pull and run Docker image from Docker Hub 

### Build and Run with Docker

1. **Clone the repository** (if not already done):
   ```bash
   git clone https://github.com/chinmay-sawant/gopdfsuit.git
   cd gopdfsuit
   ```

2. **Build the Docker image**:
   ```bash
   make docker
   # Or manually:
   # docker build -f dockerfolder/Dockerfile --build-arg VERSION=1.0.0 -t gopdfsuit:1.0.0 .
   # docker run -d -p 8080:8080 gopdfsuit:1.0.0
   ```

3. **Access the application**:
   - Open your browser to `http://localhost:8080`

### Alternative Docker Build Options

GoPdfSuit provides multiple Dockerfile options for different gochromedp compatibility needs:

#### Standard Build (Recommended)
```bash
docker build -f dockerfolder/Dockerfile --build-arg VERSION=1.0.0 -t gopdfsuit:1.0.0 .
```

#### Troubleshooting gochromedp/Chromium Installation

If you encounter issues with gochromedp or Chromium installation:

1. **Ensure Chromium/Chrome is installed** and accessible
2. **Try the Ubuntu-based build** (Option 2) for better compatibility
3. **Use the modern build** (Option 3) if you don't need specific Chromium versions
4. **Check the standard build** (Option 1) which includes multiple fallback methods

### Docker Hub

The image is available on Docker Hub:

- **Repository**: `chinmaysawant/gopdfsuit` (configurable via `DOCKERUSERNAME`)
- **Latest Version**: `chinmaysawant/gopdfsuit:latest`

#### Pull from Docker Hub
```bash
docker pull chinmaysawant/gopdfsuit:latest
docker run -d -p 8080:8080 chinmaysawant/gopdfsuit:latest
```

#### Run Docker Image Using Make 
```bash
make pull 
```

#### Push to Docker Hub (for maintainers)
```bash
make docker && make dockertag
```

### Docker Configuration

- **Port**: 8080 (configurable via `EXPOSE` in Dockerfile)
- **Base Images**:
  - `Dockerfile`: Alpine Linux (minimal, with gochromedp and Chromium)
- **Multi-stage Build**: Optimized for smaller final image size
- **Included Files**: Web assets, LICENSE, README, and sample templates

## 📡 API Reference

### PDF Viewer Web Interface

**New Feature:** Interactive web-based PDF viewer and template editor.

**Endpoint:** `GET /` (Root endpoint)

**Query Parameters:**
- `file` (optional): JSON template filename to load automatically

**Examples:**
```
http://localhost:8080/
http://localhost:8080/?file=temp_multiplepage.json
```

### PDF Template Editor

**New Feature:** Visual drag-and-drop PDF template editor.

**Endpoint:** `GET /editor`

**Query Parameters:**
- `file` (optional): JSON template filename to load automatically

**Examples:**
```
http://localhost:8080/editor
http://localhost:8080/editor?file=temp_multiplepage.json
```

**Features:**
- 🎨 **Drag-and-Drop Interface**: Visual template building with drag-and-drop components
- 📋 **Real-time JSON Generation**: Live JSON template generation as you build
- 🔧 **Component Properties**: Editable properties panel for each component
- 📄 **Live PDF Preview**: Generate and preview PDFs instantly
- 💾 **Template Loading**: Load existing templates for editing
- 📱 **Responsive Design**: Works on desktop, tablet, and mobile devices
- 🎨 **Theme Support**: Multiple gradient themes and dark/light mode

### Template Data API

**Endpoint:** `GET /api/v1/template-data`

**Query Parameters:**
- `file` (required): JSON template filename

**Security Features:**
- ✅ **Path Traversal Protection**: Only filenames (no directories) allowed
- ✅ **File Extension Validation**: Only `.json` files accepted
- ✅ **JSON Validation**: Template structure validation before serving

**Example:**
```bash
curl "http://localhost:8080/api/v1/template-data?file=temp_multiplepage.json"
```

### Generate Template-based PDF

**Endpoint:** `POST /api/v1/generate/template-pdf`

**Headers:**
- `Content-Type: application/json`

**Request Body Structure:**
```json
{
  "config": {
    "pageBorder": "1:1:1:1",
    "page": "A4",
    "pageAlignment": 1
  },
  "title": {
    "props": "font1:24:100:center:0:0:1:0",
    "text": "Multi-Page Document Title"
  },
  "table": [
    {
      "maxcolumns": 4,
      "rows": [
        {
          "row": [
            {
              "props": "font1:12:100:left:1:1:1:1",
              "text": "Bold Field Name:"
            },
            {
              "props": "font1:12:000:left:1:1:1:1",
              "text": "Normal Field Value"
            },
            {
              "props": "font1:12:010:left:1:1:1:1",
              "text": "Italic Text"
            },
            {
              "props": "font1:12:111:right:1:1:1:1",
              "text": "Bold+Italic+Underline"
            }
          ]
        }
      ]
    }
  ],
  "footer": {
    "font": "font1:10:001:center",
    "text": "Multi-page Footer"
  }
}
```

**Template Configuration Properties:**

- **config.pageBorder**: `"left:right:top:bottom"` - Border widths for page edges
- **config.page**: Page size specification
  - `"A4"` - 8.27 × 11.69 inches (595 × 842 points) - **Default**
  - `"LETTER"` - 8.5 × 11 inches (612 × 792 points)
  - `"LEGAL"` - 8.5 × 14 inches (612 × 1008 points)
  - `"A3"` - 11.69 × 16.54 inches (842 × 1191 points)
  - `"A5"` - 5.83 × 8.27 inches (420 × 595 points)
- **config.pageAlignment**: Page orientation
  - `1` - **Portrait** (vertical) - **Default**
  - `2` - **Landscape** (horizontal)
- **config.watermark**: (optional) Text rendered diagonally (bottom-left to top-right) in light gray across every page. Automatically sized proportionally to page size.

**Template Properties Explained:**

- **props**: `"fontname:fontsize:style:alignment:left:right:top:bottom"`
  - `fontname`: Font identifier (font1, font2, etc.)
  - `fontsize`: Font size in points
  - `style`: **3-digit style code** for text formatting:
    - **First digit (Bold)**: `1` = bold, `0` = normal weight
    - **Second digit (Italic)**: `1` = italic, `0` = normal style  
    - **Third digit (Underline)**: `1` = underlined, `0` = no underline
    - Examples:
      - `000` = Normal text
      - `100` = **Bold** text
      - `010` = *Italic* text
      - `001` = <u>Underlined</u> text
      - `110` = ***Bold + Italic***
      - `101` = **<u>Bold + Underlined</u>**
      - `011` = *<u>Italic + Underlined</u>*
      - `111` = ***<u>Bold + Italic + Underlined</u>***
  - `alignment`: left, center, or right
  - `left:right:top:bottom`: Border widths for cell edges
- **chequebox**: Boolean value for checkbox state (true = checked, false = unchecked)

**Automatic Page Break Features:**
- ✅ **Height Tracking**: Monitors content height and automatically creates new pages
- ✅ **Page Size Aware**: Respects selected page dimensions for break calculations
- ✅ **Border Preservation**: Page borders are drawn on every new page
- ✅ **Content Continuity**: Tables and content flow seamlessly across pages
- ✅ **Page Numbering**: Automatic "Page X of Y" numbering in bottom right corner

**Response:**
- **Content-Type:** `application/pdf`
- **File:** `template-pdf-<timestamp>.pdf` (auto-download)

### PDF Merge

**New Feature:** Combine multiple PDF files into a single document with drag-and-drop interface.

**Endpoint:** `POST /api/v1/merge`

**Web Interface:** `GET /merge`

**Headers:**
- `Content-Type: multipart/form-data`

**Form Data Parameters:**
- `pdf` (required): One or more PDF files to merge (repeatable)

**Features:**
- 🎯 **Drag & Drop Interface**: Intuitive file upload with visual feedback
- 🔄 **File Reordering**: Drag files to change merge order before processing
- 👁️ **Live Preview**: Preview merged PDF with page navigation
- 📱 **Responsive Design**: Works on desktop, tablet, and mobile devices
- 🎨 **Theme Support**: Multiple gradient themes and dark/light mode

**Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/merge" \
  -F "pdf=@file1.pdf" \
  -F "pdf=@file2.pdf" \
  -F "pdf=@file3.pdf" \
  --output merged.pdf
```

**Web Interface Access:**
```
http://localhost:8080/merge
```

### HTML to PDF Converter

**New Feature:** Convert HTML content or web pages to PDF using gochromedp with full control over page settings.

**Endpoint:** `GET /htmltopdf`

**API Endpoint:** `POST /api/v1/htmltopdf`

**Headers:** `Content-Type: application/json`

**Request Body:**
```json
{
  "html": "<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>",
  "url": "https://example.com",
  "page_size": "A4",
  "orientation": "Portrait",
  "margin_top": "10mm",
  "margin_right": "10mm",
  "margin_bottom": "10mm",
  "margin_left": "10mm",
  "dpi": 300,
  "grayscale": false,
  "low_quality": false
}
```

**Features:**
- ✅ **HTML Content**: Convert raw HTML strings to PDF
- ✅ **URL Conversion**: Convert web pages directly to PDF
- ✅ **Page Control**: Custom page sizes (A4, Letter, Legal, A3)
- ✅ **Orientation**: Portrait and Landscape modes
- ✅ **Margins**: Precise margin control
- ✅ **Quality**: DPI settings for high-quality output
- ✅ **Color Modes**: Grayscale and color options
- ✅ **File Size**: Low quality option for smaller files

**Web Interface Access:**
```
http://localhost:8080/htmltopdf
```

### HTML to Image Converter

**New Feature:** Convert HTML content or web pages to images (PNG, JPG, SVG) using gochromedp.

**Endpoint:** `GET /htmltoimage`

**API Endpoint:** `POST /api/v1/htmltoimage`

**Headers:** `Content-Type: application/json`

**Request Body:**
```json
{
  "html": "<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>",
  "url": "https://example.com",
  "format": "png",
  "width": 800,
  "height": 600,
  "quality": 94,
  "zoom": 1.0,
  "crop_width": 400,
  "crop_height": 300,
  "crop_x": 0,
  "crop_y": 0
}
```

**Features:**
- ✅ **Multiple Formats**: PNG, JPG, SVG output
- ✅ **Dimensions**: Custom width and height
- ✅ **Quality Control**: 1-100 quality settings
- ✅ **Zoom**: Scale content with zoom factor
- ✅ **Cropping**: Precise image cropping options
- ✅ **Real-time Preview**: Live HTML preview before conversion

**Web Interface Access:**
```
http://localhost:8080/htmltoimage
```

### PDF Form Filling

**Endpoint:** `POST /api/v1/fill`

**Headers:**
- `Content-Type: multipart/form-data`

**Form Data Parameters:**
- `pdf` (required): The source PDF file
- `xfdf` (required): The XFDF file with field data

**Example:**
```bash
curl -X POST "http://localhost:8080/api/v1/fill" \
  -F "pdf=@patient.pdf" \
  -F "xfdf=@patient.xfdf" \
  --output filled.pdf
```

## 🧪 Usage Examples

### 🖥️ Web Interface Usage

1. **PDF Viewer:**
   ```
   http://localhost:8080/
   ```

2. **Template Editor:**
   ```
   http://localhost:8080/editor
   ```

3. **PDF Merger:**
   ```
   http://localhost:8080/merge
   ```

4. **PDF Filler:**
   ```
   http://localhost:8080/filler
   ```

5. **HTML to PDF Converter:**
   ```
   http://localhost:8080/htmltopdf
   ```

6. **HTML to Image Converter:**
   ```
   http://localhost:8080/htmltoimage
   ```

### 📱 Multi-Page Healthcare Form (Web Interface)

1. Navigate to: `http://localhost:8080/?file=temp_multiplepage.json`
2. The interface will automatically load and display the template
3. Click "Generate PDF" to create a multi-page healthcare form
4. Use the page navigation controls to browse through pages
5. Download the PDF using the download button

### 📱 Multi-Page Healthcare Form (cURL)
```bash
curl -X POST "http://localhost:8080/api/v1/generate/template-pdf" \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "pageBorder": "2:2:2:2",
      "page": "LETTER",
      "pageAlignment": 1
    },
    "title": {
      "props": "font1:20:110:center:0:0:2:0",
      "text": "Patient Encounter Form - Multi Page"
    },
    "table": [
      {
        "maxcolumns": 4,
        "rows": [
          {
            "row": [
              {
                "props": "font1:12:100:left:1:0:1:1",
                "text": "Patient Name:"
              },
              {
                "props": "font1:12:000:left:0:1:1:1",
                "text": "John Doe"
              },
              {
                "props": "font1:12:100:left:1:0:1:1",
                "text": "DOB:"
              },
              {
                "props": "font1:12:010:left:0:1:1:1",
                "text": "01/15/1980"
              }
            ]
          }
        ]
      }
    ],
    "footer": {
      "font": "font1:10:001:center",
      "text": "Confidential Medical Document - Auto Pagination"
    }
  }' \
  --output patient-form-multipage.pdf
```

### 🖼️ Landscape Layout Example (Python)
```python
import requests
import json

url = "http://localhost:8080/api/v1/generate/template-pdf"
template = {
    "config": {
        "pageBorder": "1:1:1:1",
        "page": "A4",
        "pageAlignment": 2  # Landscape orientation
    },
    "title": {
        "props": "font1:22:111:center:0:0:2:0",
        "text": "Landscape Survey Form"
    },
    "table": [
        {
            "maxcolumns": 6,  # More columns fit in landscape
            "rows": [
                {
                    "row": [
                        {
                            "props": "font1:14:100:left:1:1:1:1",
                            "text": "Question 1:"
                        },
                        {
                            "props": "font1:12:000:center:1:1:1:1",
                            "chequebox": True
                        },
                        {
                            "props": "font1:12:010:left:1:1:1:1",
                            "text": "Excellent"
                        },
                        {
                            "props": "font1:12:000:center:1:1:1:1",
                            "chequebox": False
                        },
                        {
                            "props": "font1:12:010:left:1:1:1:1",
                            "text": "Good"
                        },
                        {
                            "props": "font1:12:000:left:1:1:1:1",
                            "text": "Average"
                        }
                    ]
                }
            ]
        }
    ],
    "footer": {
        "font": "font1:10:001:right",
        "text": "Landscape Page Layout"
    }
}

response = requests.post(url, json=template)
with open("survey-landscape.pdf", "wb") as f:
    f.write(response.content)
```

### 📄 Large Document with Auto Page Breaks
```json
{
  "config": {
    "pageBorder": "1:1:1:1",
    "page": "LEGAL",
    "pageAlignment": 1
  },
  "title": {
    "props": "font1:18:100:center:0:0:1:0",
    "text": "Large Multi-Page Document"
  },
  "table": [
    {
      "maxcolumns": 2,
      "rows": [
        // Add many rows here - system will automatically create new pages
        {
          "row": [
            {
              "props": "font1:12:100:left:1:1:1:1",
              "text": "Section 1: Introduction"
            },
            {
              "props": "font1:12:000:left:1:1:1:1",
              "text": "This document demonstrates automatic page breaks..."
            }
          ]
        }
        // ... more rows will automatically flow to new pages
      ]
    }
  ],
  "footer": {
    "font": "font1:10:000:center",
    "text": "Document continues across multiple pages automatically"
  }
}
```

### 🌐 HTML to PDF Conversion (Web Interface)

1. Navigate to: `http://localhost:8080/htmltopdf`
2. Choose input type: HTML Content or URL
3. Enter your HTML or paste a URL
4. Configure PDF options (page size, margins, orientation)
5. Click "Convert to PDF" to download
6. Use "Preview HTML" to see how it will look before conversion

### 🌐 HTML to Image Conversion (Web Interface)

1. Navigate to: `http://localhost:8080/htmltoimage`
2. Choose input type: HTML Content or URL
3. Enter your HTML or paste a URL
4. Configure image options (format, dimensions, quality)
5. Set cropping options if needed
6. Click "Convert to Image" to download
7. Use "Preview HTML" to preview before conversion

### 📄 HTML to PDF Conversion (cURL)
```bash
curl -X POST "http://localhost:8080/api/v1/htmltopdf" \
  -H "Content-Type: application/json" \
  -d '{
    "html": "<!DOCTYPE html><html><head><title>Sample</title></head><body><h1>Hello World</h1><p>This is a sample PDF generated from HTML.</p></body></html>",
    "page_size": "A4",
    "orientation": "Portrait",
    "margin_top": "20mm",
    "margin_bottom": "20mm",
    "margin_left": "15mm",
    "margin_right": "15mm",
    "dpi": 300
  }' \
  --output sample.pdf
```

### 🖼️ HTML to Image Conversion (cURL)
```bash
curl -X POST "http://localhost:8080/api/v1/htmltoimage" \
  -H "Content-Type: application/json" \
  -d '{
    "html": "<!DOCTYPE html><html><body style=\"background: linear-gradient(45deg, #ff6b6b, #4ecdc4); color: white; text-align: center; padding: 50px;\"><h1>Beautiful Image</h1><p>Generated from HTML content</p></body></html>",
    "format": "png",
    "width": 800,
    "height": 600,
    "quality": 95,
    "zoom": 1.2
  }' \
  --output beautiful-image.png
```

### 🌐 URL to PDF Conversion (Python)
```python
import requests
import json

url = "http://localhost:8080/api/v1/htmltopdf"
data = {
    "url": "https://example.com",
    "page_size": "A4",
    "orientation": "Portrait",
    "margin_top": "10mm",
    "margin_right": "10mm",
    "margin_bottom": "10mm",
    "margin_left": "10mm",
    "dpi": 300,
    "grayscale": False
}

response = requests.post(url, json=data)
with open("website.pdf", "wb") as f:
    f.write(response.content)
```

## 🆚 Feature Comparison
| Feature | GoPdfSuit | UniPDF | Aspose.PDF | iText |
|---------|-----------|---------|------------|-------|
| **💰 Pricing** | **Free (MIT License)** | $3,990/dev/year | $1,999/dev/year | $3,800/dev/year |
| **🚀 Performance** | **Ultra Fast (179µs-1.7ms)** | Moderate | Moderate | Moderate |
| **📦 Deployment** | **Microservice/Sidecar/Docker** | Library Integration | Library Integration | Library Integration |
| **🧠 Memory Usage** | **In-Memory Processing** | File-based | Mixed | Mixed |
| **🔧 Integration** | **REST API (Language Agnostic)** | Go Library Only | .NET/Java/C++ | Java/.NET/Python |
| **📄 Template Engine** | **JSON-based Templates** | Code-based | Code-based | Code-based |
| **🌐 Web Interface** | **Built-in Viewer/Editor** | None | None | None |
| **📋 Form Filling** | **XFDF Support** | ✅ Full Support | ✅ Full Support | ✅ Full Support |
| **🔗 PDF Merge** | **Drag & Drop UI** | ✅ Programmatic | ✅ Programmatic | ✅ Programmatic |
| **🌐 HTML to PDF/Image** | **gochromedp (Chromium)** | Requires UniHTML product | Requires Aspose.HTML product | Requires custom integration |
| **📱 Multi-page Support** | **Auto Page Breaks** | ✅ Manual Control | ✅ Manual Control | ✅ Manual Control |
| **🎨 Styling** | **Font Styles + Borders** | ✅ Advanced | ✅ Advanced | ✅ Advanced |
| **☑️ Interactive Elements** | **Checkboxes** | ✅ Full Support | ✅ Full Support | ✅ Full Support |
| **📏 Page Formats** | **A3, A4, A5, Letter, Legal** | ✅ All Formats | ✅ All Formats | ✅ All Formats |
| **🔒 Security** | **Basic Validation** | ✅ Advanced | ✅ Enterprise | ✅ Enterprise |
| **🐳 Docker Support** | **✅ Built-in (Multi-stage, Alpine-based)** | ❌ Not Applicable (Library) | ❌ Not Applicable (Library) | ❌ Not Applicable (Library) |
| **🛠️ Maintenance** | **Single Binary** | Library Updates | Library Updates | Library Updates |

### 🏃‍♂️ Performance Benchmarks

**GoPdfSuit Performance (temp_multiplepage.json - 2 pages):**
```
[GIN] 2025/08/28 - 00:40:18 | 200 |       697.8µs | POST "/api/v1/generate/template-pdf"
[GIN] 2025/08/28 - 00:40:55 | 200 |      1.7542ms | POST "/api/v1/generate/template-pdf"  
[GIN] 2025/08/28 - 00:40:57 | 200 |       179.6µs | POST "/api/v1/generate/template-pdf"
[GIN] 2025/08/28 - 00:40:58 | 200 |       573.7µs | POST "/api/v1/generate/template-pdf"
```

**Key Advantages:**
- 🚀 **Ultra Fast**: Sub-millisecond to low-millisecond response times
- 💰 **Cost Effective**: MIT license vs $2k-$4k/developer/year
- 🔧 **Easy Deployment**: Microservice architecture vs library integration
- 🌐 **Language Agnostic**: REST API accessible from any programming language
- 📦 **Zero Dependencies**: Single binary deployment

**📊 [View Detailed Comparison →](comparison.html)**

## ✨ Features

- 🎯 **Template-based**: JSON-driven PDF generation
- ⚛️ **Modern Web Interface**: React-based SPA with real-time preview
- 🔗 **PDF Merge**: Combine multiple PDFs with drag-and-drop interface
- 🖊️ **Form Filling**: AcroForm/XFDF support for filling PDF forms
- 🌐 **HTML to PDF**: Convert HTML content or web pages to PDF with gochromedp
- 🖼️ **HTML to Image**: Convert HTML content or web pages to PNG/JPG/SVG with gochromedp
- 📋 **Tables & Forms**: Support for complex table layouts with automatic page breaks
- ☑️ **Checkboxes**: Interactive checkbox elements
- 🎨 **Font Styling**: Bold, italic, and underline text support
- 📄 **Multi-page Support**: Automatic page breaks and multi-page documents
- 🔢 **Page Numbering**: Automatic page numbering in "Page X of Y" format
- 📏 **Custom Page Sizes**: A4, Letter, Legal, A3, A5 support
- 🔄 **Page Orientation**: Portrait and landscape orientations
- 🔤 **Flexible Typography**: Custom fonts, sizes, and alignments
- 🔲 **Border Control**: Granular border configuration
- 🛡️ **Diagonal Watermark**: Optional per-template watermark text across all pages
- ⚡ **Fast**: In-memory PDF generation with height tracking
- 📦 **Self-contained**: Single binary deployment
- 🌐 **Cross-platform**: Runs on Windows, Linux, macOS
- 📱 **Responsive**: Mobile-friendly React interface with modern design
- 🔒 **Secure**: Path traversal protection and input validation
- 🚀 **Fast**: Vite-powered frontend with optimized builds

## 🏗️ Project Structure

```
GoPdfSuit/
├── 📁 cmd/
│   └── 📁 gopdfsuit/           # 🎯 Application entrypoint
│       └── main.go
├── 📁 frontend/                # ⚛️ React frontend (Vite)
│   ├── src/
│   │   ├── 📁 components/      # 🧩 React components
│   │   │   └── Navbar.jsx      # 🧭 Navigation component
│   │   ├── 📁 pages/           # 📄 Page components
│   │   │   ├── Home.jsx        # � Homepage with README content
│   │   │   ├── Viewer.jsx      # 👁️ PDF viewer and template processor
│   │   │   ├── Editor.jsx      # ✏️ PDF template editor
│   │   │   ├── Merge.jsx       # 🔗 PDF merge tool
│   │   │   ├── Filler.jsx      # 📝 PDF form filler
│   │   │   ├── HtmlToPdf.jsx   # 🌐 HTML to PDF converter
│   │   │   └── HtmlToImage.jsx # 🖼️ HTML to Image converter
│   │   ├── App.jsx             # ⚛️ Main app component with routing
│   │   ├── main.jsx            # 🎯 React entry point
│   │   └── index.css           # 🎨 Global styles
│   ├── package.json            # � Frontend dependencies
│   ├── vite.config.js          # ⚙️ Vite configuration
│   └── index.html              # 📄 HTML template
├── 📁 internal/
│   ├── 📁 handlers/            # 🔗 HTTP handlers and route registration
│   │   └── handlers.go
│   ├── 📁 models/              # 📊 Template data models
│   │   └── models.go
│   └── 📁 pdf/                 # 📄 PDF generation and processing
│       ├── pdf.go              # 🌐 HTML to PDF/Image conversion
│       ├── draw.go             # 🎨 Drawing utilities
│       ├── generator.go        # ⚙️ PDF generation orchestration
│       ├── merge.go            # 🔗 PDF merging
│       ├── pagemanager.go      # 📄 Page management
│       ├── types.go            # 📏 Page size definitions
│       ├── utils.go            # 🛠️ Helper utilities
│       ├── xfdf.go             # 🖊️ XFDF form filling
│       └── xfdf_example_test.go # 🧪 XFDF tests
├── 📁 docs/
│   └── 📁 assets/                # 🏗️ Built frontend assets (generated by Vite)
├── 📁 sampledata/              # 📋 Sample data and templates
│   ├── 📁 encounter_1.7/       # 🏥 Healthcare encounter samples
│   │   ├── us_encounter_form_data.xfdf
│   │   └── us_encounter_form.pdf
│   ├── 📁 json/                # 📄 JSON template samples
│   │   ├── temp_bold+italic+underline.json
│   │   ├── temp_multiplepage.json
│   │   ├── temp_og.json
│   │   └── temp.json
│   ├── 📁 patient2/            # 👤 Patient data samples
│   │   ├── filled.pdf
│   │   ├── jefferson_filled.pdf
│   │   └── patient2_chatgpt.xfdf
│   ├── 📁 patientreg/          # 📋 Patient registration samples
│   │   ├── patientreg_filled.pdf
│   │   ├── patientreg.pdf
│   │   └── patientreg.xfdf
│   └── 📁 pdf+xfdf/            # 📄 PDF and XFDF sample pairs
│       ├── us_hospital_encounter_acroform.pdf
│       └── us_hospital_encounter_data.xfdf
├── 📁 scripts/                 # 🛠️ Utility scripts
│   ├── add_dummy_values.go     # 📊 Add dummy data script
│   ├── flatten_pdf.go          # 📄 PDF flattening script
│   ├── flatten_pdf.py          # 🐍 Python PDF flattening script
│   └── 📁 __pycache__/         # 🐍 Python cache
├── 📁 tmp/                     # 🗂️ Temporary files
│   └── build-errors.log        # 📋 Build error logs
├── 📄 go.mod                   # 📦 Go modules file
├── 📄 go.sum                   # 📦 Go modules checksums
├── 📄 temp_multiplepage.json   # 📋 Example multi-page template
├── 📄 xfdf_debug.log           # 🐛 XFDF debugging logs
├── 📄 gopdfsuit.exe            # 🚀 Compiled Windows binary
├── 📄 .gitignore              # 🚫 Git ignore rules
└── 📖 README.md               # 📚 This documentation
```

## 🧩 XFDF / AcroForm filling (new)

This project includes a simple AcroForm/XFDF fill feature that accepts PDF bytes and XFDF (field data) and returns a filled PDF.

Endpoints and UI
- `POST /api/v1/fill` — accepts multipart/form-data with two file fields: `pdf` (the source PDF) and `xfdf` (the XFDF file). Returns `application/pdf` with the filled document as an attachment.
- `GET /filler` — simple web UI where users can upload a PDF and an XFDF file and download the filled PDF (uses the `/api/v1/fill` endpoint).

Quick curl example (multipart file upload):

```bash
curl -X POST "http://localhost:8080/api/v1/fill" \
  -F "pdf=@patient.pdf;type=application/pdf" \
  -F "xfdf=@patient.xfdf;type=application/xml" \
  --output filled.pdf
```

Server-run example (UI):

1. Start server from repo root:
```bash
go run ./cmd/gopdfsuit
```
2. Open `http://localhost:8080/filler` in your browser and upload PDF + XFDF.

Behaviour and limitations
- The filler uses a best-effort, byte-oriented approach implemented in the `internal/pdf` package: it parses XFDF, searches for AcroForm field names (heuristic `/T (name)`), and writes or inserts `/V (value)` tokens into the PDF bytes.
- For many simple AcroForm PDFs this works and the code sets `/NeedAppearances true` in the AcroForm so viewers regenerate appearances.
- Limitations: PDFs using compressed object streams, indirect references for field values, non-literal strings, or requiring generated appearance streams (`/AP`) may not render values correctly in all viewers. For robust, production-grade appearance updates, integrate a PDF library (e.g., pdfcpu or unidoc) to rebuild field appearance streams.

If you'd like, I can add a library-backed implementation that guarantees visual appearances across viewers.

## 🗺️ Roadmap & TODO

- [x] 🖥️ Web-based PDF viewer and template editor
- [x] 📋 Multi-page document support with automatic page breaks
- [x] 🔒 Security features (path traversal protection, input validation)
- [ ] 🧪 Add comprehensive unit tests
- [ ] 🎨 Support for colors and advanced styling
- [ ] 📊 Image embedding support
- [x] 🐳 Docker containerization
- [ ] 📈 Metrics and health check endpoints
- [ ] 🔐 Authentication and rate limiting
- [ ] 💾 Template storage and management
- [ ] 📧 Email delivery integration
- [ ] 📝 Template editor with validation
- [ ] 🔄 Real-time collaborative editing

## 🛠️ Development

### Building the Application
```bash
# Build for current platform
go build -o bin/gopdfsuit ./cmd/gopdfsuit

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o bin/gopdfsuit-linux ./cmd/gopdfsuit
GOOS=windows GOARCH=amd64 go build -o bin/gopdfsuit.exe ./cmd/gopdfsuit
```

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

**Page Break Logic:**
- The system tracks current Y position on each page
- When content would exceed page boundaries (considering margins), a new page is automatically created
- Each new page includes configured page borders
- Content flows seamlessly from one page to the next

**Supported Page Sizes:**
| Page Size | Dimensions (inches) | Dimensions (points) | Best For |
|-----------|-------------------|-------------------|----------|
| A4 | 8.27 × 11.69 | 595 × 842 | International standard |
| Letter | 8.5 × 11 | 612 × 792 | US standard |
| Legal | 8.5 × 14 | 612 × 1008 | Legal documents |
| A3 | 11.69 × 16.54 | 842 × 1191 | Large format |
| A5 | 5.83 × 8.27 | 420 × 595 | Small format |

## ⚠️ Production Notes

> **⚠️ Important:** The current PDF generator creates basic layouts suitable for forms and simple documents.

For production environments, consider:
- Implementing comprehensive input validation
- Adding request size limits
- Setting up proper logging and monitoring
- Implementing caching for frequently used templates
- Adding support for custom fonts and advanced layouts

## 🤝 Contributing

1. 🍴 Fork the repository
2. 🌟 Create a feature branch (`git checkout -b feature/amazing-feature`)
3. 💫 Commit your changes (`git commit -m 'Add amazing feature'`)
4. 📤 Push to the branch (`git push origin feature/amazing-feature`)
5. 🎉 Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

<div align="center">
  <p>Made with ❤️ and ☕ by <a href="https://github.com/chinmay-sawant">Chinmay Sawant</a></p>
  <p>⭐ Star this repo if you find it helpful!</p>
</div>