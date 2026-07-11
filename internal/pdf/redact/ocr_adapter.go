package redact

import (
	"bufio"
	"bytes"
	"errors"
	"image"
	_ "image/png" // Register PNG decoder for image.DecodeConfig
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/v5/internal/models"
)

type ocrWord struct {
	PageNum int
	X       float64
	Y       float64
	Width   float64
	Height  float64
	Text    string
}

// OCRProvider is an adapter interface for OCR backends.
type OCRProvider interface {
	ExtractWords(pdfBytes []byte, settings models.OCRSettings) ([]ocrWord, error)
}

type tesseractProvider struct{}

func getOCRProvider(settings models.OCRSettings) (OCRProvider, error) {
	provider := trimSpace(settings.Provider)
	// PERF-48: len precheck before EqualFold
	if provider == "" || (len(provider) == len("tesseract") && strings.EqualFold(provider, "tesseract")) {
		return tesseractProvider{}, nil
	}
	return nil, errors.New("unsupported OCR provider: " + settings.Provider)
}

func (r *Redactor) runOCRSearch(queries []models.RedactionTextQuery, settings models.OCRSettings) ([]models.RedactionRect, error) {
	if len(queries) == 0 {
		return nil, nil
	}
	p, err := getOCRProvider(settings)
	if err != nil {
		return nil, err
	}
	words, err := p.ExtractWords(r.pdfBytes, settings)
	if err != nil {
		return nil, err
	}
	var rects []models.RedactionRect
	// PERF-109: pre-compute lowered terms outside the word loop
	loweredTerms := make([]string, len(queries))
	for i, q := range queries {
		loweredTerms[i] = trimSpace(strings.ToLower(q.Text))
	}
	for _, w := range words {
		for _, term := range loweredTerms {
			if term == "" {
				continue
			}
			if containsFold(w.Text, term) {
				rects = append(rects, models.RedactionRect{
					PageNum: w.PageNum,
					X:       w.X,
					Y:       w.Y,
					Width:   w.Width,
					Height:  w.Height,
				})
				break
			}
		}
	}
	return rects, nil
}

func (tesseractProvider) ExtractWords(pdfBytes []byte, settings models.OCRSettings) ([]ocrWord, error) {
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return nil, errors.New("pdftoppm command not found for OCR pipeline")
	}
	if _, err := exec.LookPath("tesseract"); err != nil {
		return nil, errors.New("tesseract command not found for OCR pipeline")
	}

	r, err := NewRedactor(pdfBytes)
	if err != nil {
		return nil, err
	}
	info, err := r.GetPageInfo()
	if err != nil {
		return nil, err
	}

	tmpDir, err := os.MkdirTemp("", "gopdfsuit-ocr-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	pdfPath := filepath.Join(tmpDir, "input.pdf")
	if err := os.WriteFile(pdfPath, pdfBytes, 0o600); err != nil {
		return nil, err
	}

	lang := trimSpace(settings.Language)
	if lang == "" {
		lang = "eng"
	}

	// PERF-123: nil slice grows on demand (or pre-size when pages known)
	words := make([]ocrWord, 0, info.TotalPages*32)
	var pageTmp [20]byte
	// PERF-55: larger scanner buffer for long TSV lines
	scanBuf := make([]byte, 0, 64*1024)
	for page := 1; page <= info.TotalPages; page++ {
		pageStr := string(strconv.AppendInt(pageTmp[:0], int64(page), 10))
		imgBase := filepath.Join(tmpDir, "page-"+pageStr)
		imgPath := imgBase + ".png"
		pdftoppmCmd := exec.Command("pdftoppm", "-f", pageStr, "-l", pageStr, "-singlefile", "-png", pdfPath, imgBase)
		if out, err := pdftoppmCmd.CombinedOutput(); err != nil {
			return nil, errors.Join(errors.New("pdftoppm failed on page "+pageStr+": "+string(out)), err)
		}

		imgFile, err := os.Open(imgPath)
		if err != nil {
			return nil, err
		}
		cfg, _, err := image.DecodeConfig(imgFile)
		_ = imgFile.Close()
		if err != nil {
			return nil, err
		}

		tsvCmd := exec.Command("tesseract", imgPath, "stdout", "tsv", "-l", lang)
		tsvOut, err := tsvCmd.CombinedOutput()
		if err != nil {
			return nil, errors.Join(errors.New("tesseract failed on page "+pageStr+": "+string(tsvOut)), err)
		}

		pageDim := info.Pages[page-1]
		sx := pageDim.Width / float64(cfg.Width)
		sy := pageDim.Height / float64(cfg.Height)

		scanner := bufio.NewScanner(bytes.NewReader(tsvOut))
		scanner.Buffer(scanBuf, 1024*1024)
		lineNo := 0
		for scanner.Scan() {
			line := scanner.Text()
			lineNo++
			if lineNo == 1 {
				continue // header
			}
			var cols [12]string
			col := 0
			start := 0
			for i := 0; i <= len(line) && col < 12; i++ {
				if i == len(line) || line[i] == '\t' {
					cols[col] = line[start:i]
					col++
					start = i + 1
					if i == len(line) {
						break
					}
				}
			}
			if col < 12 {
				continue
			}
			text := trimSpace(cols[11])
			if text == "" {
				continue
			}
			left, errL := strconv.ParseFloat(cols[6], 64)
			top, errT := strconv.ParseFloat(cols[7], 64)
			w, errW := strconv.ParseFloat(cols[8], 64)
			h, errH := strconv.ParseFloat(cols[9], 64)
			if errL != nil || errT != nil || errW != nil || errH != nil {
				continue
			}
			words = append(words, ocrWord{
				PageNum: page,
				X:       left * sx,
				Y:       pageDim.Height - ((top + h) * sy),
				Width:   w * sx,
				Height:  h * sy,
				Text:    text,
			})
		}
		if err := scanner.Err(); err != nil {
			return nil, err
		}
	}

	return words, nil
}

// containsFold reports whether substr is within s, case-insensitively, without allocation.
// PERF-112: avoids strings.ToLower allocation by using EqualFold in a sliding window.
func containsFold(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if strings.EqualFold(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}
