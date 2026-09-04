// Error taxonomy for the gopdflib seam.
//
// gopdflib is the single validating interface shared by the Go, CGO/Python,
// HTTP, and WASM entry points, so it owns the error language: four sentinels
// (ErrInvalidInput, ErrLimitExceeded, ErrUpstream, ErrInternal) and a stable
// {code,message} envelope. Every public function wraps its failures with one
// of these sentinels via %w, so callers classify with errors.Is instead of
// re-deriving meaning from message text.
//
// The substring lists below are the pinned legacy fallback: engine errors
// predate the sentinels and carry only message text, so wrapEngineError maps
// them once at this seam. internal/handlers keeps the same list for foreign
// (non-gopdflib) errors such as test mocks.
package gopdflib

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bytedance/sonic"
)

var (
	// ErrInvalidInput marks malformed client input (bad PDF bytes, bad spec,
	// missing fields). HTTP: 422 (or 400 for transport-level rejects).
	ErrInvalidInput = errors.New("gopdflib: invalid input")
	// ErrLimitExceeded marks inputs over an accepted size cap. HTTP: 413.
	ErrLimitExceeded = errors.New("gopdflib: limit exceeded")
	// ErrUpstream marks a downstream dependency failure (font-asset fetch,
	// headless-Chrome conversion). HTTP: 502.
	ErrUpstream = errors.New("gopdflib: upstream failure")
	// ErrInternal is the fallback for engine failures that match no other
	// class. HTTP: 500.
	ErrInternal = errors.New("gopdflib: internal failure")
)

// ErrorCode is the stable machine-readable error code shipped in the
// {code,message} envelope to HTTP, CGO/Python, and WASM callers.
type ErrorCode string

const (
	// CodeInvalidInput maps to HTTP 422 (400 for transport-level rejects).
	CodeInvalidInput ErrorCode = "invalid_input"
	// CodeLimitExceeded maps to HTTP 413.
	CodeLimitExceeded ErrorCode = "limit_exceeded"
	// CodeUpstream maps to HTTP 502.
	CodeUpstream ErrorCode = "upstream"
	// CodeInternal maps to HTTP 500.
	CodeInternal ErrorCode = "internal"
)

// ErrorEnvelope is the {code,message} shape shared by every entry point.
// HTTP responses add a legacy "error" alias equal to message so existing
// clients parsing `error` keep working.
type ErrorEnvelope struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// CodeOf classifies err via the sentinel chain. Unknown or nil-adjacent
// errors map to CodeInternal; nil maps to "".
func CodeOf(err error) ErrorCode {
	if err == nil {
		return ""
	}
	switch {
	case errors.Is(err, ErrInvalidInput):
		return CodeInvalidInput
	case errors.Is(err, ErrLimitExceeded):
		return CodeLimitExceeded
	case errors.Is(err, ErrUpstream):
		return CodeUpstream
	default:
		return CodeInternal
	}
}

// EnvelopeOf folds err into the shared {code,message} shape.
func EnvelopeOf(err error) ErrorEnvelope {
	if err == nil {
		return ErrorEnvelope{Code: CodeInternal}
	}
	return ErrorEnvelope{Code: CodeOf(err), Message: err.Error()}
}

// EnvelopeJSON renders the envelope for transport error fields (CGO keeps
// its ByteResult ABI: the `error` C string carries this JSON document).
func EnvelopeJSON(err error) string {
	raw, mErr := sonic.Marshal(EnvelopeOf(err))
	if mErr != nil {
		return `{"code":"internal","message":"gopdflib: internal failure"}`
	}
	return string(raw)
}

// CodeForStatus maps an HTTP status to the envelope code, so handlers that
// reject input before reaching gopdflib (transport guards, body caps) emit
// the same codes without importing sentinel logic per call site.
func CodeForStatus(status int) ErrorCode {
	switch status {
	case http.StatusRequestEntityTooLarge, http.StatusTooManyRequests:
		return CodeLimitExceeded
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return CodeUpstream
	case http.StatusInternalServerError:
		return CodeInternal
	default:
		return CodeInvalidInput
	}
}

// StatusForCode maps an envelope code back to its HTTP status.
func StatusForCode(code ErrorCode) int {
	switch code {
	case CodeLimitExceeded:
		return http.StatusRequestEntityTooLarge
	case CodeUpstream:
		return http.StatusBadGateway
	case CodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusUnprocessableEntity
	}
}

// invalidInputSubstrings is the pinned legacy signal list for malformed
// client input. Keep in sync with internal/handlers' fallback copy.
var invalidInputSubstrings = []string{
	"invalid", "malformed", "corrupt", "parse", "password", "encrypt",
	"spec", "page", "range", "empty", "not a pdf", "unsupported",
	"too small", "trailer", "xref", "header", "crypt",
	"no valid", "missing",
}

// limitSubstrings is the pinned legacy signal list for over-cap inputs.
var limitSubstrings = []string{
	"exceed", "too large", "maximum size", "limit",
}

func containsAny(lower string, subs []string) bool {
	for _, s := range subs {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

// ClassifyMessage maps a legacy free-text error (no sentinel in the chain)
// to an envelope code using the pinned substring lists above. It is the
// single source for the handlers fallback path (pdfErrorStatus delegates
// here), so the two lists cannot drift again. Limit signals win over
// invalid-input signals, matching wrapEngineError.
func ClassifyMessage(err error) ErrorCode {
	if err == nil {
		return ""
	}
	lower := strings.ToLower(err.Error())
	if containsAny(lower, limitSubstrings) {
		return CodeLimitExceeded
	}
	if containsAny(lower, invalidInputSubstrings) {
		return CodeInvalidInput
	}
	return CodeInternal
}

// wrapEngineError folds a raw engine error into the sentinel taxonomy once,
// at this seam. Limit signals win over invalid-input signals (an oversized
// but otherwise valid PDF is a 413, not a 422); anything else is internal.
// The original error stays in the %w chain for errors.Is/As and logs.
func wrapEngineError(op string, err error) error {
	if err == nil {
		return nil
	}
	lower := strings.ToLower(err.Error())
	if containsAny(lower, limitSubstrings) {
		return fmt.Errorf("%w: %s: %w", ErrLimitExceeded, op, err)
	}
	if containsAny(lower, invalidInputSubstrings) {
		return fmt.Errorf("%w: %s: %w", ErrInvalidInput, op, err)
	}
	return fmt.Errorf("%w: %s: %w", ErrInternal, op, err)
}

// invalidInputError wraps a boundary-validation failure at a named op.
func invalidInputError(op, msg string) error {
	return fmt.Errorf("%w: %s: %s", ErrInvalidInput, op, msg)
}

// limitExceededError wraps an over-cap rejection at a named op.
func limitExceededError(op, msg string) error {
	return fmt.Errorf("%w: %s: %s", ErrLimitExceeded, op, msg)
}
