package pdf

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// PDFSigner handles digital signatures for PDF documents
type PDFSigner struct {
	config      *models.SignatureConfig
	certificate *x509.Certificate
	privateKey  crypto.PrivateKey
	certChain   []*x509.Certificate
}

// Signature field and annotation IDs
type SignatureIDs struct {
	SigFieldID     int
	SigAnnotID     int
	AppearanceID   int
	ByteRangeStart int // Position of ByteRange placeholder in PDF
	ContentsStart  int // Position of Contents placeholder in PDF
}

// OID values for CMS/PKCS#7
var (
	oidData          = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}
	oidSignedData    = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	oidSHA256        = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
	oidRSAEncryption = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	oidContentType   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}
	oidMessageDigest = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}
	oidSigningTime   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5}
)

// NewPDFSigner creates a new PDF signer from config
func NewPDFSigner(config *models.SignatureConfig) (*PDFSigner, error) {
	if config == nil || !config.Enabled {
		return nil, nil
	}

	signer := &PDFSigner{
		config: config,
	}

	// Parse certificate
	block, _ := pem.Decode([]byte(config.CertificatePEM))
	if block == nil {
		return nil, fmt.Errorf("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}
	signer.certificate = cert

	// Parse private key
	keyBlock, _ := pem.Decode([]byte(config.PrivateKeyPEM))
	if keyBlock == nil {
		return nil, fmt.Errorf("failed to parse private key PEM")
	}

	// Try parsing as PKCS#8 first, then PKCS#1
	var privateKey crypto.PrivateKey
	privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse private key: %w", err)
			}
		}
	}
	signer.privateKey = privateKey

	// Parse certificate chain if provided
	for _, chainPEM := range config.CertificateChain {
		chainBlock, _ := pem.Decode([]byte(chainPEM))
		if chainBlock != nil {
			chainCert, err := x509.ParseCertificate(chainBlock.Bytes)
			if err == nil {
				signer.certChain = append(signer.certChain, chainCert)
			}
		}
	}

	return signer, nil
}

