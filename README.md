# 📄 GoPdfSuit

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-Web%20Framework-00ADD8?style=flat)](https://gin-gonic.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> 🚀 A powerful Go web service that generates template-based PDF documents on-the-fly using JSON configurations with advanced font styling support.

## 📖 Overview

GoPdfSuit is a flexible web service built with Go and the Gin framework. It features a custom template-based PDF generator that creates professional documents from JSON templates, supporting tables, borders, checkboxes, **font styling (bold, italic, underline)**, and custom layouts without external dependencies.

## 🏗️ Project Structure

```
GoPdfSuit/
├── 📁 cmd/
│   └── 📁 gopdfsuit/           # 🎯 Application entrypoint
│       └── main.go
├── 📁 internal/
│   ├── 📁 handlers/            # 🔗 HTTP handlers and route registration
│   │   └── handlers.go
│   ├── 📁 models/              # 📊 Template data models
│   │   └── models.go
│   └── 📁 pdf/                 # 📄 Template-based PDF generation
│       └── pdf.go
├── 📄 go.mod                   # 📦 Go modules file
├── 📄 temp.json               # 📋 Example template file
├── 📄 .gitignore              # 🚫 Git ignore rules
└── 📖 README.md               # 📚 This file
```

## 🔧 Requirements

- **Go** `1.20+` (project currently targets Go 1.23)
- **Dependencies**: Automatically managed via Go modules

## ⚡ Quick Start

### 1️⃣ Clone the Repository
```bash
git clone https://github.com/chinmay-sawant/gopdfsuit.git
cd gopdfsuit
```

### 2️⃣ Install Dependencies
```bash
go mod download
```

### 3️⃣ Run the Server
```bash
# From repository root
go run ./cmd/gopdfsuit
```

### 4️⃣ Server Running
```
🌐 Server listening on: http://localhost:8080
```

## 📡 API Reference

### Generate Template-based PDF

**Endpoint:** `POST /api/v1/generate/template-pdf`

**Headers:**
- `Content-Type: application/json`

**Request Body Structure:**
```json
{
  "config": {
    "pageBorder": "1:1:1:1"
  },
  "title": {
    "props": "font1:24:100:center:0:0:1:0",
    "text": "Document Title"
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
    "font": "font1:10:001:right",
    "text": "Underlined Footer"
  }
}
```

**Template Properties Explained:**

- **config.pageBorder**: `"left:right:top:bottom"` - Border widths for page edges
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

**Response:**
- **Content-Type:** `application/pdf`
- **File:** `template-pdf-<timestamp>.pdf` (auto-download)

## 🧪 Usage Examples

### 📱 Healthcare Form with Font Styling (cURL)
```bash
curl -X POST "http://localhost:8080/api/v1/generate/template-pdf" \
  -H "Content-Type: application/json" \
  -d '{
    "config": {
      "pageBorder": "2:2:2:2"
    },
    "title": {
      "props": "font1:20:110:center:0:0:2:0",
      "text": "Patient Encounter Form"
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
          },
          {
            "row": [
              {
                "props": "font1:12:100:left:1:0:1:1",
                "text": "Gender: Male"
              },
              {
                "props": "font1:12:000:center:0:0:1:1",
                "chequebox": true
              },
              {
                "props": "font1:12:100:left:0:0:1:1",
                "text": "Female"
              },
              {
                "props": "font1:12:000:center:0:1:1:1",
                "chequebox": false
              }
            ]
          }
        ]
      }
    ],
    "footer": {
      "font": "font1:10:001:center",
      "text": "Confidential Medical Document"
    }
  }' \
  --output patient-form-styled.pdf
```

### 🪟 Windows CMD Example
```cmd
curl -X POST "http://localhost:8080/api/v1/generate/template-pdf" ^
  -H "Content-Type: application/json" ^
  -d "{\"config\":{\"pageBorder\":\"1:1:1:1\"},\"title\":{\"props\":\"font1:16:center:0:0:1:0\",\"text\":\"Invoice Template\"},\"table\":[{\"maxcolumns\":2,\"rows\":[{\"row\":[{\"props\":\"font1:12:left:1:1:1:1\",\"text\":\"Item:\"},{\"props\":\"font1:12:right:1:1:1:1\",\"text\":\"$100.00\"}]}]}],\"footer\":{\"font\":\"font1:10:center\",\"text\":\"Thank you for your business\"}}" ^
  --output invoice.pdf
```

### 🐍 Python Example with Font Styling
```python
import requests
import json

url = "http://localhost:8080/api/v1/generate/template-pdf"
template = {
    "config": {
        "pageBorder": "1:1:1:1"
    },
    "title": {
        "props": "font1:22:111:center:0:0:2:0",  # Bold + Italic + Underlined title
        "text": "Survey Form"
    },
    "table": [
        {
            "maxcolumns": 3,
            "rows": [
                {
                    "row": [
                        {
                            "props": "font1:14:100:left:1:1:1:1",  # Bold question
                            "text": "Question 1: Are you satisfied?"
                        },
                        {
                            "props": "font1:12:000:center:1:1:1:1",
                            "chequebox": True
                        },
                        {
                            "props": "font1:12:010:left:1:1:1:1",  # Italic answer
                            "text": "Yes"
                        }
                    ]
                }
            ]
        }
    ],
    "footer": {
        "font": "font1:10:001:right",  # Underlined footer
        "text": "Page 1 of 1"
    }
}

response = requests.post(url, json=template)
with open("survey-styled.pdf", "wb") as f:
    f.write(response.content)
```

## ✨ Features

- 🎯 **Template-based**: JSON-driven PDF generation
- 📋 **Tables & Forms**: Support for complex table layouts
- ☑️ **Checkboxes**: Interactive checkbox elements
- 🎨 **Font Styling**: Bold, italic, and underline text support
- 🔤 **Flexible Typography**: Custom fonts, sizes, and alignments
- 🔲 **Border Control**: Granular border configuration
- ⚡ **Fast**: In-memory PDF generation
- 📦 **Self-contained**: Single binary deployment
- 🌐 **Cross-platform**: Runs on Windows, Linux, macOS

## 🗺️ Roadmap & TODO

- [ ] 🧪 Add comprehensive unit tests
- [ ] 🎨 Support for colors and advanced styling
- [ ] 📊 Image embedding support
- [ ] 🐳 Docker containerization
- [ ] 📈 Metrics and health check endpoints
- [ ] 🔐 Authentication and rate limiting
- [ ] 📋 Multi-page document support
- [ ] 💾 Template storage and management
- [ ] 📧 Email delivery integration

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
