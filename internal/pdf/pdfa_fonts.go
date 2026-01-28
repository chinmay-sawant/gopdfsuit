package pdf

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// PDFAFontConfig holds configuration for PDF/A compliant font handling
type PDFAFontConfig struct {
	// FontsDirectory is where Liberation fonts are stored (default: ~/.gopdfsuit/fonts)
	FontsDirectory string
	// FallbackFontsDirectory is an alternative location for fonts (e.g., /usr/share/fonts/liberation)
	FallbackFontsDirectory string
	// AutoDownload enables automatic downloading of Liberation fonts if not found
	AutoDownload bool
}

// LiberationFontMapping maps standard PDF fonts to Liberation equivalents
var LiberationFontMapping = map[string]string{
	// Helvetica family -> Liberation Sans
	"Helvetica":             "LiberationSans-Regular",
	"Helvetica-Bold":        "LiberationSans-Bold",
	"Helvetica-Oblique":     "LiberationSans-Italic",
	"Helvetica-BoldOblique": "LiberationSans-BoldItalic",

	// Times family -> Liberation Serif
	"Times-Roman":      "LiberationSerif-Regular",
	"Times-Bold":       "LiberationSerif-Bold",
	"Times-Italic":     "LiberationSerif-Italic",
	"Times-BoldItalic": "LiberationSerif-BoldItalic",

	// Courier family -> Liberation Mono
	"Courier":             "LiberationMono-Regular",
	"Courier-Bold":        "LiberationMono-Bold",
	"Courier-Oblique":     "LiberationMono-Italic",
	"Courier-BoldOblique": "LiberationMono-BoldItalic",
}

// LiberationFontFiles maps font names to file names
var LiberationFontFiles = map[string]string{
	"LiberationSans-Regular":    "LiberationSans-Regular.ttf",
	"LiberationSans-Bold":       "LiberationSans-Bold.ttf",
	"LiberationSans-Italic":     "LiberationSans-Italic.ttf",
	"LiberationSans-BoldItalic": "LiberationSans-BoldItalic.ttf",

	"LiberationSerif-Regular":    "LiberationSerif-Regular.ttf",
	"LiberationSerif-Bold":       "LiberationSerif-Bold.ttf",
	"LiberationSerif-Italic":     "LiberationSerif-Italic.ttf",
	"LiberationSerif-BoldItalic": "LiberationSerif-BoldItalic.ttf",

	"LiberationMono-Regular":    "LiberationMono-Regular.ttf",
	"LiberationMono-Bold":       "LiberationMono-Bold.ttf",
	"LiberationMono-Italic":     "LiberationMono-Italic.ttf",
	"LiberationMono-BoldItalic": "LiberationMono-BoldItalic.ttf",
}

// Liberation font download URLs (GitHub releases)
// These are the open-source Liberation fonts from Red Hat
const liberationFontsBaseURL = "https://github.com/liberationfonts/liberation-fonts/releases/download/2.1.5/"
const liberationFontsZipURL = liberationFontsBaseURL + "liberation-fonts-ttf-2.1.5.tar.gz"

// PDFAFontManager manages Liberation font loading for PDF/A compliance
type PDFAFontManager struct {
	mu          sync.RWMutex
	config      PDFAFontConfig
	loadedFonts map[string]*TTFFont
	initialized bool
}

// Global PDF/A font manager
var pdfaFontManager = &PDFAFontManager{
	loadedFonts: make(map[string]*TTFFont),
}

// GetPDFAFontManager returns the global PDF/A font manager
func GetPDFAFontManager() *PDFAFontManager {
	return pdfaFontManager
}

// Initialize sets up the font manager with the given config
func (m *PDFAFontManager) Initialize(config PDFAFontConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.initialize(config)
}

