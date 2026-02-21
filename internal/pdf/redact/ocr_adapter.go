package redact

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
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
	ExtractWords(pdfBytes []byte, settings OCRSettings) ([]ocrWord, error)
}

type tesseractProvider struct{}

func getOCRProvider(settings OCRSettings) (OCRProvider, error) {
	provider := strings.TrimSpace(strings.ToLower(settings.Provider))
	if provider == "" || provider == "tesseract" {
		return tesseractProvider{}, nil
	}
	return nil, fmt.Errorf("unsupported OCR provider: %s", settings.Provider)
}

func runOCRSearch(pdfBytes []byte, queries []RedactionTextQuery, settings OCRSettings) ([]RedactionRect, error) {
	if len(queries) == 0 {
		return nil, nil
	}
	p, err := getOCRProvider(settings)
	if err != nil {
		return nil, err
	}
	words, err := p.ExtractWords(pdfBytes, settings)
	if err != nil {
		return nil, err
	}
	var rects []RedactionRect
	for _, w := range words {
		for _, q := range queries {
			term := strings.TrimSpace(strings.ToLower(q.Text))
			if term == "" {
				continue
			}
			if strings.Contains(strings.ToLower(w.Text), term) {
				rects = append(rects, RedactionRect{
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

func (tesseractProvider) ExtractWords(pdfBytes []byte, settings OCRSettings) ([]ocrWord, error) {
	if _, err := exec.LookPath("pdftoppm"); err != nil {
		return nil, errors.New("pdftoppm command not found for OCR pipeline")
	}
	if _, err := exec.LookPath("tesseract"); err != nil {
		return nil, errors.New("tesseract command not found for OCR pipeline")
	}

	info, err := GetPageInfo(pdfBytes)
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
	for page := 1; page <= info.TotalPages; page++ {
		imgBase := filepath.Join(tmpDir, fmt.Sprintf("page-%d", page))
		imgPath := imgBase + ".png"
		pdftoppmCmd := exec.Command("pdftoppm", "-f", strconv.Itoa(page), "-l", strconv.Itoa(page), "-singlefile", "-png", pdfPath, imgBase)
		if out, err := pdftoppmCmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("pdftoppm failed on page %d: %v (%s)", page, err, string(out))
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
			return nil, fmt.Errorf("tesseract failed on page %d: %v (%s)", page, err, string(tsvOut))
		}

		pageDim := info.Pages[page-1]
		sx := pageDim.Width / float64(cfg.Width)
		sy := pageDim.Height / float64(cfg.Height)

		scanner := bufio.NewScanner(bytes.NewReader(tsvOut))
		lineNo := 0
		for scanner.Scan() {
			line := scanner.Text()
			lineNo++
			if lineNo == 1 {
				continue // header
			}
			cols := strings.Split(line, "\t")
			if len(cols) < 12 {
				continue
			}
			text := strings.TrimSpace(cols[11])
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
