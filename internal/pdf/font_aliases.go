// Package pdf provides PDF generation and manipulation functionality.
// Font-related types have been extracted to the font subpackage.
// These type aliases maintain backward compatibility within the pdf package.
package pdf

import "github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf/font"

//nolint:revive // exported
// Type aliases for backward compatibility within this package.
// External consumers should use the font package directly.
type CustomFontRegistry = font.CustomFontRegistry
//nolint:revive // exported
type RegisteredFont = font.RegisteredFont
//nolint:revive // exported
type TTFFont = font.TTFFont
//nolint:revive // exported
type FontMetrics = font.FontMetrics
//nolint:revive // exported
type FontDescriptor = font.FontDescriptor
//nolint:revive // exported
type PDFAFontManager = font.PDFAFontManager
//nolint:revive // exported
type PDFAFontConfig = font.PDFAFontConfig

// Function aliases for backward compatibility within this package.
var (
	GetFontRegistry                      = font.GetFontRegistry
	NewFontRegistry                      = font.NewFontRegistry
	GetFontMetrics                       = font.GetFontMetrics
	GenerateFontObject                   = font.GenerateFontObject
	GenerateSimpleFontObject             = font.GenerateSimpleFontObject
	GenerateFontDescriptorObject         = font.GenerateFontDescriptorObject
	GenerateWidthsArrayObject            = font.GenerateWidthsArrayObject
	GenerateTrueTypeFontObjects          = font.GenerateTrueTypeFontObjects
	EncodeTextForCustomFont              = font.EncodeTextForCustomFont
	GetHelveticaFontResourceString       = font.GetHelveticaFontResourceString
	GetSimpleHelveticaFontResourceString = font.GetSimpleHelveticaFontResourceString
	GetAvailableFonts                    = font.GetAvailableFonts
	CompressFontData                     = font.CompressFontData
	IsCustomFont                         = font.IsCustomFont
	GetPDFAFontManager                   = font.GetPDFAFontManager
	IsStandardFont                       = font.IsStandardFont
	GetMappedFontName                    = font.GetMappedFontName
	GetLiberationFontPostScriptName      = font.GetLiberationFontPostScriptName
	ParseTTF                             = font.ParseTTF
	LoadTTFFromFile                      = font.LoadTTFFromFile
	LoadTTFFromData                      = font.LoadTTFFromData
	SubsetTTF                            = font.SubsetTTF
	SubsetTTFForText                     = font.SubsetTTFForText
	GetCompressBuffer                    = font.GetCompressBuffer
	GetZlibWriter                        = font.GetZlibWriter
	PutZlibWriter                        = font.PutZlibWriter
	CompressBufPool                      = &font.CompressBufPool
)

// Internal aliases for lower-case names previously used in the pdf package
var (
	getCompressBuffer   = font.GetCompressBuffer
	getZlibWriter       = font.GetZlibWriter
	putZlibWriter       = font.PutZlibWriter // Note: remote code uses getZlibWriter/putZlibWriter names
	compressBufPool     = &font.CompressBufPool
	generateCIDToGIDMap = font.GenerateCIDToGIDMap
)
