package pdf

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"math"
	"time"
)

// ConvertPDFDateToXMP converts a PDF date string (D:YYYYMMDDHHmmSSOHH'mm') to XMP format (YYYY-MM-DDTHH:mm:ss+HH:MM)
func ConvertPDFDateToXMP(pdfDate string) string {
	// PDF format: D:20060102150405-07'00' or D:20060102150405+05'30'
	// XMP format: 2006-01-02T15:04:05-07:00 or 2006-01-02T15:04:05+05:30
	if len(pdfDate) < 16 {
		return time.Now().Format("2006-01-02T15:04:05-07:00")
	}

	// Remove 'D:' prefix
	date := pdfDate[2:]

	year := date[0:4]
	month := date[4:6]
	day := date[6:8]
	hour := date[8:10]
	min := date[10:12]
	sec := date[12:14]

	// Parse timezone: -07'00' or +05'30'
	// Format is: sign + 2 digits + quote + 2 digits + quote
	tz := "+00:00"
	if len(date) >= 20 {
		tzSign := string(date[14])
		tzHour := date[15:17]
		// Skip the quote at position 17, get minutes at 18-19
		tzMin := date[18:20]
		tz = tzSign + tzHour + ":" + tzMin
	}

	return fmt.Sprintf("%s-%s-%sT%s:%s:%s%s", year, month, day, hour, min, sec, tz)
}

// GenerateXMPMetadata generates PDF/A-4 compliant XMP metadata (PDF 2.0 based)
// pdfDateStr should be in PDF format: D:YYYYMMDDHHmmSSOHH'mm'
func GenerateXMPMetadata(documentID string, pdfDateStr string) string {
	// Convert PDF date to XMP date format for consistency
	xmpDateStr := ConvertPDFDateToXMP(pdfDateStr)

	// PDF/A-4 is the PDF/A standard based on PDF 2.0 (ISO 32000-2)
	// pdfaid:part=4, no conformance level needed for PDF/A-4
	xmp := `<?xpacket begin="` + "\xef\xbb\xbf" + `" id="W5M0MpCehiHzreSzNTczkc9d"?>
<x:xmpmeta xmlns:x="adobe:ns:meta/">
  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">
    <rdf:Description rdf:about=""
        xmlns:dc="http://purl.org/dc/elements/1.1/"
        xmlns:xmp="http://ns.adobe.com/xap/1.0/"
        xmlns:pdf="http://ns.adobe.com/pdf/1.3/"
        xmlns:pdfaid="http://www.aiim.org/pdfa/ns/id/"
        xmlns:pdfuaid="http://www.aiim.org/pdfua/ns/id/">
      <dc:format>application/pdf</dc:format>
      <dc:creator>
        <rdf:Seq>
          <rdf:li>GoPDFSuit</rdf:li>
        </rdf:Seq>
      </dc:creator>
      <dc:title>
        <rdf:Alt>
          <rdf:li xml:lang="x-default">PDF Document</rdf:li>
        </rdf:Alt>
      </dc:title>
      <xmp:CreatorTool>GoPDFSuit PDF Generator</xmp:CreatorTool>
      <xmp:CreateDate>` + xmpDateStr + `</xmp:CreateDate>
      <xmp:ModifyDate>` + xmpDateStr + `</xmp:ModifyDate>
      <pdf:Producer>GoPDFSuit</pdf:Producer>
      <pdfaid:part>4</pdfaid:part>
      <pdfaid:rev>2020</pdfaid:rev>
    </rdf:Description>
  </rdf:RDF>
</x:xmpmeta>
<?xpacket end="w"?>`

	return xmp
}

// GenerateXMPMetadataObject generates the XMP metadata stream object
// pdfDateStr should be in PDF format: D:YYYYMMDDHHmmSSOHH'mm'
func GenerateXMPMetadataObject(objectID int, documentID string, pdfDateStr string) string {
	xmpData := GenerateXMPMetadata(documentID, pdfDateStr)

	return fmt.Sprintf("%d 0 obj\n<< /Type /Metadata /Subtype /XML /Length %d >>\nstream\n%s\nendstream\nendobj\n",
		objectID, len(xmpData), xmpData)
}

