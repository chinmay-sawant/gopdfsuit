// Package signature provides digital signature support for PDF documents.
package signature

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
	"errors"
	"math/big"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/byteconv"
	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

// PDFSigner handles digital signatures for PDF documents
type PDFSigner struct {
	config      *models.SignatureConfig
	certificate *x509.Certificate
	privateKey  crypto.PrivateKey
	certChain   []*x509.Certificate
}

// SignatureIDs holds the object IDs for a signature field and its associated annotations.
//
//nolint:revive // exported
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

// parsedSignerPEMEntry caches certificate, private key (e.g. *rsa.PrivateKey), and chain for a PEM fingerprint.
type parsedSignerPEMEntry struct {
	cert      *x509.Certificate
	key       crypto.PrivateKey
	certChain []*x509.Certificate
}

const maxSignerPEMCacheSize = 64

var (
	signerPEMMaterialCache sync.Map // hex(sha256(...)) -> *parsedSignerPEMEntry
	signerPEMCacheSize     atomic.Int64
	signerPEMCacheEvictMu  sync.Mutex
)

// padZeros holds pre-filled zero bytes for appendPad10 (PERF-119: single append instead of loop).
var padZeros = [10]byte{'0', '0', '0', '0', '0', '0', '0', '0', '0', '0'}

// appendPad10 appends a zero-padded 10-digit decimal (PERF-35/119/226: no fmt, no make+copy).
func appendPad10(dst []byte, n int) []byte {
	var buf [20]byte
	num := strconv.AppendInt(buf[:0], int64(n), 10)
	pad := 10 - len(num)
	if pad < 0 {
		pad = 0
	}
	return append(append(dst, padZeros[:pad]...), num...)
}

func signerPEMCacheKey(certPEMBytes, keyPEMBytes []byte, chain []string) string {
	h := sha256.New()
	h.Write(certPEMBytes)
	h.Write([]byte{0})
	h.Write(keyPEMBytes)
	for _, c := range chain {
		h.Write([]byte{1})
		h.Write(byteconv.StringToBytes(c))
	}
	var k [32]byte
	h.Sum(k[:0])
	return string(k[:])
}

func storeSignerPEMCache(key string, ent *parsedSignerPEMEntry) {
	if _, loaded := signerPEMMaterialCache.LoadOrStore(key, ent); loaded {
		return
	}
	n := signerPEMCacheSize.Add(1)
	if n <= maxSignerPEMCacheSize {
		return
	}
	// Best-effort eviction: drop one arbitrary entry when over capacity.
	// Explicit unlock (PERF-31) — avoid defer on the hot store path.
	signerPEMCacheEvictMu.Lock()
	if signerPEMCacheSize.Load() <= maxSignerPEMCacheSize {
		signerPEMCacheEvictMu.Unlock()
		return
	}
	signerPEMMaterialCache.Range(func(k, _ any) bool {
		if ks, ok := k.(string); ok && ks != key {
			signerPEMMaterialCache.Delete(ks)
			signerPEMCacheSize.Add(-1)
			return false
		}
		return true
	})
	signerPEMCacheEvictMu.Unlock()
}

func parseSignerPEMMaterials(certPEM, keyPEM string, chainPEMs []string) (*x509.Certificate, crypto.PrivateKey, []*x509.Certificate, error) {
	certPEMBytes := byteconv.StringToBytes(certPEM)
	keyPEMBytes := byteconv.StringToBytes(keyPEM)

	cacheKey := signerPEMCacheKey(certPEMBytes, keyPEMBytes, chainPEMs)
	if v, ok := signerPEMMaterialCache.Load(cacheKey); ok {
		ent := v.(*parsedSignerPEMEntry)
		return ent.cert, ent.key, ent.certChain, nil
	}

	// PERF-231: PEM/key parsing only reached on cache miss (line 130 checks cache first)
	block, _ := pem.Decode(certPEMBytes)
	if block == nil {
		return nil, nil, nil, errors.New("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, nil, errors.Join(errors.New("failed to parse certificate"), err)
	}

	keyBlock, _ := pem.Decode(keyPEMBytes)
	if keyBlock == nil {
		return nil, nil, nil, errors.New("failed to parse private key PEM")
	}

	var privateKey crypto.PrivateKey
	privateKey, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
		if err != nil {
			privateKey, err = x509.ParseECPrivateKey(keyBlock.Bytes)
			if err != nil {
				return nil, nil, nil, errors.Join(errors.New("failed to parse private key"), err)
			}
		}
	}

	var chain []*x509.Certificate
	for _, chainPEM := range chainPEMs {
		chainBlock, _ := pem.Decode(byteconv.StringToBytes(chainPEM))
		if chainBlock == nil {
			continue
		}
		chainCert, perr := x509.ParseCertificate(chainBlock.Bytes)
		if perr == nil {
			chain = append(chain, chainCert)
		}
	}

	storeSignerPEMCache(cacheKey, &parsedSignerPEMEntry{
		cert:      cert,
		key:       privateKey,
		certChain: chain,
	})
	return cert, privateKey, chain, nil
}

