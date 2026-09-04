package handlers

import (
	"errors"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
)

// redactApplyForm carries the raw multipart text fields of a redact-apply
// request. It decouples parsing from Gin so the parser below is a pure
// function over values and errors, tested without an HTTP layer.
type redactApplyForm struct {
	mode       string
	password   string
	blocks     string
	textSearch string
	ocr        string
	// redactions is the legacy alias for blocks sent by the old frontend.
	redactions string
	// text is the legacy plain-text search field for one-shot apply.
	text string
}

// parseRedactApply parses blocks, textSearch (object array or plain string
// array), ocr, and the legacy redactions/text compat fields into apply
// options. It performs no I/O and imports no HTTP framework.
func parseRedactApply(form redactApplyForm) (models.ApplyRedactionOptions, error) {
	var options models.ApplyRedactionOptions
	options.Mode = fastTrimSpace(form.mode)
	options.Password = form.password

	if form.blocks != "" {
		if err := sonic.Unmarshal([]byte(form.blocks), &options.Blocks); err != nil {
			return options, errors.New("invalid blocks json")
		}
	}

	if form.textSearch != "" {
		textSearchBytes := []byte(form.textSearch)
		if err := sonic.Unmarshal(textSearchBytes, &options.TextSearch); err != nil {
			var plain []string
			if err2 := sonic.Unmarshal(textSearchBytes, &plain); err2 != nil {
				return options, errors.New("invalid textSearch json")
			}
			for _, text := range plain {
				text = fastTrimSpace(text)
				if text == "" {
					continue
				}
				options.TextSearch = append(options.TextSearch, models.RedactionTextQuery{Text: text})
			}
		}
	}

	if fastTrimSpace(form.ocr) != "" {
		var ocr models.OCRSettings
		if err := sonic.Unmarshal([]byte(form.ocr), &ocr); err != nil {
			return options, errors.New("invalid ocr json")
		}
		options.OCR = &ocr
	}

	// Backward compatibility: old frontend sends "redactions".
	if len(options.Blocks) == 0 && form.redactions != "" {
		if err := sonic.Unmarshal([]byte(form.redactions), &options.Blocks); err != nil {
			return options, errors.New("invalid redactions json")
		}
	}

	// Backward compatibility: allow plain text search field for one-shot apply.
	if len(options.TextSearch) == 0 {
		if searchText := fastTrimSpace(form.text); searchText != "" {
			terms := parseCommaSeparatedTerms(searchText)
			if len(terms) == 0 {
				terms = []string{searchText}
			}
			options.TextSearch = make([]models.RedactionTextQuery, 0, len(terms))
			for _, t := range terms {
				options.TextSearch = append(options.TextSearch, models.RedactionTextQuery{Text: t})
			}
		}
	}
	options.TextSearch = normalizeTextSearchQueries(options.TextSearch)
	return options, nil
}