// buildSRGBICCProfile builds a valid sRGB ICC profile from scratch
// This creates a properly structured ICC v2.1 profile (more widely compatible)
// We use the INVERSE sRGB gamma (linearization curve) as the TRC because:
// 1. Hex colors like #154360 are already gamma-encoded (sRGB)
// 2. When Adobe applies this profile, it applies the TRC to "linearize" the input
// 3. If we don't compensate, colors appear washed out due to matrix transformation
// 4. By using the linearization curve, we tell Adobe to treat these as already-linear
//    which means the matrix conversion produces correct output
func buildSRGBICCProfile() []byte {
	// Use inverse sRGB gamma curve (linearization) to compensate for matrix conversion
	// This curve converts sRGB-encoded values to linear, which is what Adobe expects
	gammaTable := make([]uint16, 1024)
	for i := 0; i < 1024; i++ {
		x := float64(i) / 1023.0
		var y float64
		// sRGB to linear (inverse gamma)
		if x <= 0.04045 {
			y = x / 12.92
		} else {
			y = math.Pow((x+0.055)/1.055, 2.4)
		}
		gammaTable[i] = uint16(y * 65535.0)
	}

	// Calculate sizes - full 1024-entry curve
	headerSize := 128
	tagTableSize := 4 + 9*12 // count + 9 tags * 12 bytes each
	cprtSize := 32           // text type: 4 + 4 + 24 (text)
	descSize := 92           // mluc type
	wtptSize := 20           // XYZ type
	xyzSize := 20            // XYZ type (for rXYZ, gXYZ, bXYZ)
	curvSize := 12 + 2048    // curv header (12 bytes) + 1024 * 2 bytes

	// Calculate offsets (must be 4-byte aligned)
	cprtOffset := headerSize + tagTableSize
	descOffset := cprtOffset + cprtSize
	wtptOffset := descOffset + descSize
	rXYZOffset := wtptOffset + wtptSize
	gXYZOffset := rXYZOffset + xyzSize
	bXYZOffset := gXYZOffset + xyzSize
	curvOffset := bXYZOffset + xyzSize

	profileSize := curvOffset + curvSize

	profile := make([]byte, profileSize)

	// Write header (128 bytes)
	binary.BigEndian.PutUint32(profile[0:4], uint32(profileSize)) // Profile size
	copy(profile[4:8], []byte{0, 0, 0, 0})                        // CMM Type
	binary.BigEndian.PutUint32(profile[8:12], 0x02100000)         // Version 2.1 (more compatible)
	copy(profile[12:16], []byte("mntr"))                          // Device class: monitor
	copy(profile[16:20], []byte("RGB "))                          // Color space: RGB
	copy(profile[20:24], []byte("XYZ "))                          // PCS: XYZ
	// Date/time: 2024-01-01 00:00:00
	binary.BigEndian.PutUint16(profile[24:26], 2024) // Year
	binary.BigEndian.PutUint16(profile[26:28], 1)    // Month
	binary.BigEndian.PutUint16(profile[28:30], 1)    // Day
	binary.BigEndian.PutUint16(profile[30:32], 0)    // Hour
	binary.BigEndian.PutUint16(profile[32:34], 0)    // Minute
	binary.BigEndian.PutUint16(profile[34:36], 0)    // Second
	copy(profile[36:40], []byte("acsp"))             // Signature
	copy(profile[40:44], []byte{0, 0, 0, 0})         // Platform
	binary.BigEndian.PutUint32(profile[44:48], 0)    // Flags
	binary.BigEndian.PutUint32(profile[48:52], 0)    // Device manufacturer
	binary.BigEndian.PutUint32(profile[52:56], 0)    // Device model
	binary.BigEndian.PutUint64(profile[56:64], 0)    // Device attributes
	binary.BigEndian.PutUint32(profile[64:68], 0)    // Rendering intent (perceptual)
	// PCS illuminant (D50): X=0.9642, Y=1.0, Z=0.8249
	binary.BigEndian.PutUint32(profile[68:72], 0x0000F6D6) // X
	binary.BigEndian.PutUint32(profile[72:76], 0x00010000) // Y
	binary.BigEndian.PutUint32(profile[76:80], 0x0000D32D) // Z
	binary.BigEndian.PutUint32(profile[80:84], 0)          // Profile creator
	copy(profile[84:100], make([]byte, 16))                // Profile ID (zeros for now)
	copy(profile[100:128], make([]byte, 28))               // Reserved

	// Write tag table
	offset := headerSize
	binary.BigEndian.PutUint32(profile[offset:offset+4], 9) // Tag count
	offset += 4

	// Tag entries
	tags := []struct {
		sig    string
		offset int
		size   int
	}{
		{"cprt", cprtOffset, cprtSize},
		{"desc", descOffset, descSize},
		{"wtpt", wtptOffset, wtptSize},
		{"rXYZ", rXYZOffset, xyzSize},
		{"gXYZ", gXYZOffset, xyzSize},
		{"bXYZ", bXYZOffset, xyzSize},
		{"rTRC", curvOffset, curvSize},
		{"gTRC", curvOffset, curvSize}, // Share with rTRC
		{"bTRC", curvOffset, curvSize}, // Share with rTRC
	}

	for _, tag := range tags {
		copy(profile[offset:offset+4], []byte(tag.sig))
		binary.BigEndian.PutUint32(profile[offset+4:offset+8], uint32(tag.offset))
		binary.BigEndian.PutUint32(profile[offset+8:offset+12], uint32(tag.size))
		offset += 12
	}

	// Write cprt (copyright) - textType
	offset = cprtOffset
	copy(profile[offset:offset+4], []byte("text"))
	binary.BigEndian.PutUint32(profile[offset+4:offset+8], 0)
	copy(profile[offset+8:offset+32], []byte("Public Domain\x00"))

	// Write desc (description) - mluc type
	offset = descOffset
	copy(profile[offset:offset+4], []byte("mluc"))
	binary.BigEndian.PutUint32(profile[offset+4:offset+8], 0)    // Reserved
	binary.BigEndian.PutUint32(profile[offset+8:offset+12], 1)   // Record count
	binary.BigEndian.PutUint32(profile[offset+12:offset+16], 12) // Record size
	copy(profile[offset+16:offset+18], []byte("en"))             // Language
	copy(profile[offset+18:offset+20], []byte("US"))             // Country
	binary.BigEndian.PutUint32(profile[offset+20:offset+24], 34) // String length (bytes)
	binary.BigEndian.PutUint32(profile[offset+24:offset+28], 28) // String offset
	descText := []uint16{'s', 'R', 'G', 'B', ' ', 'I', 'E', 'C', '6', '1', '9', '6', '6', '-', '2', '.', '1'}
	for i, c := range descText {
		binary.BigEndian.PutUint16(profile[offset+28+i*2:offset+30+i*2], c)
	}

	// Write wtpt (white point) - XYZType (D50)
	offset = wtptOffset
	copy(profile[offset:offset+4], []byte("XYZ "))
	binary.BigEndian.PutUint32(profile[offset+4:offset+8], 0)
	binary.BigEndian.PutUint32(profile[offset+8:offset+12], 0x0000F6D6)  // X: 0.9642
	binary.BigEndian.PutUint32(profile[offset+12:offset+16], 0x00010000) // Y: 1.0
	binary.BigEndian.PutUint32(profile[offset+16:offset+20], 0x0000D32D) // Z: 0.8249

	// Write rXYZ (red primary) - XYZType
	offset = rXYZOffset
	copy(profile[offset:offset+4], []byte("XYZ "))
	binary.BigEndian.PutUint32(profile[offset+4:offset+8], 0)
	binary.BigEndian.PutUint32(profile[offset+8:offset+12], 0x00006FA2)  // X: 0.4361
	binary.BigEndian.PutUint32(profile[offset+12:offset+16], 0x000038F5) // Y: 0.2225
	binary.BigEndian.PutUint32(profile[offset+16:offset+20], 0x00000390) // Z: 0.0139

	// Write gXYZ (green primary) - XYZType
	offset = gXYZOffset
	copy(profile[offset:offset+4], []byte("XYZ "))
	binary.BigEndian.PutUint32(profile[offset+4:offset+8], 0)
	binary.BigEndian.PutUint32(profile[offset+8:offset+12], 0x00006299)  // X: 0.3851
	binary.BigEndian.PutUint32(profile[offset+12:offset+16], 0x0000B785) // Y: 0.7169
	binary.BigEndian.PutUint32(profile[offset+16:offset+20], 0x000018DA) // Z: 0.0971

	// Write bXYZ (blue primary) - XYZType
	offset = bXYZOffset
	copy(profile[offset:offset+4], []byte("XYZ "))
	binary.BigEndian.PutUint32(profile[offset+4:offset+8], 0)
	binary.BigEndian.PutUint32(profile[offset+8:offset+12], 0x000024A0)  // X: 0.1431
	binary.BigEndian.PutUint32(profile[offset+12:offset+16], 0x00000F84) // Y: 0.0606
	binary.BigEndian.PutUint32(profile[offset+16:offset+20], 0x0000B6CF) // Z: 0.7139

	// Write TRC (tone reproduction curve) - curvType with inverse sRGB gamma
	// This linearization curve tells Adobe our input values are sRGB-encoded
	offset = curvOffset
	copy(profile[offset:offset+4], []byte("curv"))
	binary.BigEndian.PutUint32(profile[offset+4:offset+8], 0)    // Reserved
	binary.BigEndian.PutUint32(profile[offset+8:offset+12], 1024) // Entry count
	for i, v := range gammaTable {
		binary.BigEndian.PutUint16(profile[offset+12+i*2:offset+14+i*2], v)
	}

	return profile
}

