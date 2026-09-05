package font

import (
	"os"
	"path/filepath"
	"testing"
)

// vendoredTTF reads an OFL-vendored Liberation face shipped for the browser
// build. Tests must not depend on system fonts or the network.
func vendoredTTF(t *testing.T, file string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "frontend", "public", "fonts", file))
	if err != nil {
		t.Fatalf("read vendored font %s: %v", file, err)
	}
	return data
}

func provisionManager(t *testing.T) *PDFAFontManager {
	return &PDFAFontManager{
		loadedFonts: make(map[string]*TTFFont),
		config: PDFAFontConfig{
			FontsDirectory:         filepath.Join(t.TempDir(), "missing-fonts"),
			FallbackFontsDirectory: filepath.Join(t.TempDir(), "missing-system-fonts"),
			AutoDownload:           false,
		},
		initialized: true,
	}
}

// TestRegisterLiberationSkipsEnsureWhenFacesRegistered mirrors the browser
// flow: JS pre-registers TTF bytes via RegisterFontFromData, so provision
// must not touch EnsureFontsAvailable (which rejects outright under
// GOOS=js and fails here with no fonts dir and no downloads).
func TestRegisterLiberationSkipsEnsureWhenFacesRegistered(t *testing.T) {
	manager := provisionManager(t)

	registry := NewFontRegistry()
	if err := registry.RegisterFontFromData(
		"Helvetica", vendoredTTF(t, "LiberationSans-Regular.ttf"),
	); err != nil {
		t.Fatalf("RegisterFontFromData failed: %v", err)
	}

	if err := manager.RegisterLiberationFontsForPDFA(registry, []string{"Helvetica"}); err != nil {
		t.Fatalf("provision with registered faces failed: %v", err)
	}
	if !registry.HasFont("Helvetica") {
		t.Fatal("expected Helvetica to remain registered after provision")
	}
	if manager.ensureAttempted {
		t.Fatal("expected ensure/download step to be skipped when faces are registered")
	}
}

// TestRegisterLiberationStillEnsuresWhenFacesMissing guards the other half:
// with nothing registered, provision must still go through Ensure (and fail
// here, where no fonts dir exists and downloads are off).
func TestRegisterLiberationStillEnsuresWhenFacesMissing(t *testing.T) {
	manager := provisionManager(t)

	registry := NewFontRegistry()
	if err := manager.RegisterLiberationFontsForPDFA(registry, []string{"Helvetica"}); err == nil {
		t.Fatal("expected provision with missing faces to fail via ensure")
	}
	if !manager.ensureAttempted {
		t.Fatal("expected ensure/download step to run when faces are missing")
	}
}