// initialize is the internal non-locking version of Initialize
func (m *PDFAFontManager) initialize(config PDFAFontConfig) error {
	if config.FontsDirectory == "" {
		// Default to ~/.gopdfsuit/fonts
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = "."
		}
		config.FontsDirectory = filepath.Join(homeDir, ".gopdfsuit", "fonts")
	}

	if config.FallbackFontsDirectory == "" {
		// Detect OS and set fallback directories
		switch runtime.GOOS {
		case "windows":
			// Typically fonts are in C:\Windows\Fonts, but we won't find Liberation there by default usually
			// The user can set it manually, or we rely on the download.
			config.FallbackFontsDirectory = `C:\Windows\Fonts`
		case "darwin":
			config.FallbackFontsDirectory = "/Library/Fonts"
		default:
			// Linux/Unix
			systemPaths := []string{
				"/usr/share/fonts/liberation",
				"/usr/share/fonts/truetype/liberation",
				"/usr/share/fonts/liberation-sans",
				"/usr/local/share/fonts/liberation",
			}
			for _, p := range systemPaths {
				if _, err := os.Stat(p); err == nil {
					config.FallbackFontsDirectory = p
					break
				}
			}
		}
	}

	m.config = config
	m.initialized = true
	return nil
}

// EnsureFontsAvailable ensures Liberation fonts are available, downloading if necessary
func (m *PDFAFontManager) EnsureFontsAvailable() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.initialized {
		if err := m.initialize(PDFAFontConfig{AutoDownload: true}); err != nil {
			return err
		}
	}

	// Check if fonts are already available
	fontsDir := m.findFontsDirectory()
	if fontsDir != "" {
		return nil
	}

	if !m.config.AutoDownload {
		return fmt.Errorf("liberation fonts not found. Please install them or enable auto-download.\n"+
			"On Ubuntu/Debian: sudo apt-get install fonts-liberation\n"+
			"On Fedora: sudo dnf install liberation-fonts\n"+
			"On macOS: brew install font-liberation\n"+
			"Or place TTF files in: %s", m.config.FontsDirectory)
	}

	// Create fonts directory
	if err := os.MkdirAll(m.config.FontsDirectory, 0755); err != nil {
		return fmt.Errorf("failed to create fonts directory: %w", err)
	}

	// Download fonts using the ZIP/Tarball to ensure all variants are present
	return m.downloadFonts()
}

// findFontsDirectory finds a directory containing Liberation fonts
func (m *PDFAFontManager) findFontsDirectory() string {
	// Check primary directory
	if m.checkFontDir(m.config.FontsDirectory) {
		return m.config.FontsDirectory
	}

	// Check fallback directory
	if m.checkFontDir(m.config.FallbackFontsDirectory) {
		return m.config.FallbackFontsDirectory
	}

	return ""
}

// checkFontDir checks if a directory contains at least one Liberation font
func (m *PDFAFontManager) checkFontDir(dir string) bool {
	if dir == "" {
		return false
	}

	// Check for at least one font file
	testFile := filepath.Join(dir, "LiberationSans-Regular.ttf")
	if _, err := os.Stat(testFile); err == nil {
		return true
	}

	return false
}

// downloadFonts downloads Liberation font files
func (m *PDFAFontManager) downloadFonts() error {
	fmt.Printf("Downloading Liberation fonts from %s...\n", liberationFontsZipURL)

	// Create temp file for the tar.gz
	tmpFile, err := os.CreateTemp("", "liberation-fonts-*.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer func() {
		_ = os.Remove(tmpFile.Name()) // Clean up
	}()
	defer func() {
		_ = tmpFile.Close()
	}()

	// Download the file
	resp, err := http.Get(liberationFontsZipURL)
	if err != nil {
		return fmt.Errorf("failed to download fonts: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download fonts: HTTP %d", resp.StatusCode)
	}

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save fonts archive: %w", err)
	}

	// Seek back to start
	if _, err := tmpFile.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek temp file: %w", err)
	}

	// Extract the tar.gz
	gzr, err := gzip.NewReader(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		_ = gzr.Close()
	}()

	tr := tar.NewReader(gzr)

	fmt.Printf("Extracting fonts to %s...\n", m.config.FontsDirectory)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Check if it's a TTF file
		if header.Typeflag == tar.TypeReg && strings.HasSuffix(header.Name, ".ttf") {
			// Extract just the filename (ignore directory structure in tar)
			fileName := filepath.Base(header.Name)

			// Only extract fonts we care about (optimization)
			found := false
			for _, knownFile := range LiberationFontFiles {
				if fileName == knownFile {
					found = true
					break
				}
			}

			if !found {
				continue
			}

			destPath := filepath.Join(m.config.FontsDirectory, fileName)
			outFile, err := os.Create(destPath)
			if err != nil {
				return fmt.Errorf("failed to create font file %s: %w", destPath, err)
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				_ = outFile.Close()
				return fmt.Errorf("failed to extract %s: %w", fileName, err)
			}
			if err := outFile.Close(); err != nil {
				return fmt.Errorf("failed to close file %s: %w", fileName, err)
			}
			fmt.Printf("Parsed: %s\n", fileName)
		}
	}

	return nil
}

