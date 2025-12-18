package pdf

import (
	"fmt"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// PDF 2.0 compliant font definitions for the standard 14 fonts
// These include FirstChar, LastChar, Widths, and FontDescriptor as required by Arlington Model

// FontMetrics holds the complete metrics for a standard font
type FontMetrics struct {
	BaseFont       string
	FirstChar      int
	LastChar       int
	Widths         []int
	FontDescriptor FontDescriptor
}

// FontDescriptor holds font descriptor information
type FontDescriptor struct {
	FontName    string
	Flags       int
	FontBBox    [4]int
	ItalicAngle int
	Ascent      int
	Descent     int
	CapHeight   int
	StemV       int
	XHeight     int
}

// Standard Helvetica widths for WinAnsiEncoding (chars 32-255)
// These are the actual Adobe Helvetica metrics
// Note: helveticaWidths is already defined in xfdf.go, so we use a function to get it
var standardHelveticaWidths = []int{
	278, 278, 355, 556, 556, 889, 667, 191, 333, 333, 389, 584, 278, 333, 278, 278, // 32-47
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 278, 278, 584, 584, 584, 556, // 48-63
	1015, 667, 667, 722, 722, 667, 611, 778, 722, 278, 500, 667, 556, 833, 722, 778, // 64-79
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 278, 278, 278, 469, 556, // 80-95
	333, 556, 556, 500, 556, 556, 278, 556, 556, 222, 222, 500, 222, 833, 556, 556, // 96-111
	556, 556, 333, 500, 278, 556, 500, 722, 500, 500, 500, 334, 260, 334, 584, 350, // 112-127
	556, 350, 222, 556, 333, 1000, 556, 556, 333, 1000, 667, 333, 1000, 350, 611, 350, // 128-143
	350, 222, 222, 333, 333, 350, 556, 1000, 333, 1000, 500, 333, 944, 350, 500, 667, // 144-159
	278, 333, 556, 556, 556, 556, 260, 556, 333, 737, 370, 556, 584, 333, 737, 333, // 160-175
	400, 584, 333, 333, 333, 556, 537, 278, 333, 333, 365, 556, 834, 834, 834, 611, // 176-191
	667, 667, 667, 667, 667, 667, 1000, 722, 667, 667, 667, 667, 278, 278, 278, 278, // 192-207
	722, 722, 778, 778, 778, 778, 778, 584, 778, 722, 722, 722, 722, 667, 667, 611, // 208-223
	556, 556, 556, 556, 556, 556, 889, 500, 556, 556, 556, 556, 278, 278, 278, 278, // 224-239
	556, 556, 556, 556, 556, 556, 556, 584, 611, 556, 556, 556, 556, 500, 556, 500, // 240-255
}

// Helvetica Bold widths
var helveticaBoldWidths = []int{
	278, 333, 474, 556, 556, 889, 722, 238, 333, 333, 389, 584, 278, 333, 278, 278, // 32-47
	556, 556, 556, 556, 556, 556, 556, 556, 556, 556, 333, 333, 584, 584, 584, 611, // 48-63
	975, 722, 722, 722, 722, 667, 611, 778, 722, 278, 556, 722, 611, 833, 722, 778, // 64-79
	667, 778, 722, 667, 611, 722, 667, 944, 667, 667, 611, 333, 278, 333, 584, 556, // 80-95
	333, 556, 611, 556, 611, 556, 333, 611, 611, 278, 278, 556, 278, 889, 611, 611, // 96-111
	611, 611, 389, 556, 333, 611, 556, 778, 556, 556, 500, 389, 280, 389, 584, 350, // 112-127
	556, 350, 278, 556, 500, 1000, 556, 556, 333, 1000, 667, 333, 1000, 350, 611, 350, // 128-143
	350, 278, 278, 500, 500, 350, 556, 1000, 333, 1000, 556, 333, 944, 350, 500, 667, // 144-159
	278, 333, 556, 556, 556, 556, 280, 556, 333, 737, 370, 556, 584, 333, 737, 333, // 160-175
	400, 584, 333, 333, 333, 611, 556, 278, 333, 333, 365, 556, 834, 834, 834, 611, // 176-191
	722, 722, 722, 722, 722, 722, 1000, 722, 667, 667, 667, 667, 278, 278, 278, 278, // 192-207
	722, 722, 778, 778, 778, 778, 778, 584, 778, 722, 722, 722, 722, 667, 667, 611, // 208-223
	556, 556, 556, 556, 556, 556, 889, 556, 556, 556, 556, 556, 278, 278, 278, 278, // 224-239
	611, 611, 611, 611, 611, 611, 611, 584, 611, 611, 611, 611, 611, 556, 611, 556, // 240-255
}

// Helvetica Oblique widths (same as regular Helvetica)
var standardHelveticaObliqueWidths = standardHelveticaWidths

// Helvetica Bold Oblique widths (same as Helvetica Bold)
var standardHelveticaBoldObliqueWidths = helveticaBoldWidths

// Times Roman widths for WinAnsiEncoding (chars 32-255)
var timesRomanWidths = []int{
	250, 333, 408, 500, 500, 833, 778, 180, 333, 333, 500, 564, 250, 333, 250, 278, // 32-47
	500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 278, 278, 564, 564, 564, 444, // 48-63
	921, 722, 667, 667, 722, 611, 556, 722, 722, 333, 389, 722, 611, 889, 722, 722, // 64-79
	556, 722, 667, 556, 611, 722, 722, 944, 722, 722, 611, 333, 278, 333, 469, 500, // 80-95
	333, 444, 500, 444, 500, 444, 333, 500, 500, 278, 278, 500, 278, 778, 500, 500, // 96-111
	500, 500, 333, 389, 278, 500, 500, 722, 500, 500, 444, 480, 200, 480, 541, 350, // 112-127
	500, 350, 333, 500, 444, 1000, 500, 500, 333, 1000, 556, 333, 889, 350, 611, 350, // 128-143
	350, 333, 333, 444, 444, 350, 500, 1000, 333, 980, 389, 333, 722, 350, 444, 722, // 144-159
	250, 333, 500, 500, 500, 500, 200, 500, 333, 760, 276, 500, 564, 333, 760, 333, // 160-175
	400, 564, 300, 300, 333, 500, 453, 250, 333, 300, 310, 500, 750, 750, 750, 444, // 176-191
	722, 722, 722, 722, 722, 722, 889, 667, 611, 611, 611, 611, 333, 333, 333, 333, // 192-207
	722, 722, 722, 722, 722, 722, 722, 564, 722, 722, 722, 722, 722, 722, 556, 500, // 208-223
	444, 444, 444, 444, 444, 444, 667, 444, 444, 444, 444, 444, 278, 278, 278, 278, // 224-239
	500, 500, 500, 500, 500, 500, 500, 564, 500, 500, 500, 500, 500, 500, 500, 500, // 240-255
}

// Times Bold widths
var timesBoldWidths = []int{
	250, 333, 555, 500, 500, 1000, 833, 278, 333, 333, 500, 570, 250, 333, 250, 278, // 32-47
	500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 333, 333, 570, 570, 570, 500, // 48-63
	930, 722, 667, 722, 722, 667, 611, 778, 778, 389, 500, 778, 667, 944, 722, 778, // 64-79
	611, 778, 722, 556, 667, 722, 722, 1000, 722, 722, 667, 333, 278, 333, 581, 500, // 80-95
	333, 500, 556, 444, 556, 444, 333, 500, 556, 278, 333, 556, 278, 833, 556, 500, // 96-111
	556, 556, 444, 389, 333, 556, 500, 722, 500, 500, 444, 394, 220, 394, 520, 350, // 112-127
	500, 350, 333, 500, 500, 1000, 500, 500, 333, 1000, 556, 333, 1000, 350, 667, 350, // 128-143
	350, 333, 333, 500, 500, 350, 500, 1000, 333, 1000, 389, 333, 722, 350, 444, 722, // 144-159
	250, 333, 500, 500, 500, 500, 220, 500, 333, 747, 300, 500, 570, 333, 747, 333, // 160-175
	400, 570, 300, 300, 333, 556, 540, 250, 333, 300, 330, 500, 750, 750, 750, 500, // 176-191
	722, 722, 722, 722, 722, 722, 1000, 722, 667, 667, 667, 667, 389, 389, 389, 389, // 192-207
	722, 722, 778, 778, 778, 778, 778, 570, 778, 722, 722, 722, 722, 722, 611, 556, // 208-223
	500, 500, 500, 500, 500, 500, 722, 444, 444, 444, 444, 444, 278, 278, 278, 278, // 224-239
	500, 556, 500, 500, 500, 500, 500, 570, 500, 556, 556, 556, 556, 500, 556, 500, // 240-255
}

// Times Italic widths
var timesItalicWidths = []int{
	250, 333, 420, 500, 500, 833, 778, 214, 333, 333, 500, 675, 250, 333, 250, 278, // 32-47
	500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 333, 333, 675, 675, 675, 500, // 48-63
	920, 611, 611, 667, 722, 611, 611, 722, 722, 333, 444, 667, 556, 833, 667, 722, // 64-79
	611, 722, 611, 500, 556, 722, 611, 833, 611, 556, 556, 389, 278, 389, 422, 500, // 80-95
	333, 500, 500, 444, 500, 444, 278, 500, 500, 278, 278, 444, 278, 722, 500, 500, // 96-111
	500, 500, 389, 389, 278, 500, 444, 667, 444, 444, 389, 400, 275, 400, 541, 350, // 112-127
	500, 350, 333, 500, 556, 889, 500, 500, 333, 1000, 500, 333, 944, 350, 556, 350, // 128-143
	350, 333, 333, 556, 556, 350, 500, 889, 333, 980, 389, 333, 667, 350, 389, 556, // 144-159
	250, 389, 500, 500, 500, 500, 275, 500, 333, 760, 276, 500, 675, 333, 760, 333, // 160-175
	400, 675, 300, 300, 333, 500, 523, 250, 333, 300, 310, 500, 750, 750, 750, 500, // 176-191
	611, 611, 611, 611, 611, 611, 889, 667, 611, 611, 611, 611, 333, 333, 333, 333, // 192-207
	722, 667, 722, 722, 722, 722, 722, 675, 722, 722, 722, 722, 722, 556, 611, 500, // 208-223
	500, 500, 500, 500, 500, 500, 667, 444, 444, 444, 444, 444, 278, 278, 278, 278, // 224-239
	500, 500, 500, 500, 500, 500, 500, 675, 500, 500, 500, 500, 500, 444, 500, 444, // 240-255
}

// Times Bold Italic widths
var timesBoldItalicWidths = []int{
	250, 389, 555, 500, 500, 833, 778, 278, 333, 333, 500, 570, 250, 333, 250, 278, // 32-47
	500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 333, 333, 570, 570, 570, 500, // 48-63
	832, 667, 667, 667, 722, 667, 667, 722, 778, 389, 500, 667, 611, 889, 722, 722, // 64-79
	611, 722, 667, 556, 611, 722, 667, 889, 667, 611, 611, 333, 278, 333, 570, 500, // 80-95
	333, 500, 500, 444, 500, 444, 333, 500, 556, 278, 278, 500, 278, 778, 556, 500, // 96-111
	500, 500, 389, 389, 278, 556, 444, 667, 500, 444, 389, 348, 220, 348, 570, 350, // 112-127
	500, 350, 333, 500, 500, 1000, 500, 500, 333, 1000, 556, 333, 944, 350, 611, 350, // 128-143
	350, 333, 333, 500, 500, 350, 500, 1000, 333, 1000, 389, 333, 722, 350, 389, 611, // 144-159
	250, 389, 500, 500, 500, 500, 220, 500, 333, 747, 266, 500, 606, 333, 747, 333, // 160-175
	400, 570, 300, 300, 333, 576, 500, 250, 333, 300, 300, 500, 750, 750, 750, 500, // 176-191
	667, 667, 667, 667, 667, 667, 944, 667, 667, 667, 667, 667, 389, 389, 389, 389, // 192-207
	722, 722, 722, 722, 722, 722, 722, 570, 722, 722, 722, 722, 722, 611, 611, 500, // 208-223
	500, 500, 500, 500, 500, 500, 722, 444, 444, 444, 444, 444, 278, 278, 278, 278, // 224-239
	500, 556, 500, 500, 500, 500, 500, 570, 500, 556, 556, 556, 556, 444, 500, 444, // 240-255
}

// Courier widths (monospace - all characters same width)
var courierWidths = []int{
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 32-47
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 48-63
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 64-79
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 80-95
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 96-111
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 112-127
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 128-143
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 144-159
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 160-175
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 176-191
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 192-207
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 208-223
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 224-239
	600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, 600, // 240-255
}

// Symbol font widths
var symbolWidths = []int{
	250, 333, 713, 500, 549, 833, 778, 439, 333, 333, 500, 549, 250, 549, 250, 278, // 32-47
	500, 500, 500, 500, 500, 500, 500, 500, 500, 500, 278, 278, 549, 549, 549, 444, // 48-63
	549, 722, 667, 722, 612, 611, 763, 603, 722, 333, 631, 722, 686, 889, 722, 722, // 64-79
	768, 741, 556, 592, 611, 690, 439, 768, 645, 795, 611, 333, 863, 333, 658, 500, // 80-95
	500, 631, 549, 549, 494, 439, 521, 411, 603, 329, 603, 549, 549, 576, 521, 549, // 96-111
	549, 521, 549, 603, 439, 576, 713, 686, 493, 686, 494, 480, 200, 480, 549, 350, // 112-127
	500, 350, 333, 500, 549, 1000, 500, 500, 333, 1000, 500, 333, 500, 350, 500, 350, // 128-143
	350, 333, 333, 549, 549, 350, 500, 1000, 333, 1000, 500, 333, 500, 350, 500, 500, // 144-159
	250, 620, 247, 549, 167, 713, 500, 753, 753, 753, 753, 1042, 987, 603, 987, 603, // 160-175
	400, 549, 411, 549, 549, 713, 494, 460, 549, 549, 549, 549, 1000, 603, 1000, 658, // 176-191
	823, 686, 795, 987, 768, 768, 823, 768, 768, 713, 713, 713, 713, 713, 713, 713, // 192-207
	768, 713, 790, 790, 890, 823, 549, 250, 713, 603, 603, 1042, 987, 603, 987, 603, // 208-223
	494, 329, 790, 790, 786, 713, 384, 384, 384, 384, 384, 384, 494, 494, 494, 494, // 224-239
	329, 274, 686, 686, 686, 384, 384, 384, 384, 384, 384, 494, 494, 494, 250, 250, // 240-255
}

// ZapfDingbats widths
var zapfDingbatsWidths = []int{
	278, 974, 961, 974, 980, 719, 789, 790, 791, 690, 960, 939, 549, 855, 911, 933, // 32-47
	911, 945, 974, 755, 846, 762, 761, 571, 677, 763, 760, 759, 754, 494, 552, 537, // 48-63
	577, 692, 786, 788, 788, 790, 793, 794, 816, 823, 789, 841, 823, 833, 816, 831, // 64-79
	923, 744, 723, 749, 790, 792, 695, 776, 768, 792, 759, 707, 708, 682, 701, 826, // 80-95
	815, 789, 789, 707, 687, 696, 689, 786, 787, 713, 791, 785, 791, 873, 761, 762, // 96-111
	762, 759, 759, 892, 892, 788, 784, 438, 138, 277, 415, 392, 392, 668, 668, 350, // 112-127
	278, 350, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, // 128-143
	278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, 278, // 144-159
	278, 732, 544, 544, 910, 667, 760, 760, 776, 595, 694, 626, 788, 788, 788, 788, // 160-175
	788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, // 176-191
	788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, 788, // 192-207
	788, 788, 788, 788, 894, 838, 1016, 458, 748, 924, 748, 918, 927, 928, 928, 834, // 208-223
	873, 828, 924, 924, 917, 930, 931, 463, 883, 836, 836, 867, 867, 696, 696, 874, // 224-239
	278, 874, 760, 946, 771, 865, 771, 888, 967, 888, 831, 873, 927, 970, 918, 278, // 240-255
}

// GetFontMetrics returns the complete font metrics for a given font name
func GetFontMetrics(fontName string) FontMetrics {
	switch fontName {
	case "Helvetica":
		return FontMetrics{
			BaseFont:  "Helvetica",
			FirstChar: 32,
			LastChar:  255,
			Widths:    standardHelveticaWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Helvetica",
				Flags:       32, // Non-symbolic
				FontBBox:    [4]int{-166, -225, 1000, 931},
				ItalicAngle: 0,
				Ascent:      718,
				Descent:     -207,
				CapHeight:   718,
				StemV:       88,
				XHeight:     523,
			},
		}
	case "Helvetica-Bold":
		return FontMetrics{
			BaseFont:  "Helvetica-Bold",
			FirstChar: 32,
			LastChar:  255,
			Widths:    helveticaBoldWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Helvetica-Bold",
				Flags:       32 | 262144, // Non-symbolic + ForceBold
				FontBBox:    [4]int{-170, -228, 1003, 962},
				ItalicAngle: 0,
				Ascent:      718,
				Descent:     -207,
				CapHeight:   718,
				StemV:       140,
				XHeight:     532,
			},
		}
	case "Helvetica-Oblique":
		return FontMetrics{
			BaseFont:  "Helvetica-Oblique",
			FirstChar: 32,
			LastChar:  255,
			Widths:    standardHelveticaObliqueWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Helvetica-Oblique",
				Flags:       32 | 64, // Non-symbolic + Italic
				FontBBox:    [4]int{-170, -225, 1116, 931},
				ItalicAngle: -12,
				Ascent:      718,
				Descent:     -207,
				CapHeight:   718,
				StemV:       88,
				XHeight:     523,
			},
		}
	case "Helvetica-BoldOblique":
		return FontMetrics{
			BaseFont:  "Helvetica-BoldOblique",
			FirstChar: 32,
			LastChar:  255,
			Widths:    standardHelveticaBoldObliqueWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Helvetica-BoldOblique",
				Flags:       32 | 64 | 262144, // Non-symbolic + Italic + ForceBold
				FontBBox:    [4]int{-174, -228, 1114, 962},
				ItalicAngle: -12,
				Ascent:      718,
				Descent:     -207,
				CapHeight:   718,
				StemV:       140,
				XHeight:     532,
			},
		}
	// Times family
	case "Times-Roman":
		return FontMetrics{
			BaseFont:  "Times-Roman",
			FirstChar: 32,
			LastChar:  255,
			Widths:    timesRomanWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Times-Roman",
				Flags:       34, // Serif + Non-symbolic
				FontBBox:    [4]int{-168, -218, 1000, 898},
				ItalicAngle: 0,
				Ascent:      683,
				Descent:     -217,
				CapHeight:   662,
				StemV:       84,
				XHeight:     450,
			},
		}
	case "Times-Bold":
		return FontMetrics{
			BaseFont:  "Times-Bold",
			FirstChar: 32,
			LastChar:  255,
			Widths:    timesBoldWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Times-Bold",
				Flags:       34 | 262144, // Serif + Non-symbolic + ForceBold
				FontBBox:    [4]int{-168, -218, 1000, 935},
				ItalicAngle: 0,
				Ascent:      683,
				Descent:     -217,
				CapHeight:   676,
				StemV:       139,
				XHeight:     461,
			},
		}
	case "Times-Italic":
		return FontMetrics{
			BaseFont:  "Times-Italic",
			FirstChar: 32,
			LastChar:  255,
			Widths:    timesItalicWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Times-Italic",
				Flags:       34 | 64, // Serif + Non-symbolic + Italic
				FontBBox:    [4]int{-169, -217, 1010, 883},
				ItalicAngle: -15,
				Ascent:      683,
				Descent:     -217,
				CapHeight:   653,
				StemV:       76,
				XHeight:     442,
			},
		}
	case "Times-BoldItalic":
		return FontMetrics{
			BaseFont:  "Times-BoldItalic",
			FirstChar: 32,
			LastChar:  255,
			Widths:    timesBoldItalicWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Times-BoldItalic",
				Flags:       34 | 64 | 262144, // Serif + Non-symbolic + Italic + ForceBold
				FontBBox:    [4]int{-200, -218, 996, 921},
				ItalicAngle: -15,
				Ascent:      683,
				Descent:     -217,
				CapHeight:   669,
				StemV:       121,
				XHeight:     462,
			},
		}
	// Courier family
	case "Courier":
		return FontMetrics{
			BaseFont:  "Courier",
			FirstChar: 32,
			LastChar:  255,
			Widths:    courierWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Courier",
				Flags:       33, // FixedPitch + Non-symbolic
				FontBBox:    [4]int{-23, -250, 715, 805},
				ItalicAngle: 0,
				Ascent:      629,
				Descent:     -157,
				CapHeight:   562,
				StemV:       51,
				XHeight:     426,
			},
		}
	case "Courier-Bold":
		return FontMetrics{
			BaseFont:  "Courier-Bold",
			FirstChar: 32,
			LastChar:  255,
			Widths:    courierWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Courier-Bold",
				Flags:       33 | 262144, // FixedPitch + Non-symbolic + ForceBold
				FontBBox:    [4]int{-113, -250, 749, 801},
				ItalicAngle: 0,
				Ascent:      629,
				Descent:     -157,
				CapHeight:   562,
				StemV:       106,
				XHeight:     439,
			},
		}
	case "Courier-Oblique":
		return FontMetrics{
			BaseFont:  "Courier-Oblique",
			FirstChar: 32,
			LastChar:  255,
			Widths:    courierWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Courier-Oblique",
				Flags:       33 | 64, // FixedPitch + Non-symbolic + Italic
				FontBBox:    [4]int{-27, -250, 849, 805},
				ItalicAngle: -12,
				Ascent:      629,
				Descent:     -157,
				CapHeight:   562,
				StemV:       51,
				XHeight:     426,
			},
		}
	case "Courier-BoldOblique":
		return FontMetrics{
			BaseFont:  "Courier-BoldOblique",
			FirstChar: 32,
			LastChar:  255,
			Widths:    courierWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Courier-BoldOblique",
				Flags:       33 | 64 | 262144, // FixedPitch + Non-symbolic + Italic + ForceBold
				FontBBox:    [4]int{-57, -250, 869, 801},
				ItalicAngle: -12,
				Ascent:      629,
				Descent:     -157,
				CapHeight:   562,
				StemV:       106,
				XHeight:     439,
			},
		}
	// Symbol fonts
	case "Symbol":
		return FontMetrics{
			BaseFont:  "Symbol",
			FirstChar: 32,
			LastChar:  255,
			Widths:    symbolWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "Symbol",
				Flags:       4, // Symbolic
				FontBBox:    [4]int{-180, -293, 1090, 1010},
				ItalicAngle: 0,
				Ascent:      800,
				Descent:     -200,
				CapHeight:   800,
				StemV:       85,
				XHeight:     500,
			},
		}
	case "ZapfDingbats":
		return FontMetrics{
			BaseFont:  "ZapfDingbats",
			FirstChar: 32,
			LastChar:  255,
			Widths:    zapfDingbatsWidths,
			FontDescriptor: FontDescriptor{
				FontName:    "ZapfDingbats",
				Flags:       4, // Symbolic
				FontBBox:    [4]int{-1, -143, 981, 820},
				ItalicAngle: 0,
				Ascent:      800,
				Descent:     -200,
				CapHeight:   800,
				StemV:       90,
				XHeight:     500,
			},
		}
	default:
		// Default to Helvetica
		return GetFontMetrics("Helvetica")
	}
}

