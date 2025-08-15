# 📄 GoPdfSuit

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Gin Framework](https://img.shields.io/badge/Gin-Web%20Framework-00ADD8?style=flat)](https://gin-gonic.com/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

> 🚀 A lightweight Go web service that generates PDF invoices on-the-fly and returns them as downloadable attachments.

## 📖 Overview

GoPdfSuit is a simple yet powerful web service built with Go and the Gin framework. It features a custom in-memory PDF generator that creates minimal invoice PDFs without external dependencies, making it perfect for demonstration, testing, and lightweight production use.

## 🏗️ Project Structure

```
GoPdfSuit/
├── 📁 cmd/
│   └── 📁 gopdfsuit/           # 🎯 Application entrypoint
│       └── main.go
├── 📁 internal/
│   ├── 📁 handlers/            # 🔗 HTTP handlers and route registration
│   │   └── handlers.go
│   ├── 📁 models/              # 📊 Request/response data models
│   │   └── models.go
│   └── 📁 pdf/                 # 📄 PDF generation logic
│       └── pdf.go
├── 📄 go.mod                   # 📦 Go modules file
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

### Generate PDF Invoice

**Endpoint:** `POST /api/v1/generate/pdf`

**Headers:**
- `Content-Type: application/json`

**Request Body:**
```json
{
  "customer_name": "Acme Corporation",
  "items": [
    {
      "name": "Premium Widget",
      "quantity": 2,
      "price": 29.99
    },
    {
      "name": "Professional Service",
      "quantity": 1,
      "price": 99.95
    }
  ],
  "total": 159.93
}
```

**Response:**
- **Content-Type:** `application/pdf`
- **File:** `invoice-<timestamp>.pdf` (auto-download)

**Error Response:**
```json
{
  "error": "Invalid request data: <details>"
}
```

## 🧪 Usage Examples

### 📱 Using cURL (Unix/Linux/macOS)
```bash
curl -X POST "http://localhost:8080/api/v1/generate/pdf" \
  -H "Content-Type: application/json" \
  -d '{
    "customer_name": "Tech Solutions Inc",
    "items": [
      {"name": "Cloud Hosting", "quantity": 1, "price": 49.99},
      {"name": "SSL Certificate", "quantity": 2, "price": 15.00}
    ],
    "total": 79.99
  }' \
  --output invoice.pdf
```

### 🪟 Using cURL (Windows CMD)
```cmd
curl -X POST "http://localhost:8080/api/v1/generate/pdf" ^
  -H "Content-Type: application/json" ^
  -d "{\"customer_name\":\"Tech Solutions Inc\",\"items\":[{\"name\":\"Cloud Hosting\",\"quantity\":1,\"price\":49.99},{\"name\":\"SSL Certificate\",\"quantity\":2,\"price\":15.00}],\"total\":79.99}" ^
  --output invoice.pdf
```

### 🐍 Using Python
```python
import requests
import json

url = "http://localhost:8080/api/v1/generate/pdf"
data = {
    "customer_name": "Python Developers LLC",
    "items": [
        {"name": "API Development", "quantity": 1, "price": 500.00},
        {"name": "Testing Suite", "quantity": 1, "price": 200.00}
    ],
    "total": 700.00
}

response = requests.post(url, json=data)
with open("invoice.pdf", "wb") as f:
    f.write(response.content)
```

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

# Run tests with verbose output
go test -v ./...
```

## ✨ Features

- 🎯 **Lightweight**: No external PDF libraries required
- ⚡ **Fast**: In-memory PDF generation
- 🔧 **Simple API**: RESTful JSON interface
- 📦 **Self-contained**: Single binary deployment
- 🌐 **Cross-platform**: Runs on Windows, Linux, macOS
- 🛡️ **Type-safe**: Full Go type checking

## 🗺️ Roadmap & TODO

- [ ] 🧪 Add comprehensive unit tests
- [ ] 📊 Enhanced PDF layouts with tables and styling
- [ ] 🐳 Docker containerization
- [ ] 📈 Metrics and health check endpoints
- [ ] 🔐 Authentication and rate limiting
- [ ] 📋 Template-based PDF generation
- [ ] 🎨 Custom branding and themes
- [ ] 💾 Persistent invoice storage
- [ ] 📧 Email delivery integration

## ⚠️ Production Notes

> **⚠️ Important:** The current PDF generator is intentionally minimal and designed for demonstration purposes.

For production environments, consider:
- Using established PDF libraries like [gofpdf](https://github.com/jung-kurt/gofpdf) or [chromedp](https://github.com/chromedp/chromedp)
- Implementing streaming for large documents
- Adding comprehensive error handling and logging
- Setting up proper monitoring and metrics
- Implementing request validation and sanitization

## 🤝 Contributing

1. 🍴 Fork the repository
2. 🌟 Create a feature branch (`git checkout -b feature/amazing-feature`)
3. 💫 Commit your changes (`git commit -m 'Add amazing feature'`)
4. 📤 Push to the branch (`git push origin feature/amazing-feature`)
5. 🎉 Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- 🌐 [Gin Web Framework](https://gin-gonic.com/) for the excellent HTTP router
- 📖 [PDF Reference Manual](https://www.adobe.com/content/dam/acom/en/devnet/pdf/pdfs/pdf_reference_archives/PDFReference.pdf) for PDF format specifications
- 🚀 Go community for continuous inspiration

---

<div align="center">
  <p>Made with ❤️ and ☕ by <a href="https://github.com/chinmay-sawant">Chinmay Sawant</a></p>
  <p>⭐ Star this repo if you find it helpful!</p>
</div>
