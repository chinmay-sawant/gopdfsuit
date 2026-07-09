package pdf

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/byteconv"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

// PDFAHandler handles PDF/A compliance features, including metadata and color profiles.
//
//nolint:revive // exported
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
	xmp.Grow(8192)
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
	xmp.WriteString(`      <pdfaid:part>`)
	var ptmp [8]byte
	xmp.Write(strconv.AppendInt(ptmp[:0], int64(part), 10))
	xmp.WriteString(`</pdfaid:part>`)
	xmp.WriteString("\n")
	if part == 4 {
		xmp.WriteString(`      <pdfaid:rev>2020</pdfaid:rev>`)
		xmp.WriteString("\n")
	} else if conformance != "" {
		xmp.WriteString(`      <pdfaid:conformance>`)
		xmp.WriteString(conformance)
		xmp.WriteString(`</pdfaid:conformance>`)
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
	xmp.WriteString(`      <xmp:CreateDate>`)
	xmp.WriteString(createDate)
	xmp.WriteString(`</xmp:CreateDate>`)
	xmp.WriteString("\n")
	xmp.WriteString(`      <xmp:ModifyDate>`)
	xmp.WriteString(modifyDate)
	xmp.WriteString(`</xmp:ModifyDate>`)
	xmp.WriteString("\n")
	xmp.WriteString(`      <xmp:MetadataDate>`)
	xmp.WriteString(modifyDate)
	xmp.WriteString(`</xmp:MetadataDate>`)
	xmp.WriteString("\n")
	if h.config.Creator != "" {
		xmp.WriteString(`      <xmp:CreatorTool>`)
		xmp.WriteString(escapeXML(h.config.Creator))
		xmp.WriteString(`</xmp:CreatorTool>`)
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
		xmp.WriteString(`          <rdf:li xml:lang="x-default">`)
		xmp.WriteString(escapeXML(h.config.Title))
		xmp.WriteString(`</rdf:li>`)
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
		xmp.WriteString(`          <rdf:li>`)
		xmp.WriteString(escapeXML(h.config.Author))
		xmp.WriteString(`</rdf:li>`)
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
		xmp.WriteString(`          <rdf:li xml:lang="x-default">`)
		xmp.WriteString(escapeXML(h.config.Subject))
		xmp.WriteString(`</rdf:li>`)
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
			kw = trimSpace(kw)
			if kw != "" {
				xmp.WriteString(`          <rdf:li>`)
				xmp.WriteString(escapeXML(kw))
				xmp.WriteString(`</rdf:li>`)
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
	xmp.WriteString(`      <xmpMM:DocumentID>uuid:`)
	xmp.WriteString(documentID)
	xmp.WriteString(`</xmpMM:DocumentID>`)
	xmp.WriteString("\n")
	xmp.WriteString(`      <xmpMM:InstanceID>uuid:`)
	xmp.WriteString(documentID)
	xmp.WriteString(`</xmpMM:InstanceID>`)
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

	streamContent := byteconv.StringToBytes(xmpContent)

	// Encrypt if needed
	if h.encryptor != nil {
		streamContent = h.encryptor.EncryptStream(streamContent, h.metadataObjID, 0)
	}

	// Build metadata dict without re-converting streamContent to string (PERF-32)
	var metaSB strings.Builder
	metaSB.Grow(64 + len(streamContent))
	metaSB.WriteString("<< /Type /Metadata /Subtype /XML /Length ")
	metaSB.Write(strconv.AppendInt(nil, int64(len(streamContent)), 10))
	metaSB.WriteString(" >>\nstream\n")
	metaSB.Write(streamContent)
	metaSB.WriteString("\nendstream")
	metadataDict := metaSB.String()

	return h.metadataObjID, metadataDict
}

// GenerateOutputIntent generates the OutputIntent for PDF/A with embedded sRGB ICC profile
// Returns (outputIntentObjID, []strings of objects, compressedICCData)
func (h *PDFAHandler) GenerateOutputIntent(iccID, outputIntentID int) (int, []string, []byte) {
	objects := make([]string, 0, 2)
	var sb strings.Builder

	// Create ICC profile object (sRGB)
	// This is a minimal sRGB ICC profile for PDF/A compliance
	if iccID > 0 {
		h.iccProfileObjID = iccID
	} else {
		h.iccProfileObjID = h.pageManager.NextObjectID
		h.pageManager.NextObjectID++
	}

	// Use cached compressed ICC profile (PERF-217: avoid recompression)
	compressedData := bytes.Clone(GetSRGBICCCompressed())

	// Encrypt compressed ICC profile stream if needed
	if h.encryptor != nil {
		compressedData = h.encryptor.EncryptStream(compressedData, h.iccProfileObjID, 0)
	}

	sb.Reset()
	sb.Grow(128)
	var tmp [20]byte
	sb.Write(strconv.AppendInt(tmp[:0], int64(h.iccProfileObjID), 10))
	sb.WriteString(" 0 obj\n<< /N 3 /Length ")
	sb.Write(strconv.AppendInt(tmp[:0], int64(len(compressedData)), 10))
	sb.WriteString(" /Filter /FlateDecode /Alternate /DeviceRGB >>\nstream\n")
	objects = append(objects, sb.String())

	// Create OutputIntent object
	if outputIntentID > 0 {
		h.outputIntentObjID = outputIntentID
	} else {
		h.outputIntentObjID = h.pageManager.NextObjectID
		h.pageManager.NextObjectID++
	}

	// Encrypt string values in OutputIntent dictionary if needed
	idStr := "(sRGB IEC61966-2.1)" //nolint:goconst
	regStr := "(http://www.color.org)"
	infoStr := "(sRGB IEC61966-2.1)"

	if h.encryptor != nil {
		idEnc := h.encryptor.EncryptString([]byte("sRGB IEC61966-2.1"), h.outputIntentObjID, 0)
		idStr = "<" + hex.EncodeToString(idEnc) + ">"

		regEnc := h.encryptor.EncryptString([]byte("http://www.color.org"), h.outputIntentObjID, 0)
		regStr = "<" + hex.EncodeToString(regEnc) + ">"

		infoEnc := h.encryptor.EncryptString([]byte("sRGB IEC61966-2.1"), h.outputIntentObjID, 0)
		infoStr = "<" + hex.EncodeToString(infoEnc) + ">"
	}

	sb.Reset()
	sb.Grow(160 + len(idStr) + len(regStr) + len(infoStr))
	sb.Write(strconv.AppendInt(tmp[:0], int64(h.outputIntentObjID), 10))
	sb.WriteString(" 0 obj\n<< /Type /OutputIntent /S /GTS_PDFA1 /OutputConditionIdentifier ")
	sb.WriteString(idStr)
	sb.WriteString(" /RegistryName ")
	sb.WriteString(regStr)
	sb.WriteString(" /Info ")
	sb.WriteString(infoStr)
	sb.WriteString(" /DestOutputProfile ")
	sb.Write(strconv.AppendInt(tmp[:0], int64(h.iccProfileObjID), 10))
	sb.WriteString(" 0 R >>\nendobj")
	objects = append(objects, sb.String())

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

// GenerateCatalogExtras returns additional catalog entries for PDF/A
func (h *PDFAHandler) GenerateCatalogExtras() string {
	var extras strings.Builder
	var tmp [20]byte

	if h.metadataObjID > 0 {
		extras.WriteString(" /Metadata ")
		extras.Write(strconv.AppendInt(tmp[:0], int64(h.metadataObjID), 10))
		extras.WriteString(" 0 R")
	}

	if h.outputIntentObjID > 0 {
		extras.WriteString(" /OutputIntents [")
		extras.Write(strconv.AppendInt(tmp[:0], int64(h.outputIntentObjID), 10))
		extras.WriteString(" 0 R]")
	}

	// PDF/A requires MarkInfo with Marked = true for tagged PDF
	extras.WriteString(" /MarkInfo << /Marked true >>")

	return extras.String()
}