// GetLiberationFont loads and returns a Liberation font by its PDF standard name
// e.g., "Helvetica" returns the LiberationSans-Regular font
func (m *PDFAFontManager) GetLiberationFont(standardFontName string) (*TTFFont, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Map standard font name to Liberation equivalent
	liberationName, ok := LiberationFontMapping[standardFontName]
	if !ok {
		return nil, fmt.Errorf("no Liberation font mapping for: %s", standardFontName)
	}

	// Check if already loaded
	if font, ok := m.loadedFonts[liberationName]; ok {
		return font, nil
	}

	// Find fonts directory
	fontsDir := m.findFontsDirectory()
	if fontsDir == "" {
		return nil, fmt.Errorf("liberation fonts not found. Run EnsureFontsAvailable() first")
	}

	// Load the font
	fileName := LiberationFontFiles[liberationName]
	fontPath := filepath.Join(fontsDir, fileName)

	font, err := LoadTTFFromFile(fontPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %w", fontPath, err)
	}

	m.loadedFonts[liberationName] = font
	return font, nil
}

// RegisterLiberationFontsForPDFA registers all required Liberation fonts with the font registry
// In PDF/A mode, standard fonts are replaced with Liberation equivalents
// This registers them under their STANDARD names so getFontReference picks them up
func (m *PDFAFontManager) RegisterLiberationFontsForPDFA(registry *CustomFontRegistry, usedStandardFonts []string) error {
	if err := m.EnsureFontsAvailable(); err != nil {
		return err
	}

	for _, stdFont := range usedStandardFonts {
		// Skip if not a mappable standard font
		if _, ok := LiberationFontMapping[stdFont]; !ok {
			continue
		}

		// Skip if already registered (check under the STANDARD font name)
		if registry.HasFont(stdFont) {
			continue
		}

		font, err := m.GetLiberationFont(stdFont)
		if err != nil {
			return err
		}

		// Register under the STANDARD font name, not the Liberation name
		// This way getFontReference will find it and use the embedded font
		if err := registry.RegisterFont(stdFont, font); err != nil {
			return err
		}
	}

	return nil
}

// GetMappedFontName returns the Liberation font name for a standard font
// Returns the original name if no mapping exists
func GetMappedFontName(standardFontName string, pdfaMode bool) string {
	if !pdfaMode {
		return standardFontName
	}

	if liberationName, ok := LiberationFontMapping[standardFontName]; ok {
		return liberationName
	}

	return standardFontName
}

// IsStandardFont checks if a font name is a standard PDF Type 1 font
func IsStandardFont(fontName string) bool {
	_, ok := LiberationFontMapping[fontName]
	return ok
}

// GetLiberationFontPostScriptName returns the PostScript name for a Liberation font
func GetLiberationFontPostScriptName(liberationName string) string {
	// Liberation fonts use their name as PostScript name with hyphens
	psNames := map[string]string{
		"LiberationSans-Regular":     "LiberationSans",
		"LiberationSans-Bold":        "LiberationSans-Bold",
		"LiberationSans-Italic":      "LiberationSans-Italic",
		"LiberationSans-BoldItalic":  "LiberationSans-BoldItalic",
		"LiberationSerif-Regular":    "LiberationSerif",
		"LiberationSerif-Bold":       "LiberationSerif-Bold",
		"LiberationSerif-Italic":     "LiberationSerif-Italic",
		"LiberationSerif-BoldItalic": "LiberationSerif-BoldItalic",
		"LiberationMono-Regular":     "LiberationMono",
		"LiberationMono-Bold":        "LiberationMono-Bold",
		"LiberationMono-Italic":      "LiberationMono-Italic",
		"LiberationMono-BoldItalic":  "LiberationMono-BoldItalic",
	}

	if psName, ok := psNames[liberationName]; ok {
		return psName
	}
	return liberationName
}