// GenerateFontObject creates a complete PDF 2.0 compliant font object
// Returns the font object string and the FontDescriptor object ID used
func GenerateFontObject(fontName string, fontObjectID, fontDescriptorID, widthsArrayID int) string {
	metrics := GetFontMetrics(fontName)

	// Build compact font dictionary without deprecated /Name field for PDF 2.0
	return fmt.Sprintf("%d 0 obj\n<</Type/Font/Subtype/Type1/BaseFont/%s/Encoding/WinAnsiEncoding/FirstChar %d/LastChar %d/Widths %d 0 R/FontDescriptor %d 0 R>>\nendobj\n",
		fontObjectID, metrics.BaseFont, metrics.FirstChar, metrics.LastChar, widthsArrayID, fontDescriptorID)
}

// GenerateFontDescriptorObject creates a FontDescriptor object
func GenerateFontDescriptorObject(fontName string, objectID int) string {
	metrics := GetFontMetrics(fontName)
	fd := metrics.FontDescriptor

	// Compact format - single line
	return fmt.Sprintf("%d 0 obj\n<</Type/FontDescriptor/FontName/%s/Flags %d/FontBBox[%d %d %d %d]/ItalicAngle %d/Ascent %d/Descent %d/CapHeight %d/StemV %d/XHeight %d>>\nendobj\n",
		objectID, fd.FontName, fd.Flags, fd.FontBBox[0], fd.FontBBox[1], fd.FontBBox[2], fd.FontBBox[3],
		fd.ItalicAngle, fd.Ascent, fd.Descent, fd.CapHeight, fd.StemV, fd.XHeight)
}

