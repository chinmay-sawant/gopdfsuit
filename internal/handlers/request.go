package handlers

import (
	"context"
	"errors"
	"io"
	"log"
	"mime/multipart"
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
	// maxMultipartBodyBytes caps non-merge multipart request bodies before
	// multipart parsing begins.
	maxMultipartBodyBytes = 40 << 20
	// maxMergeBodyBytes caps both the merge request body and accepted PDF bytes.
	maxMergeBodyBytes = 128 << 20
	// maxMergeFiles caps the number of PDFs accepted by one merge request.
	maxMergeFiles = 32
)

// errFetchURLBlocked is returned when a requested fetch URL targets a
// non-public address (SSRF guard).
var errFetchURLBlocked = errors.New("url target is not allowed")

type fetchURLResolver interface {
	LookupIP(context.Context, string, string) ([]net.IP, error)
}

var defaultFetchURLResolver fetchURLResolver = net.DefaultResolver

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
// errors resolve through the shared gopdflib taxonomy (CodeOf); the legacy
// substring fallback for foreign errors that carry only message text (e.g.
// test mocks) delegates to gopdflib.ClassifyMessage, the single source that
// also feeds wrapEngineError, so the signal lists cannot drift (notably
// "limit" is consistently an over-cap signal on both sides).
func pdfErrorStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	if errors.Is(err, gopdflib.ErrInternal) {
		return http.StatusInternalServerError
	}
	if code := gopdflib.CodeOf(err); code != gopdflib.CodeInternal {
		return gopdflib.StatusForCode(code)
	}
	if errors.Is(err, font.ErrUpstream) {
		return http.StatusBadGateway
	}
	var httpErr *font.HTTPStatusError
	if errors.As(err, &httpErr) {
		return http.StatusBadGateway
	}
	return gopdflib.StatusForCode(gopdflib.ClassifyMessage(err))
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
// client message plus the mapped status code and envelope code. fallbackMsg
// is the message for unclassified (500) failures, letting redaction and
// other ops name the operation without owning a parallel abort helper.
func abortPDFError(c *gin.Context, err error, fallbackMsg string) {
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
		message = fallbackMsg
	}
	c.JSON(status, errorBody(gopdflib.CodeForStatus(status), message))
}

// validateFetchURL allows only http/https URLs whose host does not resolve to
// loopback, private, link-local (incl. cloud metadata), or multicast targets.
// It is the outer SSRF guard for the htmltopdf/htmltoimage URL path: handlers
// run it before internal/pdf builds a gowkhtmltopdf Content{URL:}, and the
// adapter additionally sets RestrictedNetworkPolicy (blocks private
// destinations plus cross-host redirects) for the page fetch and subresource
// CSS inside the engine. Base/Allow/AllowLocalFiles stay at engine defaults
// (local file reads disabled), so subresource CSS loads over http(s) only.
func validateFetchURL(ctx context.Context, raw string) error {
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
	ips, err := defaultFetchURLResolver.LookupIP(ctx, "ip", host)
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

// pprofForbiddenResp uses the shared errorBody envelope so pprof denials
// carry the same code/message/error shape as every other handler rejection.
var pprofForbiddenResp = errorBody(gopdflib.CodeForStatus(http.StatusForbidden), "Forbidden: Pprof is only accessible from localhost")

// overLimitMessage names the upload kind in 413 rejections so every handler
// shares one policy vocabulary.
func overLimitMessage(kind string) string {
	if kind == UploadKindFont {
		return "font file exceeds maximum size"
	}
	return kind + " exceeds maximum size"
}

// readUploadData reads an opened multipart file header through the
// service-owned body-limit policy. It returns nil after writing the
// rejection when the handler must abort.
func readUploadData(c *gin.Context, fh *multipart.FileHeader, kind string) []byte {
	f, err := fh.Open()
	if err != nil {
		log.Printf("upload: open %s failed: %v", fh.Filename, err)
		abortError(c, http.StatusInternalServerError, "failed to process upload")
		return nil
	}
	defer func() { _ = f.Close() }()

	data, ok, err := pdfService.ReadUpload(f, kind)
	if err != nil {
		log.Printf("upload: read %s failed: %v", fh.Filename, err)
		abortError(c, http.StatusBadRequest, "invalid request")
		return nil
	}
	if !ok {
		abortError(c, http.StatusRequestEntityTooLarge, overLimitMessage(kind))
		return nil
	}
	if len(data) == 0 {
		abortError(c, http.StatusBadRequest, kind+" file is empty")
		return nil
	}
	return data
}

// readSingleUpload opens the form file for field and reads it through the
// service-owned body-limit policy, rejecting oversized uploads with 413 and
// missing or empty uploads with 400. It returns nil when the handler must
// abort (response already written). This is the single upload policy for
// the redact, compress, split, fill, and font handlers.
func readSingleUpload(c *gin.Context, field, kind string) []byte {
	if !ensureMultipartBodyLimit(c) {
		return nil
	}
	fh, err := c.FormFile(field)
	if err != nil {
		if isBodyTooLargeErr(err) {
			abortError(c, http.StatusRequestEntityTooLarge, overLimitMessage(kind))
			return nil
		}
		abortError(c, http.StatusBadRequest, field+" file is required")
		return nil
	}
	return readUploadData(c, fh, kind)
}

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
