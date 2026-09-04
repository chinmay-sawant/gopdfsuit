//go:build js

package font

// isWASM reports whether this build targets the browser (GOOS=js, wasm).
// Font provisioning (net/http fetch plus os.CreateTemp) and the Liberation
// auto-download have no browser equivalent; the WASM path is to fetch font
// bytes in JS and register them via
// (CustomFontRegistry).RegisterFontFromData or RegisterFontFromBase64,
// which stay pure in-memory. Defined here behind //go:build js with its
// mirror in wasm_other.go so non-WASM builds pay nothing.
func isWASM() bool { return true }
