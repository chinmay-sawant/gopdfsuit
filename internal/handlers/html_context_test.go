package handlers

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/gopdfsuit/v6/internal/handlers/mocks"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"go.uber.org/mock/gomock"
)

type contextHTMLService struct {
	*mocks.MockPDFService
	pdfContext   context.Context
	imageContext context.Context
}

func (s *contextHTMLService) HTMLToPDFContext(ctx context.Context, _ models.HTMLToPDFRequest) ([]byte, error) {
	s.pdfContext = ctx
	return []byte("pdf"), nil
}

func (s *contextHTMLService) HTMLToImageContext(ctx context.Context, _ models.HTMLToImageRequest) ([]byte, error) {
	s.imageContext = ctx
	return []byte("image"), nil
}

func TestHTMLServiceReceivesCallerContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	service := &contextHTMLService{MockPDFService: mocks.NewMockPDFService(ctrl)}
	previous := pdfService
	pdfService = service
	t.Cleanup(func() { pdfService = previous })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if _, err := htmlToPDF(ctx, models.HTMLToPDFRequest{HTML: "<p>pdf</p>"}); err != nil {
		t.Fatalf("htmlToPDF: %v", err)
	}
	if _, err := htmlToImage(ctx, models.HTMLToImageRequest{HTML: "<p>image</p>"}); err != nil {
		t.Fatalf("htmlToImage: %v", err)
	}
	if service.pdfContext != ctx || service.imageContext != ctx {
		t.Fatal("HTML service did not receive the caller context")
	}
}
