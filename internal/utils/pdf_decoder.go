package utils

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// ExtractPDFFormFields extracts form field names and their types from PDF bytes
// by decompressing FlateDecode streams and parsing /T entries
func ExtractPDFFormFields(pdfBytes []byte) (map[string]string, error) {
	if len(pdfBytes) == 0 {
		return nil, errors.New("empty PDF bytes")
	}

	fields := make(map[string]string)

	// First, extract fields from uncompressed portions
	extractFieldsFromBytes(pdfBytes, fields)

	// Then, find and decompress all FlateDecode streams
	streamPattern := regexp.MustCompile(`(?s)stream\r?\n(.*?)endstream`)
	streamMatches := streamPattern.FindAllSubmatch(pdfBytes, -1)

	for _, match := range streamMatches {
		if len(match) < 2 {
			continue
		}

		rawStream := match[1]

		// Try to decompress with zlib (FlateDecode)
		decompressed, err := decompressFlateDecode(rawStream)
		if err != nil {
			continue // Skip streams that can't be decompressed
		}

		// Extract fields from decompressed data
		extractFieldsFromBytes(decompressed, fields)
	}

	return fields, nil
}

// decompressFlateDecode attempts to decompress FlateDecode stream data
func decompressFlateDecode(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// extractFieldsFromBytes extracts /T field entries from raw bytes
func extractFieldsFromBytes(data []byte, fields map[string]string) {
	// Pattern for /T (field_name) entries
	fieldPattern := regexp.MustCompile(`/T\s*\(([^)]*)\)`)
	fieldMatches := fieldPattern.FindAllSubmatch(data, -1)

	for _, match := range fieldMatches {
		if len(match) < 2 {
			continue
		}
		fieldName := string(match[1])
		if fieldName != "" {
			// For now, we'll set empty values - these will be filled by XFDF or form data
			fields[fieldName] = ""
		}
	}

	// Pattern for /T <hex_encoded_field_name> entries
	hexFieldPattern := regexp.MustCompile(`/T\s*<([^>]*)>`)
	hexMatches := hexFieldPattern.FindAllSubmatch(data, -1)

	for _, match := range hexMatches {
		if len(match) < 2 {
			continue
		}
		hexData := string(match[1])
		if fieldName := decodeHexString(hexData); fieldName != "" {
			fields[fieldName] = ""
		}
	}
}

// decodeHexString decodes a hex-encoded string to regular text
func decodeHexString(hexStr string) string {
	// Remove whitespace
	hexStr = strings.ReplaceAll(hexStr, " ", "")
	hexStr = strings.ReplaceAll(hexStr, "\n", "")
	hexStr = strings.ReplaceAll(hexStr, "\r", "")

	if len(hexStr)%2 != 0 {
		return "" // Invalid hex string
	}

	var result strings.Builder
	for i := 0; i < len(hexStr); i += 2 {
		if i+1 >= len(hexStr) {
			break
		}

		hexByte := hexStr[i : i+2]
		var b byte
		if _, err := fmt.Sscanf(hexByte, "%02x", &b); err != nil {
			continue
		}

		// Only include printable ASCII characters
		if b >= 32 && b <= 126 {
			result.WriteByte(b)
		}
	}

	return result.String()
}

// MergeFormFields merges PDF-extracted fields with XFDF/form data
// PDF fields provide the structure, XFDF/form data provides the values
func MergeFormFields(pdfFields, formData map[string]string) map[string]string {
	merged := make(map[string]string)

	// Start with PDF fields (empty values)
	for name := range pdfFields {
		merged[name] = ""
	}

	// Override with form data values
	for name, value := range formData {
		merged[name] = value
	}

	return merged
}