// CreateSignatureField creates the signature field and annotation objects
func (s *PDFSigner) CreateSignatureField(pageManager *PageManager, pageDims PageDimensions, fontID int) *SignatureIDs {
	ids := &SignatureIDs{}

	// Determine signature rectangle
	sigX := s.config.X
	sigY := s.config.Y
	sigW := s.config.Width
	sigH := s.config.Height

	if sigW <= 0 {
		sigW = 250 // Wider to accommodate signature icon and text
	}
	if sigH <= 0 {
		sigH = 100 // Taller to fit more lines (signer, date, reason, location)
	}

	// Default position: bottom right of first page
	if sigX <= 0 {
		sigX = pageDims.Width - sigW - margin
	}
	if sigY <= 0 {
		sigY = margin
	}

	// Create appearance stream for visible signature
	if s.config.Visible {
		ids.AppearanceID = s.createSignatureAppearance(pageManager, sigW, sigH, fontID)
	}

	// Create signature value dictionary (will be filled during signing)
	sigValueID := pageManager.NextObjectID
	pageManager.NextObjectID++

	signerName := s.config.Name
	if signerName == "" && s.certificate != nil {
		signerName = s.certificate.Subject.CommonName
	}

	// Build signature value dictionary
	var sigValueDict strings.Builder
	sigValueDict.WriteString("<< /Type /Sig")
	sigValueDict.WriteString(" /Filter /Adobe.PPKLite")
	sigValueDict.WriteString(" /SubFilter /adbe.pkcs7.detached")

	// ByteRange placeholder - will be replaced during actual signing
	// Format: [0 offset1 offset2 length] where offset1 is start of /Contents, offset2 is end of /Contents
	sigValueDict.WriteString(" /ByteRange [0 0000000000 0000000000 0000000000]")

	// Contents placeholder - hex-encoded PKCS#7 signature
	// Reserve space for signature (8192 bytes = 16384 hex chars)
	sigValueDict.WriteString(" /Contents <")
	sigValueDict.WriteString(strings.Repeat("0", 16384))
	sigValueDict.WriteString(">")

	if s.config.Reason != "" {
		sigValueDict.WriteString(fmt.Sprintf(" /Reason (%s)", escapeText(s.config.Reason)))
	}
	if s.config.Location != "" {
		sigValueDict.WriteString(fmt.Sprintf(" /Location (%s)", escapeText(s.config.Location)))
	}
	if s.config.ContactInfo != "" {
		sigValueDict.WriteString(fmt.Sprintf(" /ContactInfo (%s)", escapeText(s.config.ContactInfo)))
	}
	if signerName != "" {
		sigValueDict.WriteString(fmt.Sprintf(" /Name (%s)", escapeText(signerName)))
	}

	// Signing time - PDF date format: D:YYYYMMDDHHmmSSOHH'mm'
	// where O is + or -, HH is timezone hours, mm is timezone minutes
	now := time.Now()
	_, tzOffset := now.Zone()
	tzSign := "+"
	if tzOffset < 0 {
		tzSign = "-"
		tzOffset = -tzOffset
	}
	tzHours := tzOffset / 3600
	tzMinutes := (tzOffset % 3600) / 60
	sigValueDict.WriteString(fmt.Sprintf(" /M (D:%s%s%02d'%02d')", now.Format("20060102150405"), tzSign, tzHours, tzMinutes))

	sigValueDict.WriteString(" >>")

	pageManager.ExtraObjects[sigValueID] = sigValueDict.String()

	// Create signature field widget annotation
	sigAnnotID := pageManager.NextObjectID
	pageManager.NextObjectID++
	ids.SigAnnotID = sigAnnotID

	var annotDict strings.Builder
	annotDict.WriteString("<< /Type /Annot /Subtype /Widget")
	annotDict.WriteString(" /FT /Sig")
	annotDict.WriteString(" /T (Signature1)")
	// PDF/UA-2 compliance: Widget annotations must have a label or Contents entry
	annotDict.WriteString(" /Contents (Digital Signature)")

	var annotBuf []byte
	annotBuf = append(annotBuf, " /V "...)
	annotBuf = strconv.AppendInt(annotBuf, int64(sigValueID), 10)
	annotBuf = append(annotBuf, " 0 R"...)
	annotDict.Write(annotBuf)

	annotDict.WriteString(" /F 132") // Print + Locked

	// Rectangle for visible/invisible signature
	if s.config.Visible {
		annotDict.WriteString(fmt.Sprintf(" /Rect [%s %s %s %s]",
			fmtNum(sigX), fmtNum(sigY), fmtNum(sigX+sigW), fmtNum(sigY+sigH)))
		if ids.AppearanceID > 0 {
			annotDict.WriteString(fmt.Sprintf(" /AP << /N %d 0 R >>", ids.AppearanceID))
		}
	} else {
		// Invisible signature - zero-size rectangle
		annotDict.WriteString(" /Rect [0 0 0 0]")
	}

	// Page reference - will be set when we know page object ID
	targetPage := s.config.Page
	if targetPage <= 0 {
		targetPage = 1
	}
	pageObjID := 3 + (targetPage - 1) // Pages start at object 3
	annotDict.WriteString(fmt.Sprintf(" /P %d 0 R", pageObjID))

	annotDict.WriteString(" >>")

	pageManager.ExtraObjects[sigAnnotID] = annotDict.String()

	// Add annotation to the appropriate page
	pageIndex := targetPage - 1
	if pageIndex < 0 {
		pageIndex = 0
	}
	for len(pageManager.PageAnnots) <= pageIndex {
		pageManager.PageAnnots = append(pageManager.PageAnnots, []int{})
	}
	pageManager.PageAnnots[pageIndex] = append(pageManager.PageAnnots[pageIndex], sigAnnotID)

	ids.SigFieldID = sigAnnotID // In this implementation, field = annotation

	return ids
}

