package models

// PageDetail represents the dimensions of a single PDF page with its number
type PageDetail struct {
	PageNum int     `json:"pageNum"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
}

// PageInfo contains metadata about PDF pages
type PageInfo struct {
	TotalPages int          `json:"totalPages"`
	Pages      []PageDetail `json:"pages"`
}

// TextPosition represents the position of text on a page
type TextPosition struct {
	Text   string  `json:"text"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// RedactionRect represents a region to redact
type RedactionRect struct {
	PageNum int     `json:"pageNum"` // 1-based
	X       float64 `json:"x"`
	Y       float64 `json:"y"`
	Width   float64 `json:"width"`
	Height  float64 `json:"height"`
}

// RedactionTextQuery describes text-based redaction criteria.
type RedactionTextQuery struct {
	Text string `json:"text"`
}

// ApplyRedactionOptions represents a unified redaction request.
type ApplyRedactionOptions struct {
	Blocks     []RedactionRect      `json:"blocks,omitempty"`
	TextSearch []RedactionTextQuery `json:"textSearch,omitempty"`
	Mode       string               `json:"mode,omitempty"`     // secure_required | visual_allowed
	Password   string               `json:"password,omitempty"` // reserved for encrypted inputs
	OCR        *OCRSettings         `json:"ocr,omitempty"`
}

// OCRSettings is an extension point for OCR providers.
type OCRSettings struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider,omitempty"`
	Language string `json:"language,omitempty"`
}

// PageCapability describes whether a page contains text or image-like content.
type PageCapability struct {
	PageNum   int    `json:"pageNum"`
	Type      string `json:"type"` // text | image_only | mixed | unknown
	HasText   bool   `json:"hasText"`
	HasImage  bool   `json:"hasImage"`
	OCREnable bool   `json:"ocrEnabled"`
	Note      string `json:"note,omitempty"`
}

// RedactionApplyReport provides explicit safety/capability metadata.
type RedactionApplyReport struct {
	Mode              string           `json:"mode"`
	SecurityOutcome   string           `json:"securityOutcome"` // secure|visual_only|failed
	AppliedSecure     bool             `json:"appliedSecure"`
	AppliedVisual     bool             `json:"appliedVisual"`
	GeneratedRects    int              `json:"generatedRects"`
	AppliedRectangles int              `json:"appliedRectangles"`
	MatchedTextCount  int              `json:"matchedTextCount"`
	Capabilities      []PageCapability `json:"capabilities,omitempty"`
	UnsupportedPages  []int            `json:"unsupportedPages,omitempty"`
	Warnings          []string         `json:"warnings,omitempty"`
}
