package pdf

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v4/internal/models"
	"github.com/chinmay-sawant/gopdfsuit/v4/internal/pdf/redact"
	"github.com/chinmay-sawant/gopdfsuit/v4/typstsyntax"
)

func TestTypstMathStress_RenderToContentStream(t *testing.T) {
	t.Parallel()

	expressions := []string{
		"$ integral_0^1 x^2 d x $",
		"$ integral_a^b frac(1, 1 + x^2) d x $",
		"$ integral_(0)^oo e^(-x) d x $",
		"$ integral.double_A f(x, y) d x d y $",
		"$ integral.triple_V rho(x, y, z) d V $",
		"$ integral.cont_C F dot d r $",
		"$ partial f / partial x $",
		"$ partial^2 u / partial x^2 $",
		"$ nabla dot F = 0 $",
		"$ nabla times F = 0 $",
		"$ sum_(k=1)^n k = frac(n(n+1), 2) $",
		"$ product_(i=1)^n i = n! $",
		"$ lim_(x -> 0) frac(sin x, x) = 1 $",
		"$ RR^3 subset.eq CC^3 $",
		"$ forall x in RR, exists y in RR: y >= x $",
		"$ alpha + beta = gamma $",
		"$ theta != pi $",
		"$ vec(1, 2, 3) $",
		"$ mat(1, 0; 0, 1) $",
		"$ sqrt(x^2 + y^2) $",
		"$ root(3, x) $",
	}

	ctx := &typstsyntax.RenderContext{
		X:         72,
		Y:         720,
		FontSize:  12,
		FontRef:   "/F1",
		TextColor: "0 0 0",
		EstimateWidth: func(text string, fontSize float64) float64 {
			return float64(len([]rune(text))) * fontSize * 0.5
		},
	}

	integralExprCount := 0
	for _, expr := range expressions {
		ast := typstsyntax.ParseMath(expr)
		if ast == nil {
			t.Fatalf("ParseMath(%q) returned nil", expr)
		}

		flattened := typstsyntax.FlattenToText(ast)
		if strings.TrimSpace(flattened) == "" {
			t.Fatalf("FlattenToText(%q) returned empty text", expr)
		}

		if strings.Contains(expr, "integral") {
			integralExprCount++
			if !containsAnyIntegralGlyph(flattened) {
				t.Fatalf("expression %q did not render integral symbol in flattened text: %q", expr, flattened)
			}
		}

		layout := typstsyntax.RenderMathToLayout(expr, ctx)
		if layout == nil {
			t.Fatalf("RenderMathToLayout(%q) returned nil", expr)
		}
		if layout.Width <= 0 {
			t.Fatalf("RenderMathToLayout(%q) returned non-positive width: %f", expr, layout.Width)
		}
		if len(layout.Elements) == 0 {
			t.Fatalf("RenderMathToLayout(%q) returned no layout elements", expr)
		}

		var stream bytes.Buffer
		typstsyntax.RenderToContentStream(&stream, layout, ctx)
		out := stream.String()
		if !strings.Contains(out, "BT\n") || !strings.Contains(out, "ET\n") {
			t.Fatalf("RenderToContentStream(%q) missing BT/ET operators", expr)
		}

		if strings.Contains(expr, "integral") && !containsAnyIntegralGlyph(out) {
			t.Fatalf("RenderToContentStream(%q) did not include integral symbol in stream", expr)
		}
	}

	if integralExprCount < 5 {
		t.Fatalf("expected at least 5 integral expressions, got %d", integralExprCount)
	}
}