// NewPDFSigner creates a new PDF signer from config
func NewPDFSigner(config *models.SignatureConfig) (*PDFSigner, error) {
	if config == nil || !config.Enabled {
		return nil, nil
	}

	cert, privateKey, chain, err := parseSignerPEMMaterials(
		config.CertificatePEM,
		config.PrivateKeyPEM,
		config.CertificateChain,
	)
	if err != nil {
		return nil, err
	}

	return &PDFSigner{
		config:      config,
		certificate: cert,
		privateKey:  privateKey,
		certChain:   chain,
	}, nil
}

// CreateSignatureField creates the signature field and annotation objects
func (s *PDFSigner) CreateSignatureField(pageManager SignaturePageContext, pageDims PageDimensions, fontID int) *SignatureIDs {
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
		sigX = pageDims.Width - sigW - pageManager.GetMargins().Right
	}
	if sigY <= 0 {
		sigY = pageManager.GetMargins().Bottom
	}

	// Get current time once for both appearance and signing time (PERF-40)
	now := time.Now()
	if s.config.Visible {
		ids.AppearanceID = s.createSignatureAppearance(pageManager, sigW, sigH, fontID, now)
	}

	// Create signature value dictionary (will be filled during signing)
	sigValueID := pageManager.AllocObjectID()

	signerName := s.config.Name
	if signerName == "" && s.certificate != nil {
		signerName = s.certificate.Subject.CommonName
	}

	// Build signature value dictionary
	var sigValueDict strings.Builder
	sigValueDict.Grow(16500)
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

	// PERF-35: no fmt.Sprintf interface boxing
	if s.config.Reason != "" {
		sigValueDict.WriteString(" /Reason (")
		sigValueDict.WriteString(escapeText(s.config.Reason))
		sigValueDict.WriteByte(')')
	}
	if s.config.Location != "" {
		sigValueDict.WriteString(" /Location (")
		sigValueDict.WriteString(escapeText(s.config.Location))
		sigValueDict.WriteByte(')')
	}
	if s.config.ContactInfo != "" {
		sigValueDict.WriteString(" /ContactInfo (")
		sigValueDict.WriteString(escapeText(s.config.ContactInfo))
		sigValueDict.WriteByte(')')
	}
	if signerName != "" {
		sigValueDict.WriteString(" /Name (")
		sigValueDict.WriteString(escapeText(signerName))
		sigValueDict.WriteByte(')')
	}

	// Signing time - PDF date format: D:YYYYMMDDHHmmSSOHH'mm'
	// where O is + or -, HH is timezone hours, mm is timezone minutes
	_, tzOffset := now.Zone()
	tzSign := "+"
	if tzOffset < 0 {
		tzSign = "-"
		tzOffset = -tzOffset
	}
	tzHours := tzOffset / 3600
	tzMinutes := (tzOffset % 3600) / 60
	// PERF-35: build PDF date without fmt.Sprintf
	sigValueDict.WriteString(" /M (D:")
	sigValueDict.WriteString(now.Format("20060102150405"))
	sigValueDict.WriteString(tzSign)
	var mTmp [4]byte
	if tzHours < 10 {
		sigValueDict.WriteByte('0')
	}
	sigValueDict.Write(strconv.AppendInt(mTmp[:0], int64(tzHours), 10))
	sigValueDict.WriteByte('\'')
	if tzMinutes < 10 {
		sigValueDict.WriteByte('0')
	}
	sigValueDict.Write(strconv.AppendInt(mTmp[:0], int64(tzMinutes), 10))
	sigValueDict.WriteString("')")

	sigValueDict.WriteString(" >>")

	pageManager.SetExtraObject(sigValueID, sigValueDict.String())

	// Create signature field widget annotation
	sigAnnotID := pageManager.AllocObjectID()
	ids.SigAnnotID = sigAnnotID

	var annotDict strings.Builder
	annotDict.Grow(320)
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
		annotDict.WriteString(" /Rect [")
		annotDict.WriteString(fmtNum(sigX))
		annotDict.WriteByte(' ')
		annotDict.WriteString(fmtNum(sigY))
		annotDict.WriteByte(' ')
		annotDict.WriteString(fmtNum(sigX + sigW))
		annotDict.WriteByte(' ')
		annotDict.WriteString(fmtNum(sigY + sigH))
		annotDict.WriteByte(']')
		if ids.AppearanceID > 0 {
			annotDict.WriteString(" /AP << /N ")
			var idTmp [12]byte
			annotDict.Write(strconv.AppendInt(idTmp[:0], int64(ids.AppearanceID), 10))
			annotDict.WriteString(" 0 R >>")
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
	annotDict.WriteString(" /P ")
	var pTmp [12]byte
	annotDict.Write(strconv.AppendInt(pTmp[:0], int64(pageObjID), 10))
	annotDict.WriteString(" 0 R")

	annotDict.WriteString(" >>")

	pageManager.SetExtraObject(sigAnnotID, annotDict.String())

	// Add annotation to the appropriate page
	pageIndex := targetPage - 1
	if pageIndex < 0 {
		pageIndex = 0
	}
	pageManager.AppendPageAnnot(pageIndex, sigAnnotID)

	ids.SigFieldID = sigAnnotID // In this implementation, field = annotation

	return ids
}

