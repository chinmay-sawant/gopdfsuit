// Package fontutils provides cross-platform math font discovery, download, and management.
// It handles font availability across Linux, macOS, Windows, and GCP App Engine environments.
package fontutils

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	provfont "github.com/chinmay-sawant/gopdfsuit/v6/internal/pdf/font"
)

// MathFontInfo describes a math-capable font with its system paths and download URL.
type MathFontInfo struct {
	Name        string // Human-readable name
	FileName    string // File name (e.g., "NotoSansMath-Regular.ttf")
	LinuxPaths  []string
	MacPaths    []string
	WinPaths    []string
	DownloadURL string // GitHub raw URL for fallback download
	// SHA256 pins the expected download bytes (hex). Downloads whose digest
	// mismatches are rejected. Recorded 2026-09-04; if upstream re-releases
	// the file the mismatch error names the URL so the pin can be refreshed.
	SHA256 string
}

// mathFonts defines all supported math fonts with cross-platform paths and download URLs.
var mathFonts = []MathFontInfo{
	{
		Name:     "NotoSansMath",
		FileName: "NotoSansMath-Regular.ttf",
		LinuxPaths: []string{
			"/usr/share/fonts/truetype/noto/NotoSansMath-Regular.ttf",
			"/usr/share/fonts/opentype/noto/NotoSansMath-Regular.otf",
			"/usr/share/fonts/noto/NotoSansMath-Regular.ttf",
		},
		MacPaths: []string{
			"/Library/Fonts/NotoSansMath-Regular.ttf",
			filepath.Join(os.Getenv("HOME"), "Library/Fonts/NotoSansMath-Regular.ttf"),
		},
		WinPaths: []string{
			filepath.Join(os.Getenv("WINDIR"), "Fonts", "NotoSansMath-Regular.ttf"),
			`C:\Windows\Fonts\NotoSansMath-Regular.ttf`,
		},
		DownloadURL: "https://github.com/notofonts/math/raw/refs/heads/main/fonts/NotoSansMath/unhinted/ttf/NotoSansMath-Regular.ttf",
		SHA256:      "05a0ba975a623d7c1b72bb16621801bc975b001491add5e36f3871cd0cd2fada",
	},
	{
		Name:     "DejaVuSans",
		FileName: "DejaVuSans.ttf",
		LinuxPaths: []string{
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/dejavu/DejaVuSans.ttf",
		},
		MacPaths: []string{
			"/Library/Fonts/DejaVuSans.ttf",
			filepath.Join(os.Getenv("HOME"), "Library/Fonts/DejaVuSans.ttf"),
		},
		WinPaths: []string{
			filepath.Join(os.Getenv("WINDIR"), "Fonts", "DejaVuSans.ttf"),
			`C:\Windows\Fonts\DejaVuSans.ttf`,
		},
		DownloadURL: "https://github.com/dejavu-fonts/dejavu-fonts/raw/refs/heads/master/src/DejaVuSans.ttf",
		SHA256:      "2ba21d21b6edd3e5d8837bab94d112cfa78f824e7cc346d08277819a5a77b1b4",
	},
	{
		Name:     "LiberationSans",
		FileName: "LiberationSans-Regular.ttf",
		LinuxPaths: []string{
			"/usr/share/fonts/truetype/liberation2/LiberationSans-Regular.ttf",
			"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
			"/usr/share/fonts/liberation/LiberationSans-Regular.ttf",
		},
		MacPaths: []string{
			"/Library/Fonts/LiberationSans-Regular.ttf",
			filepath.Join(os.Getenv("HOME"), "Library/Fonts/LiberationSans-Regular.ttf"),
		},
		WinPaths: []string{
			filepath.Join(os.Getenv("WINDIR"), "Fonts", "LiberationSans-Regular.ttf"),
			`C:\Windows\Fonts\LiberationSans-Regular.ttf`,
		},
		DownloadURL: "https://github.com/liberationfonts/liberation-fonts/raw/refs/heads/main/src/LiberationSans-Regular.ttf",
		SHA256:      "f2c9774fcfb226ac2efad0896b97388c96340f468db5482dd9d186d243eabe55",
	},
}

// fontsDir returns the directory where downloaded fonts are stored.
// Uses a "fonts" subdirectory relative to the executable, or /tmp/gopdfsuit-fonts as fallback.
func fontsDir() string {
	// Check for explicit override (useful for GCP App Engine / Cloud Run)
	if dir := os.Getenv("GOPDFSUIT_FONTS_DIR"); dir != "" {
		return dir
	}

	// Default: /tmp/gopdfsuit-fonts (writable on all platforms including App Engine)
	return filepath.Join(os.TempDir(), "gopdfsuit-fonts")
}

