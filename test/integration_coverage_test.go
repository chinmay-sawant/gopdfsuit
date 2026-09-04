package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/chinmay-sawant/gopdfsuit/v6/pkg/gopdflib"
)

// coverageDummyPDF seeds a small in-memory PDF for coverage tests.
func (s *IntegrationSuite) coverageDummyPDF(title, text string) []byte {
	s.T().Helper()
	src, err := gopdflib.GeneratePDF(gopdflib.PDFTemplate{
		Config: gopdflib.Config{Page: "A4", PageAlignment: 1},
		Title:  gopdflib.Title{Props: "Helvetica:12:000:left:0:0:0:0", Text: title + " " + text},
	})
	s.NoError(err)
	s.True(bytes.HasPrefix(src, []byte("%PDF-")), "dummy seed is not a PDF")
	return src
}

// TestCoverageGeneratePDF covers Go library PDF generation.
func (s *IntegrationSuite) TestCoverageGeneratePDF() {
	src := s.coverageDummyPDF("Coverage Dummy", "seeded paragraph for generate")
	s.Contains(string(src[:8]), "%PDF-")
	s.Contains(string(src), "%%EOF")
}

// TestCoverageReadPDF covers the read-only APIs (page info / text positions / search).
func (s *IntegrationSuite) TestCoverageReadPDF() {
	src := s.coverageDummyPDF("Coverage Read", "seeded paragraph for read")
	info, err := gopdflib.GetPageInfo(src)
	s.NoError(err)
	s.GreaterOrEqual(info.TotalPages, 1)
	s.Len(info.Pages, info.TotalPages)

	positions, err := gopdflib.ExtractTextPositions(src, 1)
	s.NoError(err)
	_ = positions

	rects, err := gopdflib.FindTextOccurrences(src, "seeded")
	s.NoError(err)
	_ = rects

	caps, err := gopdflib.AnalyzePageCapabilities(src)
	s.NoError(err)
	s.NotEmpty(caps)
}

// TestCoverageCompressLibrary covers gopdflib.CompressPDF incl. non-PDF rejection.
func (s *IntegrationSuite) TestCoverageCompressLibrary() {
	src := s.coverageDummyPDF("Coverage Compress", "seeded paragraph for compress")
	out, err := gopdflib.CompressPDF(src, gopdflib.CompressOptions{Level: gopdflib.CompressMedium})
	s.NoError(err)
	s.True(bytes.HasPrefix(out, []byte("%PDF-")))

	_, err = gopdflib.CompressPDF([]byte("not-a-pdf"), gopdflib.CompressOptions{})
	s.Error(err)
}

// TestCoverageCompressHTTP covers POST /api/v1/compress over the test server.
func (s *IntegrationSuite) TestCoverageCompressHTTP() {
	src := s.coverageDummyPDF("Coverage CompressHTTP", "seeded paragraph for http compress")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("pdf", "dummy.pdf")
	s.NoError(err)
	_, err = part.Write(src)
	s.NoError(err)
	s.NoError(writer.WriteField("level", "medium"))
	s.NoError(writer.Close())

	resp, err := s.client.Post(s.ts.URL+"/api/v1/compress", writer.FormDataContentType(), body)
	s.NoError(err)
	defer func() {
		_ = resp.Body.Close()
	}()
	s.Equal(http.StatusOK, resp.StatusCode)
	s.Equal("application/pdf", resp.Header.Get("Content-Type"))
	out, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.True(bytes.HasPrefix(out, []byte("%PDF-")))
}

// TestCoverageMergeSplitFillLibrary covers merge / split / parse / fill at lib level.
func (s *IntegrationSuite) TestCoverageMergeSplitFillLibrary() {
	a := s.coverageDummyPDF("Coverage A", "seeded doc a")
	b := s.coverageDummyPDF("Coverage B", "seeded doc b")

	merged, err := gopdflib.MergePDFs([][]byte{a, b})
	s.NoError(err)
	s.True(bytes.HasPrefix(merged, []byte("%PDF-")))

	parts, err := gopdflib.SplitPDF(merged, gopdflib.SplitSpec{Pages: []int{1}})
	s.NoError(err)
	s.Len(parts, 1)
	s.True(bytes.HasPrefix(parts[0], []byte("%PDF-")))

	pages, err := gopdflib.ParsePageSpec("1-3,5", 10)
	s.NoError(err)
	s.Equal([]int{1, 2, 3, 5}, pages)

	_, err = gopdflib.FillPDFWithXFDF(a, []byte{})
	s.Error(err)
}

// TestCoverageComplianceSmoke asserts PDF/A-4 + PDF/UA-2 XMP markers without veraPDF.
func (s *IntegrationSuite) TestCoverageComplianceSmoke() {
	src := s.coverageDummyPDF("Coverage Compliance", "seeded paragraph for compliance")
	s.Contains(string(src), "pdfaid")
	s.Contains(string(src), "pdfuaid")
}