// GetSRGBICCProfile returns the complete sRGB ICC profile
func GetSRGBICCProfile() []byte {
	return buildSRGBICCProfile()
}

// GenerateICCProfileObject generates the ICC profile stream object for sRGB
// Returns the bytes to write to the PDF buffer
func GenerateICCProfileObject(objectID int) []byte {
	// Get the complete sRGB ICC profile
	iccProfile := GetSRGBICCProfile()

	// Compress the ICC profile
	var compressedBuf bytes.Buffer
	zlibWriter := zlib.NewWriter(&compressedBuf)
	zlibWriter.Write(iccProfile)
	zlibWriter.Close()
	compressedData := compressedBuf.Bytes()

	// Build the object with proper binary stream handling
	var result bytes.Buffer
	result.WriteString(fmt.Sprintf("%d 0 obj\n<< /Filter /FlateDecode /Length %d /N 3 /Alternate /DeviceRGB >>\nstream\n",
		objectID, len(compressedData)))
	result.Write(compressedData)
	result.WriteString("\nendstream\nendobj\n")

	return result.Bytes()
}

// GenerateGrayICCProfileObject generates the ICC profile stream object for DeviceGray
// Returns the bytes to write to the PDF buffer
func GenerateGrayICCProfileObject(objectID int) []byte {
	// Build a simple Gray ICC profile
	grayProfile := buildGrayICCProfile()

	// Compress the ICC profile
	var compressedBuf bytes.Buffer
	zlibWriter := zlib.NewWriter(&compressedBuf)
	zlibWriter.Write(grayProfile)
	zlibWriter.Close()
	compressedData := compressedBuf.Bytes()

	// Build the object
	var result bytes.Buffer
	result.WriteString(fmt.Sprintf("%d 0 obj\n<< /Filter /FlateDecode /Length %d /N 1 /Alternate /DeviceGray >>\nstream\n",
		objectID, len(compressedData)))
	result.Write(compressedData)
	result.WriteString("\nendstream\nendobj\n")

	return result.Bytes()
}

