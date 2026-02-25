// Package signature provides digital signature support for PDF documents.
package signature

// SignaturePageContext abstracts the PageManager for signature operations,
// breaking the circular dependency between the root pdf package and this subpackage.
//nolint:revive // exported
type SignaturePageContext interface {
	// AllocObjectID allocates and returns a new PDF object ID.
	AllocObjectID() int
	// SetExtraObject stores an extra PDF object dictionary by ID.
	SetExtraObject(id int, content string)
	// AppendPageAnnot appends an annotation object ID to the given page index.
	AppendPageAnnot(pageIndex int, annotID int)
	// GetMargins returns the page margins.
	GetMargins() PageMargins
	// FontHas returns true if the named font is registered.
	FontHas(name string) bool
	// FontMarkChars marks characters as used for font subsetting.
	FontMarkChars(name string, text string)
	// EncodeTextForFont encodes text using the named custom font.
	EncodeTextForFont(fontName, text string) string
}

// PageMargins holds margin values relevant to signature positioning.
type PageMargins struct {
	Right  float64
	Bottom float64
}

// PageDimensions holds page width/height for signature layout.
type PageDimensions struct {
	Width  float64
	Height float64
}
