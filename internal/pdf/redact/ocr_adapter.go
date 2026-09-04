package redact

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/png" // Register PNG decoder for image.DecodeConfig
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

type ocrWord struct {
	PageNum int
	X       float64
	Y       float64
	Width   float64
	Height  float64
	Text    string
}

// errOCRUnsupportedWASM is returned whenever OCR is requested under GOOS=js.
// pdftoppm/tesseract subprocesses plus a temp directory cannot exist in the
// browser, so OCR must fail fast here instead of reaching os/exec. The
// message carries "unsupported" so the pkg/gopdflib seam classifies it as
// CodeInvalidInput (invalid_input) through its legacy substring fallback
// (this package cannot import pkg/gopdflib: import cycle via internal/pdf).
var errOCRUnsupportedWASM = errors.New("redact: OCR unsupported in WASM (GOOS=js): pdftoppm/tesseract subprocesses cannot run in the browser; use the text path (FindTextOccurrences) or run OCR server-side")

// OCRProvider is an adapter interface for OCR backends.
type OCRProvider interface {
	ExtractWords(pdfBytes []byte, settings models.OCRSettings) ([]ocrWord, error)
}

type tesseractProvider struct{}

// ocrCommandTimeout bounds each pdftoppm/tesseract invocation so a crafted
// PDF cannot hang the pipeline forever. It is a var so tests can shrink it.
var ocrCommandTimeout = 5 * time.Minute

func ocrCommandContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), ocrCommandTimeout)
}

func getOCRProvider(settings models.OCRSettings) (OCRProvider, error) {
	provider := strings.ToLower(settings.Provider)
	for len(provider) > 0 && provider[0] <= ' ' {
		provider = provider[1:]
	}
	for len(provider) > 0 && provider[len(provider)-1] <= ' ' {
		provider = provider[:len(provider)-1]
	}
	if provider == "" || provider == "tesseract" {
		return tesseractProvider{}, nil
	}
	return nil, errors.New("unsupported OCR provider: " + settings.Provider)
}

func (r *Redactor) runOCRSearch(queries []models.RedactionTextQuery, settings models.OCRSettings) ([]models.RedactionRect, error) {
	if isWASM() {
		return nil, errOCRUnsupportedWASM
	}
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
	if len(words) == 0 {
		return nil, nil
	}

	// Pre-normalize query terms once (lowercase + trim) instead of per (word, query) pair.
	normTerms := make([]string, len(queries))
	for i, q := range queries {
		t := strings.ToLower(q.Text)
		// Manual no-alloc trim
		for len(t) > 0 && t[0] <= ' ' {
			t = t[1:]
		}
		for len(t) > 0 && t[len(t)-1] <= ' ' {
			t = t[:len(t)-1]
		}
		normTerms[i] = t
	}
	// Skip empty/whitespace-only queries up-front
	activeTerms := normTerms[:0]
	activeIdx := make([]int, 0, len(queries))
	for i, t := range normTerms {
		if t != "" {
			activeTerms = append(activeTerms, t)
			activeIdx = append(activeIdx, i)
		}
	}
	if len(activeTerms) == 0 {
		return nil, nil
	}

	var rects []models.RedactionRect
	for wi := range words {
		wLower := strings.ToLower(words[wi].Text)
		for k, term := range activeTerms {
			if strings.Contains(wLower, term) {
				rects = append(rects, models.RedactionRect{
					PageNum: words[wi].PageNum,
					X:       words[wi].X,
					Y:       words[wi].Y,
					Width:   words[wi].Width,
					Height:  words[wi].Height,
				})
				_ = activeIdx[k]
				break
			}
		}
	}
	return rects, nil
}

func (tesseractProvider) ExtractWords(pdfBytes []byte, settings models.OCRSettings) ([]ocrWord, error) {
	if isWASM() {
		return nil, errOCRUnsupportedWASM
	}
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

	lang := strings.TrimSpace(settings.Language)
	if lang == "" {
		lang = "eng"
	}

	words := make([]ocrWord, 0)
	var pageBuf [20]byte
	for page := 1; page <= info.TotalPages; page++ {
		pageStr := string(strconv.AppendInt(pageBuf[:0], int64(page), 10))
		imgBase := filepath.Join(tmpDir, "page-"+pageStr)
		imgPath := imgBase + ".png"
		pdftoppmCtx, pdftoppmCancel := ocrCommandContext()
		pdftoppmCmd := ocrCommand(pdftoppmCtx, "pdftoppm", "-f", pageStr, "-l", pageStr, "-singlefile", "-png", pdfPath, imgBase)
		out, pdftoppmErr := pdftoppmCmd.CombinedOutput()
		pdftoppmCancel()
		if pdftoppmErr != nil {
			return nil, fmt.Errorf("pdftoppm failed on page %d: %v (%s)", page, pdftoppmErr, string(out))
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

		tsvCtx, tsvCancel := ocrCommandContext()
		tsvCmd := ocrCommand(tsvCtx, "tesseract", imgPath, "stdout", "tsv", "-l", lang)
		tsvOut, err := tsvCmd.CombinedOutput()
		tsvCancel()
		if err != nil {
			return nil, fmt.Errorf("tesseract failed on page %d: %v (%s)", page, err, string(tsvOut))
		}

		pageDim := info.Pages[page-1]
		sx := pageDim.Width / float64(cfg.Width)
		sy := pageDim.Height / float64(cfg.Height)

		scanner := bufio.NewScanner(bytes.NewReader(tsvOut))
		scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024)
		lineNo := 0
		for scanner.Scan() {
			line := scanner.Bytes()
			lineNo++
			if lineNo == 1 {
				continue // header
			}
			cols := bytes.Split(line, []byte{'\t'})
			if len(cols) < 12 {
				continue
			}
			text := string(bytes.TrimSpace(cols[11]))
			if text == "" {
				continue
			}
			left, errL := strconv.ParseFloat(string(cols[6]), 64)
			top, errT := strconv.ParseFloat(string(cols[7]), 64)
			w, errW := strconv.ParseFloat(string(cols[8]), 64)
			h, errH := strconv.ParseFloat(string(cols[9]), 64)
			if errL != nil || errT != nil || errW != nil || errH != nil {
				continue
			}

			pdfX := left * sx
			pdfY := pageDim.Height - ((top + h) * sy)
			words = append(words, ocrWord{
				PageNum: page,
				X:       pdfX,
				Y:       pdfY,
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