// buildGrayICCProfile builds a valid Gray ICC profile
func buildGrayICCProfile() []byte {
	// Pre-calculate gamma table (same sRGB gamma for gray)
	gammaTable := make([]uint16, 1024)
	for i := 0; i < 1024; i++ {
		x := float64(i) / 1023.0
		var y float64
		if x <= 0.0031308 {
			y = x * 12.92
		} else {
			y = 1.055*math.Pow(x, 1.0/2.4) - 0.055
		}
		gammaTable[i] = uint16(y * 65535.0)
	}

	headerSize := 128
	tagTableSize := 4 + 4*12 // count + 4 tags
	cprtSize := 32
	descSize := 92
	wtptSize := 20
	curvSize := 12 + 2048

	cprtOffset := headerSize + tagTableSize
	descOffset := cprtOffset + cprtSize
	wtptOffset := descOffset + descSize
	curvOffset := wtptOffset + wtptSize

	profileSize := curvOffset + curvSize
	profile := make([]byte, profileSize)

	// Header
	binary.BigEndian.PutUint32(profile[0:4], uint32(profileSize))
	binary.BigEndian.PutUint32(profile[8:12], 0x02100000) // Version 2.1 (more compatible)
	copy(profile[12:16], []byte("mntr"))                  // monitor
	copy(profile[16:20], []byte("GRAY"))                  // Gray color space
	copy(profile[20:24], []byte("XYZ "))                  // PCS
	binary.BigEndian.PutUint16(profile[24:26], 2024)
	binary.BigEndian.PutUint16(profile[26:28], 1)
	binary.BigEndian.PutUint16(profile[28:30], 1)
	copy(profile[36:40], []byte("acsp"))
	binary.BigEndian.PutUint32(profile[68:72], 0x0000F6D6)
	binary.BigEndian.PutUint32(profile[72:76], 0x00010000)
	binary.BigEndian.PutUint32(profile[76:80], 0x0000D32D)

	// Tag table
	offset := headerSize
	binary.BigEndian.PutUint32(profile[offset:offset+4], 4) // 4 tags
	offset += 4

	tags := []struct {
		sig    string
		offset int
		size   int
	}{
		{"cprt", cprtOffset, cprtSize},
		{"desc", descOffset, descSize},
		{"wtpt", wtptOffset, wtptSize},
		{"kTRC", curvOffset, curvSize},
	}

	for _, tag := range tags {
		copy(profile[offset:offset+4], []byte(tag.sig))
		binary.BigEndian.PutUint32(profile[offset+4:offset+8], uint32(tag.offset))
		binary.BigEndian.PutUint32(profile[offset+8:offset+12], uint32(tag.size))
		offset += 12
	}

	// cprt
	offset = cprtOffset
	copy(profile[offset:offset+4], []byte("text"))
	copy(profile[offset+8:offset+32], []byte("Public Domain\x00"))

	// desc
	offset = descOffset
	copy(profile[offset:offset+4], []byte("mluc"))
	binary.BigEndian.PutUint32(profile[offset+8:offset+12], 1)
	binary.BigEndian.PutUint32(profile[offset+12:offset+16], 12)
	copy(profile[offset+16:offset+18], []byte("en"))
	copy(profile[offset+18:offset+20], []byte("US"))
	binary.BigEndian.PutUint32(profile[offset+20:offset+24], 20)
	binary.BigEndian.PutUint32(profile[offset+24:offset+28], 28)
	descText := []uint16{'s', 'R', 'G', 'B', ' ', 'G', 'r', 'a', 'y'}
	for i, c := range descText {
		binary.BigEndian.PutUint16(profile[offset+28+i*2:offset+30+i*2], c)
	}

	// wtpt
	offset = wtptOffset
	copy(profile[offset:offset+4], []byte("XYZ "))
	binary.BigEndian.PutUint32(profile[offset+8:offset+12], 0x0000F6D6)
	binary.BigEndian.PutUint32(profile[offset+12:offset+16], 0x00010000)
	binary.BigEndian.PutUint32(profile[offset+16:offset+20], 0x0000D32D)

	// kTRC (gray curve)
	offset = curvOffset
	copy(profile[offset:offset+4], []byte("curv"))
	binary.BigEndian.PutUint32(profile[offset+8:offset+12], 1024)
	for i, v := range gammaTable {
		binary.BigEndian.PutUint16(profile[offset+12+i*2:offset+14+i*2], v)
	}

	return profile
}

// GenerateOutputIntentObject generates the OutputIntent object for PDF/A-4
func GenerateOutputIntentObject(objectID int, iccProfileID int) string {
	// For PDF/A-4 (PDF 2.0), use GTS_PDFA1 subtype (still valid)
	return fmt.Sprintf("%d 0 obj\n<< /Type /OutputIntent /S /GTS_PDFA1 /OutputConditionIdentifier (sRGB IEC61966-2.1) /RegistryName (http://www.color.org) /Info (sRGB IEC61966-2.1) /DestOutputProfile %d 0 R >>\nendobj\n",
		objectID, iccProfileID)
}