// createSignatureAppearance creates the visual appearance for a visible signature
func (s *PDFSigner) createSignatureAppearance(pageManager *PageManager, width, height float64, fontID int) int {
	var appearance strings.Builder

	// Yellow background with black border
	appearance.WriteString("q\n")
	appearance.WriteString("1 1 0.8 rg\n") // Light yellow background (RGB: 255, 255, 204)
	appearance.WriteString(fmt.Sprintf("0 0 %s %s re f\n", fmtNum(width), fmtNum(height)))
	appearance.WriteString("0 0 0 RG 1 w\n") // Black border
	appearance.WriteString(fmt.Sprintf("0 0 %s %s re S\n", fmtNum(width), fmtNum(height)))
	appearance.WriteString("Q\n")

	// Text content
	signerName := s.config.Name
	if signerName == "" && s.certificate != nil {
		signerName = s.certificate.Subject.CommonName
	}

	// Check if we're using a custom font (Liberation) that needs hex encoding
	registry := GetFontRegistry()
	useHexEncoding := fontID > 0 && registry.HasFont("Helvetica")

	appearance.WriteString("BT\n")
	appearance.WriteString("/F1 9 Tf\n")
	appearance.WriteString("0 0 0 rg\n")

	// Helper to format text based on font type
	formatText := func(text string) string {
		if useHexEncoding {
			// For Liberation fonts, use hex encoding
			return EncodeTextForCustomFont("Helvetica", text)
		}
		// For standard fonts, use ASCII encoding
		return "(" + escapeText(text) + ")"
	}

	// "Digitally signed by" line
	appearance.WriteString(fmt.Sprintf("5 %s Td\n", fmtNum(height-15)))
	appearance.WriteString(fmt.Sprintf("%s Tj\n", formatText("Digitally signed by:")))

	// Mark font usage for subsetting
	if useHexEncoding {
		registry.MarkCharsUsed("Helvetica", "Digitally signed by:")
	}

	// Signer name
	appearance.WriteString("0 -12 Td\n")
	appearance.WriteString(fmt.Sprintf("%s Tj\n", formatText(signerName)))
	if useHexEncoding {
		registry.MarkCharsUsed("Helvetica", signerName)
	}

	// Date
	now := time.Now()
	dateStr := "Date: " + now.Format("2006-01-02 15:04:05")
	appearance.WriteString("0 -12 Td\n")
	appearance.WriteString(fmt.Sprintf("%s Tj\n", formatText(dateStr)))
	if useHexEncoding {
		registry.MarkCharsUsed("Helvetica", dateStr)
	}

	// Reason if provided
	if s.config.Reason != "" {
		reasonStr := "Reason: " + s.config.Reason
		appearance.WriteString("0 -12 Td\n")
		appearance.WriteString(fmt.Sprintf("%s Tj\n", formatText(reasonStr)))
		if useHexEncoding {
			registry.MarkCharsUsed("Helvetica", reasonStr)
		}
	}

	// Location if provided
	if s.config.Location != "" {
		locationStr := "Location: " + s.config.Location
		appearance.WriteString("0 -12 Td\n")
		appearance.WriteString(fmt.Sprintf("%s Tj\n", formatText(locationStr)))
		if useHexEncoding {
			registry.MarkCharsUsed("Helvetica", locationStr)
		}
	}

	appearance.WriteString("ET\n")

	appearanceContent := appearance.String()

	// Create appearance XObject
	appearanceID := pageManager.NextObjectID
	pageManager.NextObjectID++

	// Construct resources dictionary using the embedded font ID
	var resourcesDict string
	if fontID > 0 {
		// Use reference to existing embedded font
		resourcesDict = fmt.Sprintf("<< /Font << /F1 %d 0 R >> >>", fontID)
	} else {
		// Fallback for non-embedded (should be avoided for PDF/A)
		resourcesDict = "<< /Font << /F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> >> >>"
	}

	appearanceDict := fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %s %s] /Resources %s /Length %d >>\nstream\n%s\nendstream",
		fmtNum(width), fmtNum(height), resourcesDict, len(appearanceContent), appearanceContent)

	pageManager.ExtraObjects[appearanceID] = appearanceDict

	return appearanceID
}

// SignPDF signs the PDF data and returns the PKCS#7 signature
// This is called after the PDF is generated to compute the actual signature
func (s *PDFSigner) SignPDF(pdfData []byte, byteRange [4]int) ([]byte, error) {
	// Compute hash of the signed data (everything except the /Contents value)
	hasher := sha256.New()
	hasher.Write(pdfData[byteRange[0]:byteRange[1]])
	hasher.Write(pdfData[byteRange[2] : byteRange[2]+byteRange[3]])
	messageDigest := hasher.Sum(nil)

	// Create PKCS#7 SignedData structure
	signedData, err := s.createPKCS7SignedData(messageDigest)
	if err != nil {
		return nil, err
	}

	return signedData, nil
}