// GenerateWidthsArrayObject creates a Widths array object
func GenerateWidthsArrayObject(fontName string, objectID int) string {
	metrics := GetFontMetrics(fontName)

	var widthsArray strings.Builder
	widthsArray.WriteString(fmt.Sprintf("%d 0 obj\n[", objectID))

	// Compact format - no newlines, minimal spacing
	for i, w := range metrics.Widths {
		if i > 0 {
			widthsArray.WriteString(" ")
		}
		widthsArray.WriteString(fmt.Sprintf("%d", w))
	}

	widthsArray.WriteString("]\nendobj\n")

	return widthsArray.String()
}

// GetHelveticaFontResourceString returns a complete inline font resource for XObjects
// This is used in form field appearance streams - optimized for minimal size
// arlingtonCompatible: if true, includes full font metrics for PDF 2.0 compliance
func GetHelveticaFontResourceString() string {
	metrics := GetFontMetrics("Helvetica")

	// Build compact widths array inline (no extra spaces)
	var widths strings.Builder
	widths.WriteString("[")
	for i, w := range metrics.Widths {
		if i > 0 {
			widths.WriteString(" ")
		}
		widths.WriteString(fmt.Sprintf("%d", w))
	}
	widths.WriteString("]")

	// Build compact font dictionary with inline FontDescriptor
	return fmt.Sprintf(`<</Type/Font/Subtype/Type1/BaseFont/Helvetica/Encoding/WinAnsiEncoding/FirstChar %d/LastChar %d/Widths %s/FontDescriptor<</Type/FontDescriptor/FontName/Helvetica/Flags 32/FontBBox[-166 -225 1000 931]/ItalicAngle 0/Ascent 718/Descent -207/CapHeight 718/StemV 88/XHeight 523>>>>`,
		metrics.FirstChar, metrics.LastChar, widths.String())
}