// downloadedFontPath returns the path where a downloaded font would be stored.
func downloadedFontPath(fileName string) string {
	return filepath.Join(fontsDir(), fileName)
}

// MathFontCandidates returns all candidate font paths for the current OS,
// including both system paths and the downloaded font fallback path.
// This replaces the hardcoded mathFontCandidates variable from generator.go.
func MathFontCandidates() []string {
	var paths []string

	for _, font := range mathFonts {
		switch runtime.GOOS {
		case "linux":
			paths = append(paths, font.LinuxPaths...)
		case "darwin":
			paths = append(paths, font.MacPaths...)
		case "windows":
			paths = append(paths, font.WinPaths...)
		default:
			// Fallback to Linux paths
			paths = append(paths, font.LinuxPaths...)
		}
		// Always include the downloaded font path as final fallback
		paths = append(paths, downloadedFontPath(font.FileName))
	}

	return paths
}

// EnsureMathFonts checks if math fonts exist on the system and downloads
// any missing ones. It blocks until all downloads finish (despite any older
// "background" wording: the internal WaitGroup is waited on before return).
// Missing fonts are non-fatal: errors are logged and dropped. Callers that
// need the errors should use EnsureMathFontsContext instead.
func EnsureMathFonts() {
	_ = EnsureMathFontsContext(context.Background())
}

// EnsureMathFontsContext is the context-aware core of EnsureMathFonts: it
// blocks until all missing-font downloads finish and returns one error per
// failed font (nil when every font was already present or downloaded).
// The context deadline propagates to the download HTTP requests.
func EnsureMathFontsContext(ctx context.Context) []error {
	return ensureMathFonts(ctx, mathFonts)
}

func ensureMathFonts(ctx context.Context, fonts []MathFontInfo) []error {
	var mu sync.Mutex
	var errs []error
	var wg sync.WaitGroup

	for _, font := range fonts {
		if fontExistsOnSystem(font) {
			log.Printf("[fontutils] Font %s found on system", font.Name)
			continue
		}

		// Check if already downloaded
		dlPath := downloadedFontPath(font.FileName)
		if _, err := os.Stat(dlPath); err == nil {
			log.Printf("[fontutils] Font %s already downloaded at %s", font.Name, dlPath)
			continue
		}

		// Download missing fonts concurrently, then wait below.
		wg.Add(1)
		go func(f MathFontInfo) {
			defer wg.Done()
			if err := downloadFontContext(ctx, f); err != nil {
				log.Printf("[fontutils] WARNING: failed to download font %s: %v", f.Name, err)
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			} else {
				log.Printf("[fontutils] Downloaded font %s to %s", f.Name, downloadedFontPath(f.FileName))
			}
		}(font)
	}

	wg.Wait()
	log.Println("[fontutils] Math font initialization complete")
	return errs
}

// fontExistsOnSystem checks if any of the system paths for a font exist.
func fontExistsOnSystem(font MathFontInfo) bool {
	var candidates []string
	switch runtime.GOOS {
	case "linux":
		candidates = font.LinuxPaths
	case "darwin":
		candidates = font.MacPaths
	case "windows":
		candidates = font.WinPaths
	default:
		candidates = font.LinuxPaths
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

// maxFontDownloadSize caps a single font download to prevent resource
// exhaustion. It is a var (not const) so tests can shrink it.
var maxFontDownloadSize = 20 * 1024 * 1024

// downloadFontContext is the context-aware download core. It delegates the
// fetch (timeout client, size cap, digest pin, temp file) to the shared
// provision helper in internal/pdf/font, then atomically installs the result.
// Fonts larger than maxFontDownloadSize are rejected and never cached.
func downloadFontContext(ctx context.Context, font MathFontInfo) error {
	if font.DownloadURL == "" {
		return fmt.Errorf("no download URL for font %s", font.Name)
	}

	destDir := fontsDir()
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create fonts dir: %w", err)
	}

	tmpPath, err := provfont.FetchToTemp(ctx, font.DownloadURL, destDir,
		font.FileName+".tmp.*", int64(maxFontDownloadSize), font.SHA256, 60*time.Second)
	if err != nil {
		return fmt.Errorf("download %s: %w", font.Name, err)
	}

	destPath := downloadedFontPath(font.FileName)
	if err := os.Rename(tmpPath, destPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}