// createPKCS7SignedData creates a PKCS#7 SignedData structure
func (s *PDFSigner) createPKCS7SignedData(messageDigest []byte) ([]byte, error) {
	// Build authenticated attributes
	signingTime := time.Now().UTC()

	// Authenticated attributes MUST be in DER-sorted order for SET encoding
	// OIDs: ContentType (1.9.3), MessageDigest (1.9.4), SigningTime (1.9.5)
	authenticatedAttrs := []attribute{
		{
			Type: oidContentType,
			Value: asn1.RawValue{
				Class:      asn1.ClassUniversal,
				Tag:        asn1.TagSet,
				IsCompound: true,
				Bytes:      mustMarshal(oidData),
			},
		},
		{
			Type: oidMessageDigest,
			Value: asn1.RawValue{
				Class:      asn1.ClassUniversal,
				Tag:        asn1.TagSet,
				IsCompound: true,
				Bytes:      mustMarshal(messageDigest),
			},
		},
		{
			Type: oidSigningTime,
			Value: asn1.RawValue{
				Class:      asn1.ClassUniversal,
				Tag:        asn1.TagSet,
				IsCompound: true,
				Bytes:      mustMarshal(signingTime),
			},
		},
	}

	// Marshal authenticated attributes for signing
	// Go defaults to SEQUENCE for slice, but we need SET for Attributes
	// Attributes are already in DER-sorted order (ContentType < MessageDigest < SigningTime by OID)
	seqBytes, err := asn1.Marshal(authenticatedAttrs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authenticated attributes: %w", err)
	}

	// Change SEQUENCE tag (0x30) to SET tag (0x31)
	authAttrsBytes := make([]byte, len(seqBytes))
	copy(authAttrsBytes, seqBytes)
	if len(authAttrsBytes) > 0 {
		authAttrsBytes[0] = asn1.TagSet
	}

	// Sign the authenticated attributes (must be the SET encoding)
	authAttrsHash := sha256.Sum256(authAttrsBytes)

	var signature []byte
	switch key := s.privateKey.(type) {
	case *rsa.PrivateKey:
		signature, err = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, authAttrsHash[:])
		if err != nil {
			return nil, fmt.Errorf("failed to sign: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported key type")
	}

	// Extract content bytes for SignerInfo (strip SET tag and length)
	// because RawValue will add the [0] IMPLICIT tag and new length
	var contentBytes []byte
	if len(authAttrsBytes) > 1 {
		offset := 1 // Skip Tag
		// Check length byte
		if authAttrsBytes[offset]&0x80 == 0 {
			// Short form length
			offset++
		} else {
			// Long form length
			numBytes := int(authAttrsBytes[offset] & 0x7F)
			offset += 1 + numBytes
		}
		if offset <= len(authAttrsBytes) {
			contentBytes = authAttrsBytes[offset:]
		}
	}

	// Build SignerInfo
	sInfo := signerInfo{
		Version: 1,
		IssuerAndSerial: issuerAndSerial{
			Issuer:       asn1.RawValue{FullBytes: s.certificate.RawIssuer},
			SerialNumber: s.certificate.SerialNumber,
		},
		DigestAlgorithm: pkixAlgorithmIdentifier{
			Algorithm: oidSHA256,
		},
		AuthenticatedAttributes: asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        0,
			IsCompound: true,
			Bytes:      contentBytes,
		},
		DigestEncryptionAlgorithm: pkixAlgorithmIdentifier{
			Algorithm: oidRSAEncryption,
		},
		EncryptedDigest: signature,
	}

	// Build certificate chain bytes (signer cert + chain certs)
	var certBytes []byte
	certBytes = append(certBytes, s.certificate.Raw...)
	for _, chainCert := range s.certChain {
		certBytes = append(certBytes, chainCert.Raw...)
	}

	// Build SignedData
	sData := signedData{
		Version: 1,
		DigestAlgorithms: []pkixAlgorithmIdentifier{
			{Algorithm: oidSHA256},
		},
		ContentInfo: contentInfo{
			ContentType: oidData,
		},
		Certificates: asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        0,
			IsCompound: true,
			Bytes:      certBytes,
		},
		SignerInfos: []signerInfo{sInfo},
	}

	signedDataBytes, err := asn1.Marshal(sData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signedData: %w", err)
	}

	// Wrap in ContentInfo
	cInfo := contentInfo{
		ContentType: oidSignedData,
		Content: asn1.RawValue{
			Class:      asn1.ClassContextSpecific,
			Tag:        0,
			IsCompound: true,
			Bytes:      signedDataBytes,
		},
	}

	return asn1.Marshal(cInfo)
}

// ASN.1 structures for PKCS#7

type contentInfo struct {
	ContentType asn1.ObjectIdentifier
	Content     asn1.RawValue `asn1:"explicit,optional,tag:0"`
}

type signedData struct {
	Version          int
	DigestAlgorithms []pkixAlgorithmIdentifier `asn1:"set"`
	ContentInfo      contentInfo
	Certificates     asn1.RawValue `asn1:"optional,tag:0"`
	CRLs             asn1.RawValue `asn1:"optional,tag:1"`
	SignerInfos      []signerInfo  `asn1:"set"`
}

type signerInfo struct {
	Version                   int
	IssuerAndSerial           issuerAndSerial
	DigestAlgorithm           pkixAlgorithmIdentifier
	AuthenticatedAttributes   asn1.RawValue `asn1:"optional,tag:0"`
	DigestEncryptionAlgorithm pkixAlgorithmIdentifier
	EncryptedDigest           []byte
	UnauthenticatedAttributes asn1.RawValue `asn1:"optional,tag:1"`
}