// createSignatureAppearance creates the visual appearance for a visible signature
func (s *PDFSigner) createSignatureAppearance(pageManager SignaturePageContext, width, height float64, fontID int, now time.Time) int {
	var appearance strings.Builder
	appearance.Grow(512)

	// Yellow background with black border
	appearance.WriteString("q\n")
	appearance.WriteString("1 1 0.8 rg\n") // Light yellow background (RGB: 255, 255, 204)
	appearance.WriteString("0 0 ")
	appearance.WriteString(fmtNum(width))
	appearance.WriteByte(' ')
	appearance.WriteString(fmtNum(height))
	appearance.WriteString(" re f\n")
	appearance.WriteString("0 0 0 RG 1 w\n") // Black border
	appearance.WriteString("0 0 ")
	appearance.WriteString(fmtNum(width))
	appearance.WriteByte(' ')
	appearance.WriteString(fmtNum(height))
	appearance.WriteString(" re S\n")
	appearance.WriteString("Q\n")

	// Text content
	signerName := s.config.Name
	if signerName == "" && s.certificate != nil {
		signerName = s.certificate.Subject.CommonName
	}

	// Check if we're using a custom font (Liberation) that needs hex encoding
	useHexEncoding := fontID > 0 && pageManager.FontHas("Helvetica")

	appearance.WriteString("BT\n")
	appearance.WriteString("/F1 9 Tf\n")
	appearance.WriteString("0 0 0 rg\n")

	// Helper to format text based on font type
	formatText := func(text string) string {
		if useHexEncoding {
			// For Liberation fonts, use hex encoding
			return pageManager.EncodeTextForFont("Helvetica", text)
		}
		// For standard fonts, use ASCII encoding
		return "(" + escapeText(text) + ")"
	}

	// "Digitally signed by" line
	appearance.WriteString("5 ")
	appearance.WriteString(fmtNum(height - 15))
	appearance.WriteString(" Td\n")
	appearance.WriteString(formatText("Digitally signed by:"))
	appearance.WriteString(" Tj\n")

	// Mark font usage for subsetting
	if useHexEncoding {
		pageManager.FontMarkChars("Helvetica", "Digitally signed by:")
	}

	// Signer name
	appearance.WriteString("0 -12 Td\n")
	appearance.WriteString(formatText(signerName))
	appearance.WriteString(" Tj\n")
	if useHexEncoding {
		pageManager.FontMarkChars("Helvetica", signerName)
	}

	// Date
	dateStr := "Date: " + now.Format("2006-01-02 15:04:05")
	appearance.WriteString("0 -12 Td\n")
	appearance.WriteString(formatText(dateStr))
	appearance.WriteString(" Tj\n")
	if useHexEncoding {
		pageManager.FontMarkChars("Helvetica", dateStr)
	}

	// Reason if provided
	if s.config.Reason != "" {
		reasonStr := "Reason: " + s.config.Reason
		appearance.WriteString("0 -12 Td\n")
		appearance.WriteString(formatText(reasonStr))
		appearance.WriteString(" Tj\n")
		if useHexEncoding {
			pageManager.FontMarkChars("Helvetica", reasonStr)
		}
	}

	// Location if provided
	if s.config.Location != "" {
		locationStr := "Location: " + s.config.Location
		appearance.WriteString("0 -12 Td\n")
		appearance.WriteString(formatText(locationStr))
		appearance.WriteString(" Tj\n")
		if useHexEncoding {
			pageManager.FontMarkChars("Helvetica", locationStr)
		}
	}

	appearance.WriteString("ET\n")

	appearanceContent := appearance.String()

	// Create appearance XObject
	appearanceID := pageManager.AllocObjectID()

	// Construct resources dictionary using the embedded font ID
	var resourcesDict string
	var resB strings.Builder
	resB.Grow(64)
	if fontID > 0 {
		// Use reference to existing embedded font
		resB.WriteString("<< /Font << /F1 ")
		var fTmp [12]byte
		resB.Write(strconv.AppendInt(fTmp[:0], int64(fontID), 10))
		resB.WriteString(" 0 R >> >>")
	} else {
		// Fallback for non-embedded (should be avoided for PDF/A)
		resB.WriteString("<< /Font << /F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> >> >>")
	}
	resourcesDict = resB.String()

	var appB strings.Builder
	appB.Grow(96 + len(resourcesDict) + len(appearanceContent))
	appB.WriteString("<< /Type /XObject /Subtype /Form /BBox [0 0 ")
	appB.WriteString(fmtNum(width))
	appB.WriteByte(' ')
	appB.WriteString(fmtNum(height))
	appB.WriteString("] /Resources ")
	appB.WriteString(resourcesDict)
	appB.WriteString(" /Length ")
	var lTmp [12]byte
	appB.Write(strconv.AppendInt(lTmp[:0], int64(len(appearanceContent)), 10))
	appB.WriteString(" >>\nstream\n")
	appB.WriteString(appearanceContent)
	appB.WriteString("\nendstream")

	pageManager.SetExtraObject(appearanceID, appB.String())

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
		return nil, errors.Join(errors.New("failed to marshal authenticated attributes"), err)
	}

	// Change SEQUENCE tag (0x30) to SET tag (0x31) — seqBytes is a fresh allocation from asn1.Marshal
	authAttrsBytes := seqBytes
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
			return nil, errors.Join(errors.New("failed to sign"), err)
		}
	default:
		return nil, errors.New("unsupported key type")
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
		return nil, errors.Join(errors.New("failed to marshal signedData"), err)
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
		return pdfData, errors.New("byteRange placeholder not found")
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
			return pdfData, errors.New("contents placeholder not found")
		}
	}

	// Find the end of Contents (closing >)
	contentsStart := contentsPos + len("/Contents <")
	contentsEnd := bytes.Index(pdfData[contentsStart:], []byte(">"))
	if contentsEnd < 0 {
		return pdfData, errors.New("contents end not found")
	}
	contentsEnd += contentsStart

	// Validate the placeholder size
	placeholderSize := contentsEnd - contentsStart
	if placeholderSize != 16384 {
		return pdfData, errors.New("contents placeholder has unexpected size: " + strconv.Itoa(placeholderSize) + " (expected 16384)")
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

	// Update ByteRange in PDF (PERF-35: fixed-width digits without fmt.Sprintf)
	// Placeholder is exactly: "/ByteRange [0 0000000000 0000000000 0000000000]"
	newByteRange := make([]byte, 0, len(byteRangeMarker))
	newByteRange = append(newByteRange, "/ByteRange [0 "...)
	newByteRange = appendPad10(newByteRange, byteRange[1])
	newByteRange = append(newByteRange, ' ')
	newByteRange = appendPad10(newByteRange, byteRange[2])
	newByteRange = append(newByteRange, ' ')
	newByteRange = appendPad10(newByteRange, byteRange[3])
	newByteRange = append(newByteRange, ']')

	// Validate new ByteRange has same length as placeholder
	if len(newByteRange) != len(byteRangeMarker) {
		return pdfData, errors.New("ByteRange length mismatch: new=" + strconv.Itoa(len(newByteRange)) +
			", placeholder=" + strconv.Itoa(len(byteRangeMarker)))
	}

	// Create a copy of pdfData to modify
	result := bytes.Clone(pdfData)

	// Replace ByteRange
	copy(result[byteRangePos:byteRangePos+len(byteRangeMarker)], newByteRange)

	// Generate signature over the byte ranges (excluding Contents value)
	signature, err := signer.SignPDF(result, byteRange)
	if err != nil {
		return pdfData, errors.Join(errors.New("failed to sign PDF"), err)
	}

	// Convert signature to hex (uppercase to match PDF convention)
	sigHex := strings.ToUpper(hex.EncodeToString(signature))

	// Pad to fill the placeholder (16384 chars)
	if len(sigHex) > 16384 {
		return pdfData, errors.New("signature too large: " + strconv.Itoa(len(sigHex)/2) + " bytes (max 8192)")
	}
	sigHex += strings.Repeat("0", 16384-len(sigHex))

	// Replace Contents value
	copy(result[contentsStart:contentsEnd], byteconv.StringToBytes(sigHex))

	return result, nil
}