func TestTypstMathStress_GenerateTemplatePDFWithIntegrals(t *testing.T) {
	t.Parallel()

	mathFontPath, ok := resolveMathFontPath()
	if !ok {
		t.Skip("no unicode math-capable font found (looked for DejaVu/Noto Math). Install fonts-dejavu-core or fonts-noto-math")
	}

	const mathFontName = "MathUnicode"
	mathEnabled := true
	expressions := []string{
		"$ integral_0^1 x^2 d x $",
		"$ integral_a^b frac(1, 1 + x^2) d x $",
		"$ integral_(0)^oo e^(-x) d x $",
		"$ integral.double_A f(x, y) d x d y $",
		"$ integral.triple_V rho(x, y, z) d V $",
		"$ integral.cont_C F dot d r $",
		"$ partial^2 u / partial x^2 = 0 $",
		"$ lim_(x -> 0) frac(sin x, x) = 1 $",
		"$ nabla dot F = 0 $",
		"$ sum_(k=1)^n k = frac(n(n+1), 2) $",
	}

	rows := make([]models.Row, 0, len(expressions))
	for _, expr := range expressions {
		rows = append(rows, models.Row{Row: []models.Cell{{
			Props:       mathFontName + ":12:000:left:1:1:1:1",
			Text:        expr,
			MathEnabled: &mathEnabled,
		}}})
	}
	rowHeights := make([]float64, len(rows))
	for i := range rowHeights {
		rowHeights[i] = 36
	}

	template := models.PDFTemplate{
		Config: models.Config{
			Page: "A4",
			CustomFonts: []models.CustomFontConfig{{
				Name:     mathFontName,
				FilePath: mathFontPath,
			}},
		},
		Title: models.Title{
			Text:  "Typst Math Stress",
			Props: mathFontName + ":14:100:center:0:0:0:0",
		},
		Table: []models.Table{{
			MaxColumns: 1,
			Rows:       rows,
			RowHeights: rowHeights,
		}},
	}

	pdfBytes, err := GenerateTemplatePDF(template)
	if err != nil {
		t.Fatalf("GenerateTemplatePDF failed: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("GenerateTemplatePDF returned empty bytes")
	}

	if !bytes.Contains(pdfBytes, []byte("/Identity-H")) {
		t.Fatal("expected PDF to use Identity-H encoding for Unicode math glyphs")
	}
	if !bytes.Contains(pdfBytes, []byte("/ToUnicode")) {
		t.Fatal("expected PDF to include ToUnicode mapping for custom math font")
	}
	if !bytes.Contains(pdfBytes, []byte("/Subtype /Type0")) || !bytes.Contains(pdfBytes, []byte("/CF")) {
		t.Fatal("expected PDF resources to include embedded Type0 custom font reference")
	}
}

func TestTypstMathStress_GenerateEquationBankPDF(t *testing.T) {
	mathFontPath, ok := resolveMathFontPath()
	if !ok {
		t.Skip("no unicode math-capable font found (looked for DejaVu/Noto Math). Install fonts-dejavu-core or fonts-noto-math")
	}

	const mathFontName = "MathUnicode"
	mathEnabled := true
	type categoryBatch struct {
		name     string
		formulas []string
	}

	batches := []categoryBatch{
		{
			name: "Calculus",
			formulas: []string{
				"$ integral_0^1 x^2 d x $",
				"$ integral_a^b frac(1, 1 + x^2) d x $",
				"$ integral_(0)^oo e^(-x) d x $",
				"$ integral.double_A f(x, y) d x d y $",
				"$ integral.triple_V rho(x, y, z) d V $",
				"$ integral.cont_C F dot d r $",
				"$ partial f / partial x $",
				"$ partial^2 u / partial x^2 $",
				"$ lim_(x -> 0) frac(sin x, x) = 1 $",
				"$ lim_(n -> oo) (1 + frac(1, n))^n = e $",
				"$ d/dx (x^3) = 3x^2 $",
				"$ nabla dot F = 0 $",
				"$ nabla times F = 0 $",
				"$ sum_(k=1)^n k = frac(n(n+1), 2) $",
				"$ product_(i=1)^n i = n! $",
			},
		},
		{
			name: "Linear Algebra",
			formulas: []string{
				"$ vec(1, 2, 3) $",
				"$ vec(a, b, c) $",
				"$ mat(1, 0; 0, 1) $",
				"$ mat(1, 2; 3, 4) $",
				"$ A x = b $",
				"$ A^T A $",
				"$ det(A) = lambda_1 lambda_2 $",
				"$ rank(A) <= n $",
				"$ ker(T) subset.eq V $",
				"$ im(T) subset.eq W $",
				"$ x dot y = 0 $",
				"$ x times y $",
				"$ norm(x) = sqrt(x_1^2 + x_2^2) $",
				"$ lr( mat(1, 0; 0, 1) ) $",
				"$ A = U Sigma V^T $",
			},
		},
		{
			name: "Logic",
			formulas: []string{
				"$ forall x in RR, exists y in RR: y >= x $",
				"$ forall p, p or not p $",
				"$ p and q => p $",
				"$ p iff q $",
				"$ p implies q $",
				"$ not (p and q) iff (not p) or (not q) $",
				"$ p tack q $",
				"$ Gamma models phi $",
				"$ top != bot $",
				"$ exists x: P(x) $",
				"$ alpha -> beta $",
				"$ A <= B and B <= C => A <= C $",
			},
		},
		{
			name: "Set Theory",
			formulas: []string{
				"$ x in A $",
				"$ x in.not A $",
				"$ A subset B $",
				"$ A subset.eq B $",
				"$ A supset.eq B $",
				"$ A union B $",
				"$ A sect B $",
				"$ A \\ B $",
				"$ emptyset subset.eq A $",
				"$ RR^3 subset.eq CC^3 $",
				"$ NN subset.eq ZZ subset.eq QQ subset.eq RR subset.eq CC $",
				"$ f: A -> B $",
				"$ f: A -> B, x mapsto f(x) $",
			},
		},
		{
			name: "Symbols and Operators",
			formulas: []string{
				"$ alpha + beta = gamma $",
				"$ theta != pi $",
				"$ lambda >= 0 $",
				"$ mu <= nu $",
				"$ xi approx zeta $",
				"$ phi equiv psi $",
				"$ aleph_0 < oo $",
				"$ hbar omega $",
				"$ angle ABC = 90 degree $",
				"$ a parallel b $",
				"$ l perp m $",
				"$ sqrt(x^2 + y^2) $",
				"$ root(3, x) $",
				"$ frac(a+b, c-d) $",
				"$ hat(x) + tilde(y) + bar(z) $",
			},
		},
	}

	totalFormulaCount := 0
	rows := make([]models.Row, 0, 120)
	for _, batch := range batches {
		for _, formula := range batch.formulas {
			totalFormulaCount++
			rows = append(rows, models.Row{Row: []models.Cell{
				{Props: mathFontName + ":10:100:left:1:1:1:1", Text: batch.name},
				{Props: mathFontName + ":11:000:left:1:1:1:1", Text: formula, MathEnabled: &mathEnabled},
			}})
		}
	}
	rowHeights := make([]float64, len(rows))
	for i := range rowHeights {
		rowHeights[i] = 2.5
	}

	if totalFormulaCount < 50 || totalFormulaCount > 100 {
		t.Fatalf("equation bank size must be between 50 and 100, got %d", totalFormulaCount)
	}

	template := models.PDFTemplate{
		Config: models.Config{
			Page: "A4",
			CustomFonts: []models.CustomFontConfig{{
				Name:     mathFontName,
				FilePath: mathFontPath,
			}},
		},
		Title: models.Title{
			Text:  "Typst Syntax Equation Bank",
			Props: mathFontName + ":16:100:center:0:0:0:0",
		},
		Table: []models.Table{{
			MaxColumns:   2,
			ColumnWidths: []float64{2, 4.8},
			Rows:         rows,
			RowHeights:   rowHeights,
		}},
	}

	pdfBytes, err := GenerateTemplatePDF(template)
	if err != nil {
		t.Fatalf("GenerateTemplatePDF for equation bank failed: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("equation bank PDF is empty")
	}

	outPath, err := filepath.Abs("../../sampledata/typstsyntax/typst_math_equation_bank.pdf")
	if err != nil {
		t.Fatalf("failed to build output path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}
	if err := os.WriteFile(outPath, pdfBytes, 0o644); err != nil {
		t.Fatalf("failed to write equation bank PDF: %v", err)
	}

	fi, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("failed to stat equation bank PDF: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("equation bank PDF on disk is empty: %s", outPath)
	}

	red, err := redact.NewRedactor(pdfBytes)
	if err != nil {
		t.Fatalf("redact.NewRedactor failed: %v", err)
	}

	info, err := red.GetPageInfo()
	if err != nil {
		t.Fatalf("GetPageInfo failed: %v", err)
	}
	if info.TotalPages < 2 {
		t.Fatalf("expected equation bank PDF to span multiple pages, got %d", info.TotalPages)
	}

	if !bytes.Contains(pdfBytes, []byte("/Identity-H")) {
		t.Fatal("expected equation bank PDF to use Identity-H encoding for Unicode math glyphs")
	}
	if !bytes.Contains(pdfBytes, []byte("/ToUnicode")) {
		t.Fatal("expected equation bank PDF to include ToUnicode mapping for custom math font")
	}
	if !bytes.Contains(pdfBytes, []byte("/Subtype /Type0")) || !bytes.Contains(pdfBytes, []byte("/CF")) {
		t.Fatal("expected equation bank PDF resources to include embedded Type0 custom font reference")
	}
}

func TestTypstMathStress_GenerateImageStyleShowcasePDF(t *testing.T) {
	mathFontPath, ok := resolveMathFontPath()
	if !ok {
		t.Skip("no unicode math-capable font found (looked for DejaVu/Noto Math). Install fonts-dejavu-core or fonts-noto-math")
	}

	const mathFontName = "MathUnicode"
	mathEnabled := true

	showcaseExpressions := []string{
		"$ dot(x)_j = frac(q, c) sum_(l=1)^n dot(x)_l [frac(partial A, partial x_j) - frac(partial A_j, partial x_l)] $",
		"$ - frac(partial A_j, partial t) - c frac(partial phi, partial x_j) $",
		"$ F_(mu,nu) = lr( mat(0, B_z, -B_y, -iE_x; -B_z, 0, B_x, -iE_y; B_y, -B_x, 0, -iE_z; iE_x, iE_y, iE_z, 0) ) $",
		"$ Sigma = sum_(l=1)^n [a_l + b_l] $",
		"$ sum_(k=0)^n {(k+1)} = [frac((n+1)(n+2), 2)] $",
		"$ [lr((ax + b)(cx + d))] = [acx^2 + (ad + bc)x + bd] $",
		"$ lr((ax + b)(cx + d)) = acx^2 + (ad + bc)x + bd $",
		"$ (ax)(d) + (b)(cx) arrow.r (ad + bc)x $",
	}

	rows := make([]models.Row, 0, len(showcaseExpressions)+3)
	rowHeights := make([]float64, 0, len(showcaseExpressions)+3)

	rows = append(rows, models.Row{Row: []models.Cell{{
		Props: mathFontName + ":13:100:left:1:1:1:1",
		Text:  "Image-Inspired Typst Math Showcase",
	}}})
	rowHeights = append(rowHeights, 8)

	for _, expr := range showcaseExpressions {
		ast := typstsyntax.ParseMath(expr)
		if ast == nil {
			t.Fatalf("ParseMath(%q) returned nil", expr)
		}
		flat := strings.TrimSpace(typstsyntax.FlattenToText(ast))
		if flat == "" {
			t.Fatalf("FlattenToText(%q) returned empty text", expr)
		}
		if strings.Contains(expr, "sum_") && !strings.Contains(flat, "∑") {
			t.Fatalf("FlattenToText(%q) did not include sigma symbol: %q", expr, flat)
		}
		if strings.Contains(expr, "[") && !strings.Contains(flat, "[") {
			t.Fatalf("FlattenToText(%q) did not include opening bracket: %q", expr, flat)
		}

		rows = append(rows, models.Row{Row: []models.Cell{{
			Props:       mathFontName + ":12:000:left:1:1:1:1",
			Text:        expr,
			MathEnabled: &mathEnabled,
		}}})

		rowHeights = append(rowHeights, 5)

	}

	template := models.PDFTemplate{
		Config: models.Config{
			Page: "A4",
			CustomFonts: []models.CustomFontConfig{{
				Name:     mathFontName,
				FilePath: mathFontPath,
			}},
		},
		Title: models.Title{
			Text:  "Typst Figure-Like Math Output",
			Props: mathFontName + ":15:100:center:0:0:0:0",
		},
		Table: []models.Table{{
			MaxColumns: 1,
			Rows:       rows,
			RowHeights: rowHeights,
		}},
	}

	pdfBytes, err := GenerateTemplatePDF(template)
	if err != nil {
		t.Fatalf("GenerateTemplatePDF for image-style showcase failed: %v", err)
	}
	if len(pdfBytes) == 0 {
		t.Fatal("image-style showcase PDF is empty")
	}

	outPath, err := filepath.Abs("../../sampledata/typstsyntax/typst_math_image_style_showcase.pdf")
	if err != nil {
		t.Fatalf("failed to build output path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}
	if err := os.WriteFile(outPath, pdfBytes, 0o644); err != nil {
		t.Fatalf("failed to write image-style showcase PDF: %v", err)
	}

	fi, err := os.Stat(outPath)
	if err != nil {
		t.Fatalf("failed to stat image-style showcase PDF: %v", err)
	}
	if fi.Size() == 0 {
		t.Fatalf("image-style showcase PDF on disk is empty: %s", outPath)
	}

	if !bytes.Contains(pdfBytes, []byte("/Identity-H")) {
		t.Fatal("expected image-style showcase PDF to use Identity-H encoding for Unicode math glyphs")
	}
	if !bytes.Contains(pdfBytes, []byte("/ToUnicode")) {
		t.Fatal("expected image-style showcase PDF to include ToUnicode mapping for custom math font")
	}
}

func containsAnyIntegralGlyph(text string) bool {
	return strings.Contains(text, "∫") ||
		strings.Contains(text, "∬") ||
		strings.Contains(text, "∭") ||
		strings.Contains(text, "∮")
}

func resolveMathFontPath() (string, bool) {
	candidates := []string{
		"/usr/share/fonts/truetype/noto/NotoSansMath-Regular.ttf",
		"/usr/share/fonts/opentype/noto/NotoSansMath-Regular.otf",
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation2/LiberationSans-Regular.ttf",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}

	return "", false
}
