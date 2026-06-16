package tests

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/handlers"
	"github.com/gin-gonic/gin"
)

func setupBenchmarkRouter(b *testing.B) *gin.Engine {
	b.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handlers.RegisterRoutes(router)
	return router
}

func loadFinancialReportJSON(b *testing.B) []byte {
	b.Helper()
	path := samplePath(b, "financialreport", "financial_report.json")
	data, err := os.ReadFile(path)
	if err != nil {
		b.Fatalf("read financial_report.json: %v", err)
	}
	return data
}

// BenchmarkGenerateTemplatePDF_FinancialReport measures the Gin handler path for
// POST /api/v1/generate/template-pdf using sampledata/financialreport/financial_report.json.
func BenchmarkGenerateTemplatePDF_FinancialReport(b *testing.B) {
	router := setupBenchmarkRouter(b)
	jsonData := loadFinancialReportJSON(b)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/generate/template-pdf", bytes.NewReader(jsonData))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			b.Fatalf("iteration %d: status=%d body=%s", i, w.Code, w.Body.String())
		}
		b.SetBytes(int64(w.Body.Len()))
	}
}

// BenchmarkGenerateTemplatePDF_FinancialReport_Parallel runs concurrent handler requests.
func BenchmarkGenerateTemplatePDF_FinancialReport_Parallel(b *testing.B) {
	router := setupBenchmarkRouter(b)
	jsonData := loadFinancialReportJSON(b)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/generate/template-pdf", bytes.NewReader(jsonData))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != http.StatusOK {
				b.Fatalf("status=%d", w.Code)
			}
		}
	})
}
