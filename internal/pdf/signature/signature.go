// Package signature provides digital signature support for PDF documents.
package signature

import (
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

// PDFSigner handles digital signatures for PDF documents
type PDFSigner struct {
	config      *models.SignatureConfig
	certificate *x509.Certificate
	privateKey  crypto.PrivateKey
	certChain   []*x509.Certificate
	// Precomputed for PKCS#7 assembly (static per signer config).
	precomputedCertBytes []byte
	issuerAndSerial      issuerAndSerial
	digestSigAlgorithm   asn1.ObjectIdentifier
	derParts             pkcs7DERParts
	cachedPlacement      signaturePlacement
	hasCachedPlacement   bool
}

var pdfSignerCache sync.Map // signerPEMCacheKey -> *PDFSigner

var (
	marshaledOIDData        = mustMarshal(oidData)
	sha256HasherPool        sync.Pool
	authAttrsBytesPool      sync.Pool
	pkcs7MarshalBuffersPool sync.Pool
	signWorkerSlots         = make(chan struct{}, maxSignWorkers())
	hexUpperDigits          = []byte("0123456789ABCDEF")
)

type pkcs7MarshalBuffers struct{}

func maxSignWorkers() int {
	n := runtime.NumCPU() * 2
	if n < 8 {
		return 8
	}
	return n
}

func init() {
	sha256HasherPool.New = func() any { return sha256.New() }
	authAttrsBytesPool.New = func() any {
		b := make([]byte, 0, 256)
		return &b
	}
	pkcs7MarshalBuffersPool.New = func() any {
		return &pkcs7MarshalBuffers{}
	}
}

func digestByteRanges(pdfData []byte, byteRange [4]int) []byte {
	h := sha256HasherPool.Get().(hash.Hash)
	h.Reset()
	_, _ = h.Write(pdfData[byteRange[0]:byteRange[1]])
	_, _ = h.Write(pdfData[byteRange[2] : byteRange[2]+byteRange[3]])
	sum := h.Sum(nil)
	digest := make([]byte, len(sum))
	copy(digest, sum)
	sha256HasherPool.Put(h)
	return digest
}

// SignatureIDs holds the object IDs for a signature field and its associated annotations.
//
//nolint:revive // exported
type SignatureIDs struct {
	SigFieldID           int
	SigAnnotID           int
	SigValueID           int
	AppearanceID         int
	ByteRangeRel         int // ByteRange placeholder offset inside sig value object body
	ContentsDataStartRel int // hex payload start inside sig value object body
	ContentsDataEndRel   int // hex payload end inside sig value object body
}

const (
	sigByteRangePlaceholder = "/ByteRange [0 0000000000 0000000000 0000000000]"
	sigContentsHexLen       = 16384
)

// OID values for CMS/PKCS#7
var (
	oidData            = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 1}
	oidSignedData      = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 7, 2}
	oidSHA256          = asn1.ObjectIdentifier{2, 16, 840, 1, 101, 3, 4, 2, 1}
	oidRSAEncryption   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	oidECDSAWithSHA256 = asn1.ObjectIdentifier{1, 2, 840, 10045, 4, 3, 2}
	oidContentType     = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 3}
	oidMessageDigest   = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 4}
	oidSigningTime     = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 9, 5}
)

// parsedSignerPEMEntry caches certificate, private key (e.g. *rsa.PrivateKey), and chain for a PEM fingerprint.
type parsedSignerPEMEntry struct {
	cert      *x509.Certificate
	key       crypto.PrivateKey
	certChain []*x509.Certificate
}

var signerPEMMaterialCache sync.Map // hex(sha256(...)) -> *parsedSignerPEMEntry

func signerPEMCacheKey(certPEM, keyPEM string, chain []string) string {
	h := sha256.New()
	h.Write([]byte(certPEM))
	h.Write([]byte{0})
	h.Write([]byte(keyPEM))
	for _, c := range chain {
		h.Write([]byte{1})
		h.Write([]byte(c))
	}
	return hex.EncodeToString(h.Sum(nil))
}