// GetSimpleHelveticaFontResourceString returns a simple inline font resource for XObjects
// This is used when Arlington compatibility is OFF - minimal font definition
func GetSimpleHelveticaFontResourceString() string {
	return `<</Type/Font/Subtype/Type1/BaseFont/Helvetica>>`
}

// GenerateSimpleFontObject creates a simple font object (non-Arlington mode)
// This is the legacy format without FirstChar, LastChar, Widths, and FontDescriptor
func GenerateSimpleFontObject(fontName string, fontRef string, fontObjectID int) string {
	return fmt.Sprintf("%d 0 obj\n<< /Type /Font /Subtype /Type1 /Name %s /BaseFont /%s >>\nendobj\n",
		fontObjectID, fontRef, fontName)
}

// GetAvailableFonts returns the list of available fonts for PDF generation.
// This includes the standard PDF Type 1 fonts and commonly used fonts.
func GetAvailableFonts() []models.FontInfo {
	return []models.FontInfo{
		// Standard PDF Type 1 Fonts - Helvetica family (F1-F4)
		{ID: "Helvetica", Name: "Helvetica", DisplayName: "Helvetica", Reference: "/F1"},
		{ID: "Helvetica-Bold", Name: "Helvetica-Bold", DisplayName: "Helvetica Bold", Reference: "/F2"},
		{ID: "Helvetica-Oblique", Name: "Helvetica-Oblique", DisplayName: "Helvetica Italic", Reference: "/F3"},
		{ID: "Helvetica-BoldOblique", Name: "Helvetica-BoldOblique", DisplayName: "Helvetica Bold Italic", Reference: "/F4"},

		// Standard PDF Type 1 Fonts - Times family (F5-F8)
		{ID: "Times-Roman", Name: "Times-Roman", DisplayName: "Times Roman", Reference: "/F5"},
		{ID: "Times-Bold", Name: "Times-Bold", DisplayName: "Times Bold", Reference: "/F6"},
		{ID: "Times-Italic", Name: "Times-Italic", DisplayName: "Times Italic", Reference: "/F7"},
		{ID: "Times-BoldItalic", Name: "Times-BoldItalic", DisplayName: "Times Bold Italic", Reference: "/F8"},

		// Standard PDF Type 1 Fonts - Courier family (F9-F12)
		{ID: "Courier", Name: "Courier", DisplayName: "Courier", Reference: "/F9"},
		{ID: "Courier-Bold", Name: "Courier-Bold", DisplayName: "Courier Bold", Reference: "/F10"},
		{ID: "Courier-Oblique", Name: "Courier-Oblique", DisplayName: "Courier Italic", Reference: "/F11"},
		{ID: "Courier-BoldOblique", Name: "Courier-BoldOblique", DisplayName: "Courier Bold Italic", Reference: "/F12"},

		// Standard PDF Type 1 Fonts - Symbol and Decorative (F13-F14)
		{ID: "Symbol", Name: "Symbol", DisplayName: "Symbol", Reference: "/F13"},
		{ID: "ZapfDingbats", Name: "ZapfDingbats", DisplayName: "Zapf Dingbats", Reference: "/F14"},
	}
}
