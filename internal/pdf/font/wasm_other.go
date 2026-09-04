//go:build !js

package font

// isWASM is the non-browser mirror of wasm_js.go: always false, so the
// provisioning and Liberation download paths compile away no extra branch
// cost on server builds.
func isWASM() bool { return false }