func parseSignerPEMMaterials(certPEM, keyPEM string, chainPEMs []string) (*x509.Certificate, crypto.PrivateKey, []*x509.Certificate, error) {
	cacheKey := signerPEMCacheKey(certPEM, keyPEM, chainPEMs)
	if v, ok := signerPEMMaterialCache.Load(cacheKey); ok {
		ent := v.(*parsedSignerPEMEntry)
		return ent.cert, ent.key, ent.certChain, nil
	}

	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, nil, nil, errors.New("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	keyBlock, _ := pem.Decode([]byte(keyPEM))
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
				return nil, nil, nil, fmt.Errorf("failed to parse private key: %w", err)
			}
		}
	}

	var chain []*x509.Certificate
	for _, chainPEM := range chainPEMs {
		chainBlock, _ := pem.Decode([]byte(chainPEM))
		if chainBlock == nil {
			continue
		}
		chainCert, perr := x509.ParseCertificate(chainBlock.Bytes)
		if perr == nil {
			chain = append(chain, chainCert)
		}
	}

	signerPEMMaterialCache.Store(cacheKey, &parsedSignerPEMEntry{
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

	cacheKey := signerPEMCacheKey(config.CertificatePEM, config.PrivateKeyPEM, config.CertificateChain)
	if v, ok := pdfSignerCache.Load(cacheKey); ok {
		signer := v.(*PDFSigner)
		signer.hasCachedPlacement = false
		return signer, nil
	}

	cert, privateKey, chain, err := parseSignerPEMMaterials(
		config.CertificatePEM,
		config.PrivateKeyPEM,
		config.CertificateChain,
	)
	if err != nil {
		return nil, err
	}

	digestSigAlgorithm := oidRSAEncryption
	switch privateKey.(type) {
	case *ecdsa.PrivateKey:
		digestSigAlgorithm = oidECDSAWithSHA256
	case *rsa.PrivateKey:
	default:
		return nil, errors.New("unsupported private key type (use RSA or ECDSA P-256)")
	}

	signer := &PDFSigner{
		config:      config,
		certificate: cert,
		privateKey:  privateKey,
		certChain:   chain,
		issuerAndSerial: issuerAndSerial{
			Issuer:       asn1.RawValue{FullBytes: cert.RawIssuer},
			SerialNumber: cert.SerialNumber,
		},
		digestSigAlgorithm: digestSigAlgorithm,
	}
	signer.precomputedCertBytes = make([]byte, 0, len(cert.Raw)+len(chain)*512)
	signer.precomputedCertBytes = append(signer.precomputedCertBytes, cert.Raw...)
	for _, chainCert := range chain {
		signer.precomputedCertBytes = append(signer.precomputedCertBytes, chainCert.Raw...)
	}
	signer.derParts = buildPKCS7DERParts(signer)

	if existing, loaded := pdfSignerCache.LoadOrStore(cacheKey, signer); loaded {
		cached := existing.(*PDFSigner)
		cached.hasCachedPlacement = false
		return cached, nil
	}
	return signer, nil
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

	// Create appearance stream for visible signature
	if s.config.Visible {
		ids.AppearanceID = s.createSignatureAppearance(pageManager, sigW, sigH, fontID)
	}

	// Create signature value dictionary (will be filled during signing)
	sigValueID := pageManager.AllocObjectID()

	signerName := s.config.Name
	if signerName == "" && s.certificate != nil {
		signerName = s.certificate.Subject.CommonName
	}

	// Build signature value dictionary
	var sigValueDict bytes.Buffer
	sigValueDict.Grow(17000)
	sigValueDict.WriteString("<< /Type /Sig")
	sigValueDict.WriteString(" /Filter /Adobe.PPKLite")
	sigValueDict.WriteString(" /SubFilter /adbe.pkcs7.detached")

	// ByteRange placeholder - will be replaced during actual signing
	// Format: [0 offset1 offset2 length] where offset1 is start of /Contents, offset2 is end of /Contents
	sigValueDict.WriteByte(' ')
	ids.ByteRangeRel = sigValueDict.Len()
	sigValueDict.WriteString(sigByteRangePlaceholder)

	// Contents placeholder - hex-encoded PKCS#7 signature
	// Reserve space for signature (8192 bytes = 16384 hex chars)
	sigValueDict.WriteString(" /Contents <")
	ids.ContentsDataStartRel = sigValueDict.Len()
	for range sigContentsHexLen {
		sigValueDict.WriteByte('0')
	}
	ids.ContentsDataEndRel = sigValueDict.Len()
	sigValueDict.WriteString(">")

	if s.config.Reason != "" {
		sigValueDict.WriteString(" /Reason (" + escapeText(s.config.Reason) + ")")
	}
	if s.config.Location != "" {
		sigValueDict.WriteString(" /Location (" + escapeText(s.config.Location) + ")")
	}
	if s.config.ContactInfo != "" {
		sigValueDict.WriteString(" /ContactInfo (" + escapeText(s.config.ContactInfo) + ")")
	}
	if signerName != "" {
		sigValueDict.WriteString(" /Name (" + escapeText(signerName) + ")")
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
	tzHoursStr := strconv.Itoa(tzHours)
	if len(tzHoursStr) == 1 {
		tzHoursStr = "0" + tzHoursStr
	}
	tzMinsStr := strconv.Itoa(tzMinutes)
	if len(tzMinsStr) == 1 {
		tzMinsStr = "0" + tzMinsStr
	}
	sigValueDict.WriteString(" /M (D:")
	sigValueDict.WriteString(now.Format("20060102150405"))
	sigValueDict.WriteString(tzSign)
	sigValueDict.WriteString(tzHoursStr)
	sigValueDict.WriteByte('\'')
	sigValueDict.WriteString(tzMinsStr)
	sigValueDict.WriteString("')")

	sigValueDict.WriteString(" >>")

	pageManager.SetExtraObjectBytes(sigValueID, sigValueDict.Bytes())
	ids.SigValueID = sigValueID

	// Create signature field widget annotation
	sigAnnotID := pageManager.AllocObjectID()
	ids.SigAnnotID = sigAnnotID

	var annotDict strings.Builder
	annotDict.Grow(512)
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

	pageManager.SetExtraObject(sigAnnotID, annotDict.String())

	// Add annotation to the appropriate page
	pageIndex := max(targetPage-1, 0)
	pageManager.AppendPageAnnot(pageIndex, sigAnnotID)

	ids.SigFieldID = sigAnnotID // In this implementation, field = annotation

	return ids
}

// createSignatureAppearance creates the visual appearance for a visible signature
func (s *PDFSigner) createSignatureAppearance(pageManager SignaturePageContext, width, height float64, fontID int) int {
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
	appearance.WriteString(fmt.Sprintf("5 %s Td\n", fmtNum(height-15)))
	appearance.WriteString(fmt.Sprintf("%s Tj\n", formatText("Digitally signed by:")))

	// Mark font usage for subsetting
	if useHexEncoding {
		pageManager.FontMarkChars("Helvetica", "Digitally signed by:")
	}

	// Signer name
	appearance.WriteString("0 -12 Td\n")
	appearance.WriteString(fmt.Sprintf("%s Tj\n", formatText(signerName)))
	if useHexEncoding {
		pageManager.FontMarkChars("Helvetica", signerName)
	}

	// Date
	now := time.Now()
	dateStr := "Date: " + now.Format("2006-01-02 15:04:05")
	appearance.WriteString("0 -12 Td\n")
	appearance.WriteString(fmt.Sprintf("%s Tj\n", formatText(dateStr)))
	if useHexEncoding {
		pageManager.FontMarkChars("Helvetica", dateStr)
	}

	// Reason if provided
	if s.config.Reason != "" {
		reasonStr := "Reason: " + s.config.Reason
		appearance.WriteString("0 -12 Td\n")
		appearance.WriteString(fmt.Sprintf("%s Tj\n", formatText(reasonStr)))
		if useHexEncoding {
			pageManager.FontMarkChars("Helvetica", reasonStr)
		}
	}

	// Location if provided
	if s.config.Location != "" {
		locationStr := "Location: " + s.config.Location
		appearance.WriteString("0 -12 Td\n")
		appearance.WriteString(fmt.Sprintf("%s Tj\n", formatText(locationStr)))
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
	if fontID > 0 {
		// Use reference to existing embedded font
		resourcesDict = fmt.Sprintf("<< /Font << /F1 %d 0 R >> >>", fontID)
	} else {
		// Fallback for non-embedded (should be avoided for PDF/A)
		resourcesDict = "<< /Font << /F1 << /Type /Font /Subtype /Type1 /BaseFont /Helvetica >> >> >>"
	}

	appearanceDict := fmt.Sprintf("<< /Type /XObject /Subtype /Form /BBox [0 0 %s %s] /Resources %s /Length %d >>\nstream\n%s\nendstream",
		fmtNum(width), fmtNum(height), resourcesDict, len(appearanceContent), appearanceContent)

	pageManager.SetExtraObject(appearanceID, appearanceDict)

	return appearanceID
}

// SignPDF signs the PDF data and returns the PKCS#7 signature
// This is called after the PDF is generated to compute the actual signature
func (s *PDFSigner) SignPDF(pdfData []byte, byteRange [4]int) ([]byte, error) {
	messageDigest := digestByteRanges(pdfData, byteRange)

	signWorkerSlots <- struct{}{}
	signedData, err := s.createPKCS7SignedData(messageDigest)
	<-signWorkerSlots
	if err != nil {
		return nil, err
	}
	return signedData, nil
}

// createPKCS7SignedData creates a PKCS#7 SignedData structure using hand-built DER (P42).
func (s *PDFSigner) createPKCS7SignedData(messageDigest []byte) ([]byte, error) {
	signingTime := time.Now().UTC()
	authAttrsBytes := buildAuthenticatedAttributesSET(s.derParts, messageDigest, signingTime)

	authAttrsHash := sha256.Sum256(authAttrsBytes)
	contentBytes := stripOuterTLV(authAttrsBytes)

	var signature []byte
	var err error
	switch key := s.privateKey.(type) {
	case *rsa.PrivateKey:
		signature, err = rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, authAttrsHash[:])
		if err != nil {
			return nil, fmt.Errorf("failed to sign: %w", err)
		}
	case *ecdsa.PrivateKey:
		signature, err = ecdsa.SignASN1(rand.Reader, key, authAttrsHash[:])
		if err != nil {
			return nil, fmt.Errorf("failed to sign: %w", err)
		}
	default:
		return nil, errors.New("unsupported key type")
	}

	signerInfo := buildSignerInfoDER(s.derParts, contentBytes, signature)
	signedData := buildSignedDataDER(s.derParts, signerInfo)
	return buildContentInfoDER(signedData), nil
}

type issuerAndSerial struct {
	Issuer       asn1.RawValue
	SerialNumber *big.Int
}

type pkixAlgorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `asn1:"optional"`
}

func mustMarshal(v any) []byte {
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

type signaturePlacement struct {
	byteRangePos    int
	byteRangeMarker []byte
	contentsStart   int
	contentsEnd     int
	byteRange       [4]int
}

func buildSignaturePlacement(objBodyStart int, ids *SignatureIDs, pdfLen int) signaturePlacement {
	sp := signaturePlacement{
		byteRangeMarker: []byte(sigByteRangePlaceholder),
		byteRangePos:    objBodyStart + ids.ByteRangeRel,
		contentsStart:   objBodyStart + ids.ContentsDataStartRel,
		contentsEnd:     objBodyStart + ids.ContentsDataEndRel,
	}
	beforeContents := sp.contentsStart - 1
	afterContents := sp.contentsEnd + 1
	sp.byteRange = [4]int{0, beforeContents, afterContents, pdfLen - afterContents}
	return sp
}

// CacheSignaturePlacement records absolute placeholder offsets while the sig value
// object is written, avoiding a full-PDF scan during signing (P35).
func (s *PDFSigner) CacheSignaturePlacement(objBodyStart int, ids *SignatureIDs) {
	if s == nil || ids == nil || ids.SigValueID == 0 {
		return
	}
	s.cachedPlacement = buildSignaturePlacement(objBodyStart, ids, 0)
	s.hasCachedPlacement = true
}

func (s *PDFSigner) signaturePlacementForPDF(pdfData []byte) (signaturePlacement, error) {
	if s != nil && s.hasCachedPlacement {
		sp := s.cachedPlacement
		beforeContents := sp.contentsStart - 1
		afterContents := sp.contentsEnd + 1
		sp.byteRange = [4]int{0, beforeContents, afterContents, len(pdfData) - afterContents}
		return sp, nil
	}
	return locateSignaturePlacement(pdfData)
}

func locateSignaturePlacement(pdfData []byte) (signaturePlacement, error) {
	var sp signaturePlacement
	sp.byteRangeMarker = []byte(sigByteRangePlaceholder)
	sp.byteRangePos = bytes.Index(pdfData, sp.byteRangeMarker)
	if sp.byteRangePos < 0 {
		return sp, errors.New("byteRange placeholder not found")
	}

	contentsMarker := []byte("/Contents <" + strings.Repeat("0", 100))
	contentsPos := bytes.Index(pdfData, contentsMarker)
	if contentsPos < 0 {
		contentsMarker = []byte("/Contents <")
		contentsPos = bytes.Index(pdfData, contentsMarker)
		if contentsPos < 0 {
			return sp, fmt.Errorf("contents placeholder not found")
		}
	}

	sp.contentsStart = contentsPos + len("/Contents <")
	contentsEndRel := bytes.Index(pdfData[sp.contentsStart:], []byte(">"))
	if contentsEndRel < 0 {
		return sp, fmt.Errorf("contents end not found")
	}
	sp.contentsEnd = sp.contentsStart + contentsEndRel

	placeholderSize := sp.contentsEnd - sp.contentsStart
	if placeholderSize != sigContentsHexLen {
		return sp, fmt.Errorf("contents placeholder has unexpected size: %d (expected %d)", placeholderSize, sigContentsHexLen)
	}

	beforeContents := sp.contentsStart - 1
	afterContents := sp.contentsEnd + 1
	totalLength := len(pdfData)
	sp.byteRange = [4]int{0, beforeContents, afterContents, totalLength - afterContents}
	return sp, nil
}

// UpdatePDFWithSignature updates PDF bytes with the embedded PKCS#7 signature.
func UpdatePDFWithSignature(pdfData []byte, signer *PDFSigner) ([]byte, error) {
	sp, err := locateSignaturePlacement(pdfData)
	if err != nil {
		return pdfData, err
	}

	result := make([]byte, len(pdfData))
	copy(result, pdfData)
	if err := embedSignatureInPlace(result, signer, sp); err != nil {
		return pdfData, err
	}
	return result, nil
}

// UpdatePDFWithSignatureBuffer embeds the signature directly into pdfBuf without allocating a second PDF copy.
func UpdatePDFWithSignatureBuffer(pdfBuf *bytes.Buffer, signer *PDFSigner) error {
	sp, err := signer.signaturePlacementForPDF(pdfBuf.Bytes())
	if err != nil {
		return err
	}
	return embedSignatureInPlace(pdfBuf.Bytes(), signer, sp)
}

func appendByteRangeMarker(dst []byte, byteRange [4]int) []byte {
	dst = append(dst, "/ByteRange [0 "...)
	dst = appendPaddedInt(dst, byteRange[1], 10)
	dst = append(dst, ' ')
	dst = appendPaddedInt(dst, byteRange[2], 10)
	dst = append(dst, ' ')
	dst = appendPaddedInt(dst, byteRange[3], 10)
	return append(dst, ']')
}

func appendPaddedInt(dst []byte, n int, width int) []byte {
	var tmp [16]byte
	b := strconv.AppendInt(tmp[:0], int64(n), 10)
	padding := width - len(b)
	if padding > 0 {
		for range padding {
			dst = append(dst, '0')
		}
	}
	return append(dst, b...)
}

func encodeHexUpper(dst, src []byte) {
	for i, b := range src {
		dst[i*2] = hexUpperDigits[b>>4]
		dst[i*2+1] = hexUpperDigits[b&0x0f]
	}
}

func embedSignatureInPlace(pdfData []byte, signer *PDFSigner, sp signaturePlacement) error {
	var byteRangeScratch [48]byte
	newByteRange := appendByteRangeMarker(byteRangeScratch[:0], sp.byteRange)
	if len(newByteRange) != len(sp.byteRangeMarker) {
		return fmt.Errorf("ByteRange length mismatch: new=%d, placeholder=%d", len(newByteRange), len(sp.byteRangeMarker))
	}

	copy(pdfData[sp.byteRangePos:sp.byteRangePos+len(sp.byteRangeMarker)], newByteRange)

	signature, err := signer.SignPDF(pdfData, sp.byteRange)
	if err != nil {
		return fmt.Errorf("failed to sign PDF: %w", err)
	}

	contents := pdfData[sp.contentsStart:sp.contentsEnd]
	if len(signature)*2 > len(contents) {
		return fmt.Errorf("signature too large: %d bytes (max %d)", len(signature), len(contents)/2)
	}
	encodeHexUpper(contents[:len(signature)*2], signature)
	for i := len(signature) * 2; i < len(contents); i++ {
		contents[i] = '0'
	}
	return nil
}
