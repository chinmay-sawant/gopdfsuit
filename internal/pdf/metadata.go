package pdf

import (
	"bytes"
	"compress/zlib"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// PDFAHandler handles PDF/A compliance features
type PDFAHandler struct {
	config            *models.PDFAConfig
	pageManager       *PageManager
	metadataObjID     int
	outputIntentObjID int
	iccProfileObjID   int
	encryptor         ObjectEncryptor
}

// NewPDFAHandler creates a new PDF/A handler
func NewPDFAHandler(config *models.PDFAConfig, pm *PageManager, encryptor ObjectEncryptor) *PDFAHandler {
	return &PDFAHandler{
		config:      config,
		pageManager: pm,
		encryptor:   encryptor,
	}
}

// GetConformanceLevel returns the PDF/A part and conformance level
func (h *PDFAHandler) GetConformanceLevel() (part int, conformance string) {
	switch h.config.Conformance {
	case "1b":
		return 1, "B"
	case "1a":
		return 1, "A"
	case "2b":
		return 2, "B"
	case "2a":
		return 2, "A"
	case "2u":
		return 2, "U"
	case "3b":
		return 3, "B"
	case "3a":
		return 3, "A"
	case "3u":
		return 3, "U"
	case "4":
		return 4, ""
	case "4f":
		return 4, "F"
	case "4e":
		return 4, "E"
	default:
		return 4, "" // Default to PDF/A-4
	}
}

// GenerateXMPMetadata generates the XMP metadata stream for PDF/A
func (h *PDFAHandler) GenerateXMPMetadata(documentID string) (int, string) {
	part, conformance := h.GetConformanceLevel()

	// Get current time in ISO 8601 format
	now := time.Now().UTC()
	createDate := now.Format("2006-01-02T15:04:05Z")
	modifyDate := createDate

	// Build XMP packet
	var xmp strings.Builder
	xmp.WriteString(`<?xpacket begin="` + "\xef\xbb\xbf" + `" id="W5M0MpCehiHzreSzNTczkc9d"?>`)
	xmp.WriteString("\n")
	xmp.WriteString(`<x:xmpmeta xmlns:x="adobe:ns:meta/">`)
	xmp.WriteString("\n")
	xmp.WriteString(`  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">`)
	xmp.WriteString("\n")

	// PDF/UA Extension Schema definition
	xmp.WriteString(`    <rdf:Description rdf:about=""
		xmlns:pdfaExtension="http://www.aiim.org/pdfa/ns/extension/"
		xmlns:pdfaSchema="http://www.aiim.org/pdfa/ns/schema#"
		xmlns:pdfaProperty="http://www.aiim.org/pdfa/ns/property#">
	  <pdfaExtension:schemas>
		<rdf:Bag>
		  <rdf:li rdf:parseType="Resource">
			<pdfaSchema:schema>PDF/UA Universal Accessibility Schema</pdfaSchema:schema>
			<pdfaSchema:namespaceURI>http://www.aiim.org/pdfua/ns/id/</pdfaSchema:namespaceURI>
			<pdfaSchema:prefix>pdfuaid</pdfaSchema:prefix>
			<pdfaSchema:property>
			  <rdf:Seq>
				<rdf:li rdf:parseType="Resource">
				  <pdfaProperty:name>part</pdfaProperty:name>
				  <pdfaProperty:valueType>Integer</pdfaProperty:valueType>
				  <pdfaProperty:category>internal</pdfaProperty:category>
				  <pdfaProperty:description>Indicates, which part of ISO 14289 standards the document adheres to.</pdfaProperty:description>
				</rdf:li>
				<rdf:li rdf:parseType="Resource">
				  <pdfaProperty:name>rev</pdfaProperty:name>
				  <pdfaProperty:valueType>Integer</pdfaProperty:valueType>
				  <pdfaProperty:category>internal</pdfaProperty:category>
				  <pdfaProperty:description>Indicates the year of the revision of ISO 14289 standards the document adheres to.</pdfaProperty:description>
				</rdf:li>
			  </rdf:Seq>
			</pdfaSchema:property>
		  </rdf:li>
		</rdf:Bag>
	  </pdfaExtension:schemas>
	</rdf:Description>`)
	xmp.WriteString("\n")

	// PDF/A and PDF/UA identification
	xmp.WriteString(`    <rdf:Description rdf:about="" xmlns:pdfaid="http://www.aiim.org/pdfa/ns/id/" xmlns:pdfuaid="http://www.aiim.org/pdfua/ns/id/">`)
	xmp.WriteString("\n")
	xmp.WriteString(fmt.Sprintf(`      <pdfaid:part>%d</pdfaid:part>`, part))
	xmp.WriteString("\n")
	if part == 4 {
		xmp.WriteString(`      <pdfaid:rev>2020</pdfaid:rev>`)
		xmp.WriteString("\n")
	} else if conformance != "" {
		xmp.WriteString(fmt.Sprintf(`      <pdfaid:conformance>%s</pdfaid:conformance>`, conformance))
		xmp.WriteString("\n")
	}
	xmp.WriteString("\n")
	xmp.WriteString(`      <pdfuaid:part>2</pdfuaid:part>`)
	xmp.WriteString("\n")
	xmp.WriteString(`      <pdfuaid:rev>2024</pdfuaid:rev>`)
	xmp.WriteString("\n")
	xmp.WriteString(`    </rdf:Description>`)
	xmp.WriteString("\n")

	// XMP basic properties
	xmp.WriteString(`    <rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/">`)
	xmp.WriteString("\n")
	xmp.WriteString(fmt.Sprintf(`      <xmp:CreateDate>%s</xmp:CreateDate>`, createDate))
	xmp.WriteString("\n")
	xmp.WriteString(fmt.Sprintf(`      <xmp:ModifyDate>%s</xmp:ModifyDate>`, modifyDate))
	xmp.WriteString("\n")
	xmp.WriteString(fmt.Sprintf(`      <xmp:MetadataDate>%s</xmp:MetadataDate>`, modifyDate))
	xmp.WriteString("\n")
	if h.config.Creator != "" {
		xmp.WriteString(fmt.Sprintf(`      <xmp:CreatorTool>%s</xmp:CreatorTool>`, escapeXML(h.config.Creator)))
		xmp.WriteString("\n")
	} else {
		xmp.WriteString(`      <xmp:CreatorTool>GoPDFSuit</xmp:CreatorTool>`)
		xmp.WriteString("\n")
	}
	xmp.WriteString(`    </rdf:Description>`)
	xmp.WriteString("\n")

	// Dublin Core properties
	xmp.WriteString(`    <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">`)
	xmp.WriteString("\n")
	xmp.WriteString(`      <dc:format>application/pdf</dc:format>`)
	xmp.WriteString("\n")
	if h.config.Title != "" {
		xmp.WriteString(`      <dc:title>`)
		xmp.WriteString("\n")
		xmp.WriteString(`        <rdf:Alt>`)
		xmp.WriteString("\n")
		xmp.WriteString(fmt.Sprintf(`          <rdf:li xml:lang="x-default">%s</rdf:li>`, escapeXML(h.config.Title)))
		xmp.WriteString("\n")
		xmp.WriteString(`        </rdf:Alt>`)
		xmp.WriteString("\n")
		xmp.WriteString(`      </dc:title>`)
		xmp.WriteString("\n")
	}
	if h.config.Author != "" {
		xmp.WriteString(`      <dc:creator>`)
		xmp.WriteString("\n")
		xmp.WriteString(`        <rdf:Seq>`)
		xmp.WriteString("\n")
		xmp.WriteString(fmt.Sprintf(`          <rdf:li>%s</rdf:li>`, escapeXML(h.config.Author)))
		xmp.WriteString("\n")
		xmp.WriteString(`        </rdf:Seq>`)
		xmp.WriteString("\n")
		xmp.WriteString(`      </dc:creator>`)
		xmp.WriteString("\n")
	}
	if h.config.Subject != "" {
		xmp.WriteString(`      <dc:description>`)
		xmp.WriteString("\n")
		xmp.WriteString(`        <rdf:Alt>`)
		xmp.WriteString("\n")
		xmp.WriteString(fmt.Sprintf(`          <rdf:li xml:lang="x-default">%s</rdf:li>`, escapeXML(h.config.Subject)))
		xmp.WriteString("\n")
		xmp.WriteString(`        </rdf:Alt>`)
		xmp.WriteString("\n")
		xmp.WriteString(`      </dc:description>`)
		xmp.WriteString("\n")
	}
	if h.config.Keywords != "" {
		xmp.WriteString(`      <dc:subject>`)
		xmp.WriteString("\n")
		xmp.WriteString(`        <rdf:Bag>`)
		xmp.WriteString("\n")
		keywords := strings.Split(h.config.Keywords, ",")
		for _, kw := range keywords {
			kw = strings.TrimSpace(kw)
			if kw != "" {
				xmp.WriteString(fmt.Sprintf(`          <rdf:li>%s</rdf:li>`, escapeXML(kw)))
				xmp.WriteString("\n")
			}
		}
		xmp.WriteString(`        </rdf:Bag>`)
		xmp.WriteString("\n")
		xmp.WriteString(`      </dc:subject>`)
		xmp.WriteString("\n")
	}
	xmp.WriteString(`    </rdf:Description>`)
	xmp.WriteString("\n")

	// PDF properties
	xmp.WriteString(`    <rdf:Description rdf:about="" xmlns:pdf="http://ns.adobe.com/pdf/1.3/">`)
	xmp.WriteString("\n")
	xmp.WriteString(`      <pdf:Producer>GoPDFSuit</pdf:Producer>`)
	xmp.WriteString("\n")
	xmp.WriteString(`    </rdf:Description>`)
	xmp.WriteString("\n")

	// XMP Media Management
	xmp.WriteString(`    <rdf:Description rdf:about="" xmlns:xmpMM="http://ns.adobe.com/xap/1.0/mm/">`)
	xmp.WriteString("\n")
	xmp.WriteString(fmt.Sprintf(`      <xmpMM:DocumentID>uuid:%s</xmpMM:DocumentID>`, documentID))
	xmp.WriteString("\n")
	xmp.WriteString(fmt.Sprintf(`      <xmpMM:InstanceID>uuid:%s</xmpMM:InstanceID>`, documentID))
	xmp.WriteString("\n")
	xmp.WriteString(`    </rdf:Description>`)
	xmp.WriteString("\n")

	xmp.WriteString(`  </rdf:RDF>`)
	xmp.WriteString("\n")
	xmp.WriteString(`</x:xmpmeta>`)
	xmp.WriteString("\n")

	// Add padding for future editing (required by XMP spec)
	padding := strings.Repeat(" ", 2000)
	xmp.WriteString(padding)
	xmp.WriteString("\n")
	xmp.WriteString(`<?xpacket end="w"?>`)

	xmpContent := xmp.String()

	// Create metadata stream object
	h.metadataObjID = h.pageManager.NextObjectID
	h.pageManager.NextObjectID++

	var streamContent []byte = []byte(xmpContent)

	// Encrypt if needed
	if h.encryptor != nil {
		streamContent = h.encryptor.EncryptStream(streamContent, h.metadataObjID, 0)
	}

	metadataDict := fmt.Sprintf("<< /Type /Metadata /Subtype /XML /Length %d >>\nstream\n%s\nendstream",
		len(streamContent), string(streamContent))

	return h.metadataObjID, metadataDict
}

// GenerateOutputIntent generates the OutputIntent for PDF/A with embedded sRGB ICC profile
// Returns (outputIntentObjID, []strings of objects, compressedICCData)
func (h *PDFAHandler) GenerateOutputIntent(iccID, outputIntentID int) (int, []string, []byte) {
	objects := make([]string, 0, 2)

	// Create ICC profile object (sRGB)
	// This is a minimal sRGB ICC profile for PDF/A compliance
	if iccID > 0 {
		h.iccProfileObjID = iccID
	} else {
		h.iccProfileObjID = h.pageManager.NextObjectID
		h.pageManager.NextObjectID++
	}

	iccData := getSRGBICCProfile()

	// Compress the ICC profile with zlib (FlateDecode)
	var compressedBuf bytes.Buffer
	zlibWriter := zlib.NewWriter(&compressedBuf)
	zlibWriter.Write(iccData)
	zlibWriter.Close()
	compressedData := compressedBuf.Bytes()

	// Encrypt compressed ICC profile stream if needed
	if h.encryptor != nil {
		compressedData = h.encryptor.EncryptStream(compressedData, h.iccProfileObjID, 0)
	}

	iccDict := fmt.Sprintf("<< /N 3 /Length %d /Filter /FlateDecode /Alternate /DeviceRGB >>\nstream\n", len(compressedData))
	objects = append(objects, fmt.Sprintf("%d 0 obj\n%s", h.iccProfileObjID, iccDict))

	// Create OutputIntent object
	if outputIntentID > 0 {
		h.outputIntentObjID = outputIntentID
	} else {
		h.outputIntentObjID = h.pageManager.NextObjectID
		h.pageManager.NextObjectID++
	}

	// Encrypt string values in OutputIntent dictionary if needed
	idStr := "(sRGB IEC61966-2.1)"
	regStr := "(http://www.color.org)"
	infoStr := "(sRGB IEC61966-2.1)"

	if h.encryptor != nil {
		idEnc := h.encryptor.EncryptString([]byte("sRGB IEC61966-2.1"), h.outputIntentObjID, 0)
		idStr = fmt.Sprintf("<%s>", hex.EncodeToString(idEnc))

		regEnc := h.encryptor.EncryptString([]byte("http://www.color.org"), h.outputIntentObjID, 0)
		regStr = fmt.Sprintf("<%s>", hex.EncodeToString(regEnc))

		infoEnc := h.encryptor.EncryptString([]byte("sRGB IEC61966-2.1"), h.outputIntentObjID, 0)
		infoStr = fmt.Sprintf("<%s>", hex.EncodeToString(infoEnc))
	}

	outputIntentDict := fmt.Sprintf(`<< /Type /OutputIntent /S /GTS_PDFA1 /OutputConditionIdentifier %s /RegistryName %s /Info %s /DestOutputProfile %d 0 R >>`,
		idStr, regStr, infoStr, h.iccProfileObjID)
	objects = append(objects, fmt.Sprintf("%d 0 obj\n%s\nendobj", h.outputIntentObjID, outputIntentDict))

	return h.outputIntentObjID, objects, compressedData
}

// GetMetadataObjID returns the metadata object ID
func (h *PDFAHandler) GetMetadataObjID() int {
	return h.metadataObjID
}

// GetOutputIntentObjID returns the output intent object ID
func (h *PDFAHandler) GetOutputIntentObjID() int {
	return h.outputIntentObjID
}

// GetICCProfileObjID returns the ICC profile object ID
func (h *PDFAHandler) GetICCProfileObjID() int {
	return h.iccProfileObjID
}

// escapeXML escapes special XML characters
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// getSRGBICCProfile returns the complete sRGB ICC profile for PDF/A compliance
// Uses the properly built profile from pdfa.go to ensure validity
func getSRGBICCProfile() []byte {
	return GetSRGBICCProfile()
}

// GenerateCatalogExtras returns additional catalog entries for PDF/A
func (h *PDFAHandler) GenerateCatalogExtras() string {
	var extras strings.Builder

	if h.metadataObjID > 0 {
		extras.WriteString(fmt.Sprintf(" /Metadata %d 0 R", h.metadataObjID))
	}

	if h.outputIntentObjID > 0 {
		extras.WriteString(fmt.Sprintf(" /OutputIntents [%d 0 R]", h.outputIntentObjID))
	}

	// PDF/A requires MarkInfo with Marked = true for tagged PDF
	extras.WriteString(" /MarkInfo << /Marked true >>")

	return extras.String()
}