type issuerAndSerial struct {
	Issuer       asn1.RawValue
	SerialNumber *big.Int
}

type pkixAlgorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `asn1:"optional"`
}

type attribute struct {
	Type  asn1.ObjectIdentifier
	Value asn1.RawValue `asn1:"set"`
}

func mustMarshal(v interface{}) []byte {
	b, err := asn1.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// GetAcroFormSigFlags returns the SigFlags value for AcroForm when signatures are present
func GetAcroFormSigFlags() int {
	// SigFlags value 3 means:
	// Bit 1 (value 1): SignaturesExist - The document contains at least one signature field
	// Bit 2 (value 2): AppendOnly - The document should be saved incrementally
	return 3
}

// UpdatePDFWithSignature updates the PDF buffer with the actual signature
// Returns the final PDF bytes with embedded signature
func UpdatePDFWithSignature(pdfData []byte, signer *PDFSigner) ([]byte, error) {
	// Find ByteRange placeholder: /ByteRange [0 0000000000 0000000000 0000000000]
	byteRangeMarker := []byte("/ByteRange [0 0000000000 0000000000 0000000000]")
	byteRangePos := bytes.Index(pdfData, byteRangeMarker)
	if byteRangePos < 0 {
		return pdfData, fmt.Errorf("byteRange placeholder not found")
	}

	// Find Contents placeholder in signature dictionary
	// Look for the specific pattern that starts with zeros (our placeholder)
	contentsMarker := []byte("/Contents <" + strings.Repeat("0", 100)) // First 100 zeros as marker
	contentsPos := bytes.Index(pdfData, contentsMarker)
	if contentsPos < 0 {
		// Fallback to simpler search
		contentsMarker = []byte("/Contents <")
		contentsPos = bytes.Index(pdfData, contentsMarker)
		if contentsPos < 0 {
			return pdfData, fmt.Errorf("contents placeholder not found")
		}
	}

	// Find the end of Contents (closing >)
	contentsStart := contentsPos + len("/Contents <")
	contentsEnd := bytes.Index(pdfData[contentsStart:], []byte(">"))
	if contentsEnd < 0 {
		return pdfData, fmt.Errorf("contents end not found")
	}
	contentsEnd += contentsStart

	// Validate the placeholder size
	placeholderSize := contentsEnd - contentsStart
	if placeholderSize != 16384 {
		return pdfData, fmt.Errorf("contents placeholder has unexpected size: %d (expected 16384)", placeholderSize)
	}

	// Calculate byte ranges
	// ByteRange format: [offset1, length1, offset2, length2]
	// offset1 = start of first range (always 0)
	// length1 = bytes from start to just before '<' of Contents
	// offset2 = byte after '>' of Contents
	// length2 = bytes from after Contents to end of file
	beforeContents := contentsStart - 1 // Position of '<'
	afterContents := contentsEnd + 1    // Position after '>'
	totalLength := len(pdfData)

	byteRange := [4]int{0, beforeContents, afterContents, totalLength - afterContents}

	// Update ByteRange in PDF
	newByteRange := fmt.Sprintf("/ByteRange [0 %010d %010d %010d]",
		byteRange[1], byteRange[2], byteRange[3])

	// Validate new ByteRange has same length as placeholder
	if len(newByteRange) != len(byteRangeMarker) {
		return pdfData, fmt.Errorf("ByteRange length mismatch: new=%d, placeholder=%d", len(newByteRange), len(byteRangeMarker))
	}

	// Create a copy of pdfData to modify
	result := make([]byte, len(pdfData))
	copy(result, pdfData)

	// Replace ByteRange
	copy(result[byteRangePos:byteRangePos+len(byteRangeMarker)], []byte(newByteRange))

	// Generate signature over the byte ranges (excluding Contents value)
	signature, err := signer.SignPDF(result, byteRange)
	if err != nil {
		return pdfData, fmt.Errorf("failed to sign PDF: %w", err)
	}

	// Convert signature to hex (uppercase to match PDF convention)
	sigHex := strings.ToUpper(hex.EncodeToString(signature))

	// Pad to fill the placeholder (16384 chars)
	if len(sigHex) > 16384 {
		return pdfData, fmt.Errorf("signature too large: %d bytes (max 8192)", len(sigHex)/2)
	}
	sigHex = sigHex + strings.Repeat("0", 16384-len(sigHex))

	// Replace Contents value
	copy(result[contentsStart:contentsEnd], []byte(sigHex))

	return result, nil
}
