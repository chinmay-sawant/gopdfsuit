// Package fontutils provides cross-platform math font discovery, download, and management.
// It handles font availability across Linux, macOS, Windows, and GCP App Engine environments.
package fontutils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

// MathFontInfo describes a math-capable font with its system paths and download URL.
type MathFontInfo struct {
	Name        string // Human-readable name
	FileName    string // File name (e.g., "NotoSansMath-Regular.ttf")
	LinuxPaths  []string
	MacPaths    []string
	WinPaths    []string
	DownloadURL string // GitHub raw URL for fallback download
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
// any missing ones in parallel. This should be called at server startup.
// It logs progress and errors but does not return errors — missing fonts
// are non-fatal (math rendering will degrade gracefully).
func EnsureMathFonts() {
	var wg sync.WaitGroup

	for _, font := range mathFonts {
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

		// Download in background
		wg.Add(1)
		go func(f MathFontInfo) {
			defer wg.Done()
			if err := downloadFont(f); err != nil {
				log.Printf("[fontutils] WARNING: failed to download font %s: %v", f.Name, err)
			} else {
				log.Printf("[fontutils] Downloaded font %s to %s", f.Name, downloadedFontPath(f.FileName))
			}
		}(font)
	}

	wg.Wait()
	log.Println("[fontutils] Math font initialization complete")
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

// downloadFont downloads a font file from its GitHub URL to the local fonts directory.
func downloadFont(font MathFontInfo) error {
	if font.DownloadURL == "" {
		return fmt.Errorf("no download URL for font %s", font.Name)
	}

	destDir := fontsDir()
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create fonts dir: %w", err)
	}

	destPath := downloadedFontPath(font.FileName)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(font.DownloadURL) //nolint:gosec // URLs are hardcoded constants, not user input
	if err != nil {
		return fmt.Errorf("download %s: %w", font.Name, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %d", font.Name, resp.StatusCode)
	}

	// Write to temp file first then rename for atomicity
	tmpFile, err := os.CreateTemp(destDir, font.FileName+".tmp.*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Limit download size to 20MB to prevent resource exhaustion
	const maxFontSize = 20 * 1024 * 1024
	limitedReader := io.LimitReader(resp.Body, maxFontSize)

	_, err = io.Copy(tmpFile, limitedReader)
	tmpFile.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("write font file: %w", err)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename temp file: %w", err)
	}

	return nil
}
