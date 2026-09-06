package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/chinmay-sawant/gopdfsuit/v7/internal/models"
	"github.com/gin-gonic/gin"
)

// rejectHTMLSource enforces the shared HTML-source policy: either html or
// url is required, and fetch URLs pass the SSRF guard. It returns false
// after writing the rejection when the handler must abort.
func rejectHTMLSource(c *gin.Context, html, url string) bool {
	if html == "" && url == "" {
		abortError(c, http.StatusBadRequest, "either html or url is required")
		return false
	}
	if err := validateFetchURL(c.Request.Context(), url); err != nil {
		if errors.Is(err, errFetchURLBlocked) {
			abortError(c, http.StatusForbidden, "url target is not allowed")
			return false
		}
		abortError(c, http.StatusBadRequest, "invalid url")
		return false
	}
	return true
}

// handleHTMLToPDF handles HTML to PDF conversion using htmltopdf
func handleHTMLToPDF(c *gin.Context) {
	req, tooLarge, err := decodeJSONBody[models.HTMLToPDFRequest](c, maxHTMLBodyBytes)
	if tooLarge {
		abortError(c, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	if err != nil {
		abortError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if !rejectHTMLSource(c, req.HTML, req.URL) {
		return
	}

	req = newHTMLToPDFRequest(req)

	pdfBytes, err := htmlToPDF(c.Request.Context(), req)
	if err != nil {
		log.Printf("handleHTMLToPDF: conversion failed: %v", err)
		abortError(c, http.StatusInternalServerError, "PDF conversion failed")
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=converted.pdf")
	c.Data(http.StatusOK, mimeTypePDF, pdfBytes)
}

// handleHTMLToImage handles HTML to image conversion using htmltoimage
func handleHTMLToImage(c *gin.Context) {
	req, tooLarge, err := decodeJSONBody[models.HTMLToImageRequest](c, maxHTMLBodyBytes)
	if tooLarge {
		abortError(c, http.StatusRequestEntityTooLarge, "request body too large")
		return
	}
	if err != nil {
		abortError(c, http.StatusBadRequest, "invalid request")
		return
	}

	if !rejectHTMLSource(c, req.HTML, req.URL) {
		return
	}

	req = newHTMLToImageRequest(req)

	// gopdflib.ConvertHTMLToImage owns the format policy: svg has no image
	// equivalent and is rejected there, so the handler mirrors the
	// rejection up front to keep the 400 on this side of the service seam.
	if req.Format == "svg" {
		abortError(c, http.StatusBadRequest, "format svg is not supported: use png or jpg")
		return
	}

	imageBytes, err := htmlToImage(c.Request.Context(), req)
	if err != nil {
		log.Printf("handleHTMLToImage: conversion failed: %v", err)
		abortError(c, http.StatusInternalServerError, "image conversion failed")
		return
	}

	contentType := "image/png"
	switch req.Format {
	case "jpg", "jpeg": //nolint:goconst // jpeg is an accepted input alias
		contentType = "image/jpeg"
	case "svg":
		contentType = "image/svg+xml"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=converted."+req.Format)
	c.Data(http.StatusOK, contentType, imageBytes)
}

func htmlToPDF(ctx context.Context, req models.HTMLToPDFRequest) ([]byte, error) {
	if service, ok := pdfService.(ContextHTMLService); ok {
		return service.HTMLToPDFContext(ctx, req)
	}
	return pdfService.HTMLToPDF(req)
}

func htmlToImage(ctx context.Context, req models.HTMLToImageRequest) ([]byte, error) {
	if service, ok := pdfService.(ContextHTMLService); ok {
		return service.HTMLToImageContext(ctx, req)
	}
	return pdfService.HTMLToImage(req)
}
