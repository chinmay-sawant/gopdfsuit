package handlers

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/compress"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/font"
	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
	"github.com/gin-gonic/gin"
)

const mimeTypePDF = "application/pdf"

const (
	// maxPDFBytes caps single PDF uploads (mirrors compress.MaxInputBytes).
	maxPDFBytes = int64(compress.MaxInputBytes)
	// maxXFDFBytes caps XFDF form-data uploads.
	maxXFDFBytes = 8 << 20
	// maxFontBytes caps custom font uploads.
	maxFontBytes = 10 << 20
	// maxHTMLBodyBytes caps HTML-to-PDF/Image JSON request bodies.
	maxHTMLBodyBytes = 2 << 20
)

// errFetchURLBlocked is returned when a requested fetch URL targets a
// non-public address (SSRF guard).
var errFetchURLBlocked = errors.New("url target is not allowed")

// uploadLimitFor resolves the byte cap owned by the body-limit policy for a
// given upload kind. PDFService.UploadLimit delegates here so handlers and
// the service share one policy source.
func uploadLimitFor(kind string) int64 {
	switch kind {
	case UploadKindXFDF:
		return maxXFDFBytes
	case UploadKindFont:
		return maxFontBytes
	default:
		return maxPDFBytes
	}
}

// readBounded reads r capped at limit+1 bytes. ok=false means the input
// exceeded limit (caller should reject, e.g. with 413).
func readBounded(r io.Reader, limit int64) (data []byte, ok bool, err error) {
	data, err = io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, false, err
	}
	if int64(len(data)) > limit {
		return nil, false, nil
	}
	return data, true, nil
}

// pdfErrorStatus maps backend failures to HTTP codes. Sentinel-classified
// errors (errors.Is/As against the gopdflib taxonomy) win; the substring
// list below is the pinned legacy fallback for foreign errors that carry
// only message text (e.g. test mocks, pre-taxonomy engine errors). Keep it
// in sync with pkg/gopdflib's invalidInputSubstrings.
func pdfErrorStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, gopdflib.ErrInvalidInput) {
		return http.StatusUnprocessableEntity
	}
	if errors.Is(err, gopdflib.ErrLimitExceeded) {
		return http.StatusRequestEntityTooLarge
	}
	if errors.Is(err, gopdflib.ErrUpstream) || errors.Is(err, font.ErrUpstream) {
		return http.StatusBadGateway
	}
	var httpErr *font.HTTPStatusError
	if errors.As(err, &httpErr) {
		return http.StatusBadGateway
	}
	if errors.Is(err, gopdflib.ErrInternal) {
		return http.StatusInternalServerError
	}
	msg := strings.ToLower(err.Error())
	for _, s := range []string{
		"exceed", "too large", "maximum size",
	} {
		if strings.Contains(msg, s) {
			return http.StatusRequestEntityTooLarge
		}
	}
	for _, s := range []string{
		"invalid", "malformed", "corrupt", "parse", "password", "encrypt",
		"spec", "page", "range", "empty", "not a pdf", "unsupported",
		"too small", "trailer", "xref", "header", "crypt",
		"no valid", "missing",
	} {
		if strings.Contains(msg, s) {
			return http.StatusUnprocessableEntity
		}
	}
	return http.StatusInternalServerError
}

// errorBody renders the shared {code,message} envelope plus a legacy
// "error" alias equal to message, so clients parsing `error` keep working.
func errorBody(code gopdflib.ErrorCode, message string) gin.H {
	return gin.H{"code": string(code), "message": message, "error": message}
}

// abortError replies with status and the envelope code for that status,
// keeping the caller's message text unchanged.
func abortError(c *gin.Context, status int, message string) {
	c.JSON(status, errorBody(gopdflib.CodeForStatus(status), message))
}

// abortErrorAndStop is abortError followed by Abort (for call sites that
// previously used AbortWithStatusJSON).
func abortErrorAndStop(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, errorBody(gopdflib.CodeForStatus(status), message))
}

// abortPDFError logs backend detail server-side and replies with a generic
// client message plus the mapped status code and envelope code.
func abortPDFError(c *gin.Context, err error) {
	log.Printf("pdf handler %s failed: %v", c.Request.URL.Path, err)
	status := pdfService.ClassifyError(err)
	var message string
	switch status {
	case http.StatusUnprocessableEntity:
		message = "invalid PDF input"
	case http.StatusRequestEntityTooLarge:
		message = "pdf exceeds maximum size"
	case http.StatusBadGateway:
		message = "upstream dependency failed"
	default:
		status = http.StatusInternalServerError
		message = "PDF processing failed"
	}
	c.JSON(status, errorBody(gopdflib.CodeForStatus(status), message))
}

// validateFetchURL allows only http/https URLs whose host does not resolve to
// loopback, private, link-local (incl. cloud metadata), or multicast targets.
func validateFetchURL(raw string) error {
	if raw == "" {
		return nil // HTML-content path: no fetch performed.
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return errors.New("invalid url")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("url scheme must be http or https")
	}
	host := u.Hostname()
	if host == "" {
		return errors.New("invalid url")
	}
	lower := strings.ToLower(strings.TrimSuffix(host, "."))
	if lower == "localhost" || strings.HasSuffix(lower, ".localhost") {
		return errFetchURLBlocked
	}
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedFetchIP(ip) {
			return errFetchURLBlocked
		}
		return nil
	}
	ips, err := net.LookupIP(host)
	if err != nil || len(ips) == 0 {
		return errors.New("invalid url")
	}
	for _, ip := range ips {
		if isBlockedFetchIP(ip) {
			return errFetchURLBlocked
		}
	}
	return nil
}

func isBlockedFetchIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsMulticast() {
		return true
	}
	if ip.IsPrivate() || !ip.IsGlobalUnicast() {
		return true
	}
	return false
}

var pprofForbiddenResp = gin.H{"error": "Forbidden: Pprof is only accessible from localhost"}

// isLoopbackPeer reports whether req arrived over a loopback connection,
// based on the direct peer address. Unlike Gin's ClientIP it ignores
// X-Forwarded-For, which any client can forge when trusted proxies include
// public ranges (Gin's default).
func isLoopbackPeer(req *http.Request) bool {
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(strings.Trim(host, "[]"))
	return ip != nil && ip.IsLoopback()
}
