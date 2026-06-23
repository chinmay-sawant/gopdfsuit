package pdf

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/font"
)

var compressedSRGBICCProfileCache = sync.OnceValue(compressSRGBICCProfile)
var xmpPacketPadding = strings.Repeat(" ", 2000)

type xmpMetadataTemplate struct {
	prefix                   string
	betweenCreateAndModify   string
	betweenModifyAndMetadata string
	beforeDocumentID         string
	betweenDocumentIDs       string
	suffix                   string
	sizeHint                 int
}

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
	xmpTemplate       xmpMetadataTemplate
}

// NewPDFAHandler creates a new PDF/A handler
func NewPDFAHandler(config *models.PDFAConfig, pm *PageManager, encryptor ObjectEncryptor) *PDFAHandler {
	return &PDFAHandler{
		config:      config,
		pageManager: pm,
		encryptor:   encryptor,
		xmpTemplate: buildXMPMetadataTemplate(config),
	}
}

// GetConformanceLevel returns the PDF/A part and conformance level
func (h *PDFAHandler) GetConformanceLevel() (part int, conformance string) {
	return getPDFAConformanceLevel(h.config)
}

func getPDFAConformanceLevel(config *models.PDFAConfig) (part int, conformance string) {
	if config == nil {
		return 4, ""
	}

	switch config.Conformance {
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

func buildXMPMetadataTemplate(config *models.PDFAConfig) xmpMetadataTemplate {
	part, conformance := getPDFAConformanceLevel(config)

	title := ""
	author := ""
	subject := ""
	keywords := ""
	creatorTool := "GoPDFSuit"
	if config != nil {
		title = config.Title
		author = config.Author
		subject = config.Subject
		keywords = config.Keywords
		if config.Creator != "" {
			creatorTool = config.Creator
		}
	}

	var prefix strings.Builder
	prefix.Grow(6144 + len(title) + len(author) + len(subject) + len(keywords) + len(creatorTool))
	prefix.WriteString(`<?xpacket begin="` + "\xef\xbb\xbf" + `" id="W5M0MpCehiHzreSzNTczkc9d"?>`)
	prefix.WriteString("\n")
	prefix.WriteString(`<x:xmpmeta xmlns:x="adobe:ns:meta/">`)
	prefix.WriteString("\n")
	prefix.WriteString(`  <rdf:RDF xmlns:rdf="http://www.w3.org/1999/02/22-rdf-syntax-ns#">`)
	prefix.WriteString("\n")
	prefix.WriteString(`    <rdf:Description rdf:about=""
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
	prefix.WriteString("\n")
	prefix.WriteString(`    <rdf:Description rdf:about="" xmlns:pdfaid="http://www.aiim.org/pdfa/ns/id/" xmlns:pdfuaid="http://www.aiim.org/pdfua/ns/id/">`)
	prefix.WriteString("\n")
	prefix.WriteString("      <pdfaid:part>")
	prefix.WriteString(strconv.Itoa(part))
	prefix.WriteString("</pdfaid:part>")
	prefix.WriteString("\n")
	if part == 4 {
		prefix.WriteString(`      <pdfaid:rev>2020</pdfaid:rev>`)
		prefix.WriteString("\n")
	} else if conformance != "" {
		prefix.WriteString("      <pdfaid:conformance>" + conformance + "</pdfaid:conformance>")
		prefix.WriteString("\n")
	}
	prefix.WriteString("\n")
	prefix.WriteString(`      <pdfuaid:part>2</pdfuaid:part>`)
	prefix.WriteString("\n")
	prefix.WriteString(`      <pdfuaid:rev>2024</pdfuaid:rev>`)
	prefix.WriteString("\n")
	prefix.WriteString(`    </rdf:Description>`)
	prefix.WriteString("\n")
	prefix.WriteString(`    <rdf:Description rdf:about="" xmlns:xmp="http://ns.adobe.com/xap/1.0/">`)
	prefix.WriteString("\n")
	prefix.WriteString(`      <xmp:CreateDate>`)

	const betweenCreateAndModify = `</xmp:CreateDate>
      <xmp:ModifyDate>`
	const betweenModifyAndMetadata = `</xmp:ModifyDate>
      <xmp:MetadataDate>`
	const betweenDocumentIDs = `</xmpMM:DocumentID>
      <xmpMM:InstanceID>uuid:`

	var beforeDocumentID strings.Builder
	beforeDocumentID.Grow(2048 + len(title) + len(author) + len(subject) + len(keywords) + len(creatorTool))
	beforeDocumentID.WriteString(`</xmp:MetadataDate>`)
	beforeDocumentID.WriteString("\n")
	beforeDocumentID.WriteString(`      <xmp:CreatorTool>`)
	beforeDocumentID.WriteString(escapeXML(creatorTool))
	beforeDocumentID.WriteString(`</xmp:CreatorTool>`)
	beforeDocumentID.WriteString("\n")
	beforeDocumentID.WriteString(`    </rdf:Description>`)
	beforeDocumentID.WriteString("\n")
	beforeDocumentID.WriteString(`    <rdf:Description rdf:about="" xmlns:dc="http://purl.org/dc/elements/1.1/">`)
	beforeDocumentID.WriteString("\n")
	beforeDocumentID.WriteString(`      <dc:format>application/pdf</dc:format>`)
	beforeDocumentID.WriteString("\n")
	if title != "" {
		beforeDocumentID.WriteString(`      <dc:title>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`        <rdf:Alt>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`          <rdf:li xml:lang="x-default">`)
		beforeDocumentID.WriteString(escapeXML(title))
		beforeDocumentID.WriteString(`</rdf:li>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`        </rdf:Alt>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`      </dc:title>`)
		beforeDocumentID.WriteString("\n")
	}
	if author != "" {
		beforeDocumentID.WriteString(`      <dc:creator>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`        <rdf:Seq>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`          <rdf:li>`)
		beforeDocumentID.WriteString(escapeXML(author))
		beforeDocumentID.WriteString(`</rdf:li>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`        </rdf:Seq>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`      </dc:creator>`)
		beforeDocumentID.WriteString("\n")
	}
	if subject != "" {
		beforeDocumentID.WriteString(`      <dc:description>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`        <rdf:Alt>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`          <rdf:li xml:lang="x-default">`)
		beforeDocumentID.WriteString(escapeXML(subject))
		beforeDocumentID.WriteString(`</rdf:li>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`        </rdf:Alt>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`      </dc:description>`)
		beforeDocumentID.WriteString("\n")
	}
	if keywords != "" {
		beforeDocumentID.WriteString(`      <dc:subject>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`        <rdf:Bag>`)
		beforeDocumentID.WriteString("\n")
		for kw := range strings.SplitSeq(keywords, ",") {
			kw = strings.TrimSpace(kw)
			if kw == "" {
				continue
			}
			beforeDocumentID.WriteString(`          <rdf:li>`)
			beforeDocumentID.WriteString(escapeXML(kw))
			beforeDocumentID.WriteString(`</rdf:li>`)
			beforeDocumentID.WriteString("\n")
		}
		beforeDocumentID.WriteString(`        </rdf:Bag>`)
		beforeDocumentID.WriteString("\n")
		beforeDocumentID.WriteString(`      </dc:subject>`)
		beforeDocumentID.WriteString("\n")
	}
	beforeDocumentID.WriteString(`    </rdf:Description>`)
	beforeDocumentID.WriteString("\n")
	beforeDocumentID.WriteString(`    <rdf:Description rdf:about="" xmlns:pdf="http://ns.adobe.com/pdf/1.3/">`)
	beforeDocumentID.WriteString("\n")
	beforeDocumentID.WriteString(`      <pdf:Producer>GoPDFSuit</pdf:Producer>`)
	beforeDocumentID.WriteString("\n")
	beforeDocumentID.WriteString(`    </rdf:Description>`)
	beforeDocumentID.WriteString("\n")
	beforeDocumentID.WriteString(`    <rdf:Description rdf:about="" xmlns:xmpMM="http://ns.adobe.com/xap/1.0/mm/">`)
	beforeDocumentID.WriteString("\n")
	beforeDocumentID.WriteString(`      <xmpMM:DocumentID>uuid:`)

	var suffix strings.Builder
	suffix.Grow(256 + len(xmpPacketPadding))
	suffix.WriteString(`</xmpMM:InstanceID>`)
	suffix.WriteString("\n")
	suffix.WriteString(`    </rdf:Description>`)
	suffix.WriteString("\n")
	suffix.WriteString(`  </rdf:RDF>`)
	suffix.WriteString("\n")
	suffix.WriteString(`</x:xmpmeta>`)
	suffix.WriteString("\n")
	suffix.WriteString(xmpPacketPadding)
	suffix.WriteString("\n")
	suffix.WriteString(`<?xpacket end="w"?>`)

	prefixStr := prefix.String()
	beforeDocumentIDStr := beforeDocumentID.String()
	suffixStr := suffix.String()

	return xmpMetadataTemplate{
		prefix:                   prefixStr,
		betweenCreateAndModify:   betweenCreateAndModify,
		betweenModifyAndMetadata: betweenModifyAndMetadata,
		beforeDocumentID:         beforeDocumentIDStr,
		betweenDocumentIDs:       betweenDocumentIDs,
		suffix:                   suffixStr,
		sizeHint: len(prefixStr) + len(betweenCreateAndModify) + len(betweenModifyAndMetadata) +
			len(beforeDocumentIDStr) + len(betweenDocumentIDs) + len(suffixStr),
	}
}

// GenerateXMPMetadata generates the XMP metadata stream for PDF/A
func (h *PDFAHandler) GenerateXMPMetadata(documentID string, generatedAt time.Time) (int, string) {
	createDate := generatedAt.UTC().Format("2006-01-02T15:04:05Z")
	template := h.xmpTemplate
	var xmp strings.Builder
	xmp.Grow(template.sizeHint + len(createDate)*3 + len(documentID)*2)
	xmp.WriteString(template.prefix)
	xmp.WriteString(createDate)
	xmp.WriteString(template.betweenCreateAndModify)
	xmp.WriteString(createDate)
	xmp.WriteString(template.betweenModifyAndMetadata)
	xmp.WriteString(createDate)
	xmp.WriteString(template.beforeDocumentID)
	xmp.WriteString(documentID)
	xmp.WriteString(template.betweenDocumentIDs)
	xmp.WriteString(documentID)
	xmp.WriteString(template.suffix)

	xmpContent := []byte(xmp.String())

	// Create metadata stream object
	h.metadataObjID = h.pageManager.NextObjectID
	h.pageManager.NextObjectID++

	// Encrypt if needed
	var metaBuf strings.Builder
	if h.encryptor != nil {
		streamContent := h.encryptor.EncryptStream(xmpContent, h.metadataObjID, 0)
		metaBuf.Grow(128 + len(streamContent))
		metaBuf.WriteString("<< /Type /Metadata /Subtype /XML /Length ")
		metaBuf.WriteString(strconv.Itoa(len(streamContent)))
		metaBuf.WriteString(" >>\nstream\n")
		metaBuf.Write(streamContent)
		metaBuf.WriteString("\nendstream")
	} else {
		metaBuf.Grow(128 + len(xmpContent))
		metaBuf.WriteString("<< /Type /Metadata /Subtype /XML /Length ")
		metaBuf.WriteString(strconv.Itoa(len(xmpContent)))
		metaBuf.WriteString(" >>\nstream\n")
		metaBuf.Write(xmpContent)
		metaBuf.WriteString("\nendstream")
	}
	metadataDict := metaBuf.String()

	return h.metadataObjID, metadataDict
}

// GenerateOutputIntent generates the OutputIntent for PDF/A with embedded sRGB ICC profile
// Returns (outputIntentObjID, []strings of objects, compressedICCData)
func (h *PDFAHandler) GenerateOutputIntent(iccID, outputIntentID int) (int, []string, []byte) {
	objects := make([]string, 0, 2)
	var sb strings.Builder
	sb.Grow(160)

	// Create ICC profile object (sRGB)
	// This is a minimal sRGB ICC profile for PDF/A compliance
	if iccID > 0 {
		h.iccProfileObjID = iccID
	} else {
		h.iccProfileObjID = h.pageManager.NextObjectID
		h.pageManager.NextObjectID++
	}

	compressedData := compressedSRGBICCProfileCache()
	if h.encryptor != nil {
		// Encryption is object-ID dependent, but it can reuse the cached
		// compressed ICC stream as the input payload.
		compressedData = h.encryptor.EncryptStream(compressedData, h.iccProfileObjID, 0)
	}

	var iccBuf strings.Builder
	iccBuf.Grow(96 + len(compressedData))
	iccBuf.WriteString("<< /N 3 /Length ")
	iccBuf.WriteString(strconv.Itoa(len(compressedData)))
	iccBuf.WriteString(" /Filter /FlateDecode /Alternate /DeviceRGB >>\nstream\n")
	iccDict := iccBuf.String()
	sb.Reset()
	sb.WriteString(strconv.Itoa(h.iccProfileObjID))
	sb.WriteString(" 0 obj\n")
	sb.WriteString(iccDict)
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
		idStr = fmt.Sprintf("<%s>", hex.EncodeToString(idEnc))

		regEnc := h.encryptor.EncryptString([]byte("http://www.color.org"), h.outputIntentObjID, 0)
		regStr = fmt.Sprintf("<%s>", hex.EncodeToString(regEnc))

		infoEnc := h.encryptor.EncryptString([]byte("sRGB IEC61966-2.1"), h.outputIntentObjID, 0)
		infoStr = fmt.Sprintf("<%s>", hex.EncodeToString(infoEnc))
	}

	outputIntentDict := fmt.Sprintf(`<< /Type /OutputIntent /S /GTS_PDFA1 /OutputConditionIdentifier %s /RegistryName %s /Info %s /DestOutputProfile %d 0 R >>`,
		idStr, regStr, infoStr, h.iccProfileObjID)
	sb.Reset()
	sb.WriteString(strconv.Itoa(h.outputIntentObjID))
	sb.WriteString(" 0 obj\n")
	sb.WriteString(outputIntentDict)
	sb.WriteString("\nendobj")
	objects = append(objects, sb.String())

	return h.outputIntentObjID, objects, compressedData
}

func compressSRGBICCProfile() []byte {
	iccData := getSRGBICCProfile()
	compressedBuf := font.GetCompressBuffer()
	zlibWriter := font.GetZlibWriter(compressedBuf)
	if _, err := zlibWriter.Write(iccData); err != nil {
		_ = zlibWriter.Close()
		font.PutZlibWriter(zlibWriter)
		font.PutCompressBuffer(compressedBuf)
		return nil
	}
	_ = zlibWriter.Close()
	font.PutZlibWriter(zlibWriter)
	compressedData := make([]byte, compressedBuf.Len())
	copy(compressedData, compressedBuf.Bytes())
	font.PutCompressBuffer(compressedBuf)
	return compressedData
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
	return srgbICCProfileRaw
}

// GenerateCatalogExtras returns additional catalog entries for PDF/A
func (h *PDFAHandler) GenerateCatalogExtras() string {
	var extras strings.Builder
	extras.Grow(96)

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
