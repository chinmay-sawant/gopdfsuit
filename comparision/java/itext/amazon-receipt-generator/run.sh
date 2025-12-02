#!/bin/bash

# Amazon Receipt Generator - Run Script
# This script downloads dependencies and runs the application

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LIB_DIR="$SCRIPT_DIR/lib"
SRC_DIR="$SCRIPT_DIR/src/main/java"
BUILD_DIR="$SCRIPT_DIR/build/classes"
MAIN_CLASS="com.amazonreceipt.AmazonReceiptGenerator"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║     Amazon Receipt PDF Generator - Build & Run Script         ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# Check Java
if ! command -v java &> /dev/null; then
    echo -e "${RED}Error: Java is not installed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Java found: $(java -version 2>&1 | head -n 1)${NC}"

# Check if dependencies need to be downloaded
if [ ! -d "$LIB_DIR" ] || [ ! "$(ls -A $LIB_DIR 2>/dev/null)" ]; then
    echo -e "${YELLOW}Downloading dependencies...${NC}"
    mkdir -p "$LIB_DIR"
    
    # Define Maven Central base URL
    MAVEN_URL="https://repo1.maven.org/maven2"
    
    # iText dependencies
    DEPS=(
        "com/itextpdf/kernel/8.0.4/kernel-8.0.4.jar"
        "com/itextpdf/io/8.0.4/io-8.0.4.jar"
        "com/itextpdf/layout/8.0.4/layout-8.0.4.jar"
        "com/itextpdf/commons/8.0.4/commons-8.0.4.jar"
        "com/itextpdf/bouncy-castle-adapter/8.0.4/bouncy-castle-adapter-8.0.4.jar"
        "com/itextpdf/bouncy-castle-fips-adapter/8.0.4/bouncy-castle-fips-adapter-8.0.4.jar"
        "com/google/code/gson/gson/2.10.1/gson-2.10.1.jar"
        "org/slf4j/slf4j-api/2.0.9/slf4j-api-2.0.9.jar"
        "org/slf4j/slf4j-simple/2.0.9/slf4j-simple-2.0.9.jar"
        "org/bouncycastle/bcprov-jdk18on/1.76/bcprov-jdk18on-1.76.jar"
        "org/bouncycastle/bcpkix-jdk18on/1.76/bcpkix-jdk18on-1.76.jar"
        "org/bouncycastle/bcutil-jdk18on/1.76/bcutil-jdk18on-1.76.jar"
    )
    
    for dep in "${DEPS[@]}"; do
        filename=$(basename "$dep")
        echo "  Downloading $filename..."
        curl -sL "$MAVEN_URL/$dep" -o "$LIB_DIR/$filename" || echo "  Warning: Failed to download $filename"
    done
    
    echo -e "${GREEN}✓ Dependencies downloaded${NC}"
else
    echo -e "${GREEN}✓ Dependencies already exist${NC}"
fi

# Create build directory
mkdir -p "$BUILD_DIR"

# Compile Java sources
echo -e "${YELLOW}Compiling Java sources...${NC}"

# Build classpath
CLASSPATH=""
for jar in "$LIB_DIR"/*.jar; do
    if [ -f "$jar" ]; then
        CLASSPATH="$CLASSPATH:$jar"
    fi
done
CLASSPATH="${CLASSPATH:1}" # Remove leading colon

# Find all Java files
JAVA_FILES=$(find "$SRC_DIR" -name "*.java")

# Compile
javac -cp "$CLASSPATH" -d "$BUILD_DIR" $JAVA_FILES

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Compilation successful${NC}"
else
    echo -e "${RED}✗ Compilation failed${NC}"
    exit 1
fi

# Run the application
echo -e "${YELLOW}Running application...${NC}"
echo ""

java -cp "$BUILD_DIR:$CLASSPATH" "$MAIN_CLASS" "$@"
