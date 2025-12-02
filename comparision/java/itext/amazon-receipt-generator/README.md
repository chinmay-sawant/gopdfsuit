# 🛒 Amazon Receipt PDF Generator

A beautiful, professional PDF receipt generator built with **iText 8** - the latest version of the industry-leading PDF library for Java.

![Java](https://img.shields.io/badge/Java-11+-orange?style=flat-square&logo=java)
![iText](https://img.shields.io/badge/iText-8.0.4-blue?style=flat-square)
![Maven](https://img.shields.io/badge/Maven-3.6+-red?style=flat-square&logo=apache-maven)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

## ✨ Features

- 📄 **Beautiful PDF Generation** - Professional Amazon-style receipts
- 🎨 **Custom Color Palette** - Amazon-inspired orange and dark theme
- 📊 **Dynamic Tables** - Support for multi-column layouts with flexible widths
- 🔤 **Rich Text Formatting** - Bold, italic, underline support
- ⏱️ **Performance Timing** - Built-in start/end timers with detailed metrics
- 📝 **JSON Configuration** - Easy to customize receipt content

## 📋 Prerequisites

- **Java 11** or higher

### Verify Installation

```bash
# Check Java version
java -version
```

## 🚀 Quick Start

### Option 1: Using the Run Script (No Maven/Gradle Required)

The easiest way to run the application without installing any build tools:

```bash
cd java/amazon-receipt-generator

# Run with default sample data
./run.sh

# Run with custom input/output
./run.sh path/to/input.json path/to/output.pdf
```

The script will automatically:
1. Download all required dependencies from Maven Central
2. Compile the Java source files
3. Run the application with timing output

### Option 2: Using Maven

```bash
cd java/amazon-receipt-generator

# Clean and build
mvn clean package

# Run the application
mvn exec:java

# Or run the JAR
java -jar target/amazon-receipt-generator-1.0.0.jar
```

### Option 3: Using Gradle

```bash
cd java/amazon-receipt-generator

# Build and run
./gradlew run

# Or build the JAR
./gradlew build
java -jar build/libs/amazon-receipt-generator-1.0.0.jar
```

### Custom Input/Output Paths

```bash
# Using run script
./run.sh ../../sampledata/amazon/amazon_receipt.json my_receipt.pdf

# Using Maven
mvn exec:java -Dexec.args="path/to/input.json path/to/output.pdf"

# Using JAR
java -jar target/amazon-receipt-generator-1.0.0.jar input.json output.pdf
```

## 📁 Project Structure

```
amazon-receipt-generator/
├── pom.xml                          # Maven configuration
├── README.md                        # This file
└── src/
    └── main/
        └── java/
            └── com/
                └── amazonreceipt/
                    ├── AmazonReceiptGenerator.java    # Main entry point
                    ├── generator/
                    │   └── PdfReceiptGenerator.java   # PDF generation logic
                    ├── model/
                    │   ├── CellData.java              # Cell data model
                    │   ├── FooterSection.java         # Footer model
                    │   ├── ReceiptConfig.java         # Config model
                    │   ├── ReceiptData.java           # Root data model
                    │   ├── TableRow.java              # Table row model
                    │   ├── TableSection.java          # Table section model
                    │   └── TitleSection.java          # Title model
                    └── util/
                        ├── ColorPalette.java          # Color definitions
                        ├── FontManager.java           # Font management
                        └── TimerUtil.java             # Performance timing
```

## 📝 JSON Input Format

The application reads JSON files with the following structure:

```json
{
  "config": {
    "pageBorder": "1:1:1:1",
    "page": "A4",
    "pageAlignment": 1,
    "watermark": ""
  },
  "title": {
    "props": "font1:18:100:left:0:0:0:1",
    "text": "amazon"
  },
  "table": [
    {
      "maxcolumns": 2,
      "columnwidths": [1, 1],
      "rows": [
        {
          "row": [
            {
              "props": "font1:11:100:left:1:1:1:0",
              "text": "Order Receipt"
            }
          ]
        }
      ]
    }
  ],
  "footer": {
    "font": "font1:8:000:left",
    "text": "Thank you for shopping with Amazon!"
  }
}
```

### Props Format

The `props` field uses colon-separated values:

```
font:size:style:alignment:borderTop:borderRight:borderBottom:borderLeft
```

- **font**: Font identifier (e.g., `font1`)
- **size**: Font size in points (e.g., `11`)
- **style**: 3-digit style code
  - `100` = Bold
  - `010` = Italic
  - `001` = Underline
  - `110` = Bold + Italic
  - `000` = Normal
- **alignment**: `left`, `center`, `right`, `justified`
- **borders**: `1` = visible, `0` = hidden

## ⏱️ Timer Output

The application includes built-in performance timing:

```
╔══════════════════════════════════════════════════════════════╗
║  ⏱️  TIMER STARTED: PDF Generation (Total)                  ║
╚══════════════════════════════════════════════════════════════╝

╔══════════════════════════════════════════════════════════════╗
║  ⏱️  TIMER STOPPED: PDF Generation (Total)                  ║
╠══════════════════════════════════════════════════════════════╣
║  📊 Duration Results:                                        ║
║     • Nanoseconds:  45,234,567 ns                            ║
║     • Milliseconds: 45.235 ms                                ║
║     • Seconds:      0.045235 s                               ║
╚══════════════════════════════════════════════════════════════╝
```

## 🎨 Color Palette

The generator uses an Amazon-inspired color palette:

| Color | Hex | Usage |
|-------|-----|-------|
| Amazon Orange | `#FF9900` | Title, accents |
| Primary Dark | `#232F3E` | Headers |
| Light Gray | `#F8F9FA` | Background |
| Text Primary | `#0F1111` | Body text |
| Border Light | `#DDDDDD` | Table borders |

## 📦 Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| iText Core | 8.0.4 | PDF generation |
| Gson | 2.10.1 | JSON parsing |
| SLF4J | 2.0.9 | Logging |

## 🔧 Configuration

### Maven Properties

Edit `pom.xml` to change versions:

```xml
<properties>
    <maven.compiler.source>11</maven.compiler.source>
    <maven.compiler.target>11</maven.compiler.target>
    <itext.version>8.0.4</itext.version>
    <gson.version>2.10.1</gson.version>
</properties>
```

## 🐛 Troubleshooting

### Common Issues

1. **"Input file not found"**
   - Ensure the JSON file path is correct
   - Try using an absolute path

2. **"Java version error"**
   - Upgrade to Java 17 or higher
   - Set `JAVA_HOME` environment variable

3. **"Maven build failure"**
   - Run `mvn clean` before building
   - Check internet connection for downloading dependencies

### Debug Mode

```bash
# Run with debug output
mvn exec:java -X
```

## 📄 Sample Output

The generated PDF includes:
- Amazon-branded header with orange logo text
- Order information section
- Seller and shipping details
- Billing and shipping addresses
- Payment and shipping method details
- Itemized order list with quantities and prices
- Order summary with totals
- Customer information
- Terms and conditions
- Footer with timestamp

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📜 License

This project is licensed under the MIT License.

## 🙏 Acknowledgments

- [iText](https://itextpdf.com/) - PDF library
- [Google Gson](https://github.com/google/gson) - JSON parsing
- Amazon design system for inspiration

---

**Made with ❤️ and ☕**
