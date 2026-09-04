package handlers

import (
	"errors"
	"log"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/chinmay-sawant/gopdfsuit/v6/internal/models"
	"github.com/gin-gonic/gin"
)

// handleHTMLToPDF handles HTML to PDF conversion using htmltopdf
func handleHTMLToPDF(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxHTMLBodyBytes)

	var req models.HTMLToPDFRequest
	data, err := c.GetRawData()
	if err != nil {
		if isBodyTooLargeErr(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		log.Printf("handleHTMLToPDF: read body failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := sonic.Unmarshal(data, &req); err != nil {
		log.Printf("handleHTMLToPDF: invalid JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.HTML == "" && req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either html or url is required"})
		return
	}
	if err := validateFetchURL(req.URL); err != nil {
		if errors.Is(err, errFetchURLBlocked) {
			c.JSON(http.StatusForbidden, gin.H{"error": "url target is not allowed"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
		return
	}

	// Set defaults
	if req.PageSize == "" {
		req.PageSize = "A4"
	}
	if req.Orientation == "" {
		req.Orientation = "Portrait"
	}
	if req.MarginTop == "" {
		req.MarginTop = "10mm" //nolint:goconst
	}
	if req.MarginRight == "" {
		req.MarginRight = "10mm"
	}
	if req.MarginBottom == "" {
		req.MarginBottom = "10mm"
	}
	if req.MarginLeft == "" {
		req.MarginLeft = "10mm"
	}
	if req.DPI == 0 {
		req.DPI = 300
	}

	pdfBytes, err := pdfService.HTMLToPDF(req)
	if err != nil {
		log.Printf("handleHTMLToPDF: conversion failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF conversion failed"})
		return
	}

	c.Header("Content-Type", mimeTypePDF)
	c.Header("Content-Disposition", "attachment; filename=converted.pdf")
	c.Data(http.StatusOK, mimeTypePDF, pdfBytes)
}

// handleHTMLToImage handles HTML to image conversion using htmltoimage
func handleHTMLToImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxHTMLBodyBytes)

	var req models.HTMLToImageRequest
	data, err := c.GetRawData()
	if err != nil {
		if isBodyTooLargeErr(err) {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "request body too large"})
			return
		}
		log.Printf("handleHTMLToImage: read body failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := sonic.Unmarshal(data, &req); err != nil {
		log.Printf("handleHTMLToImage: invalid JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.HTML == "" && req.URL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "either html or url is required"})
		return
	}
	if err := validateFetchURL(req.URL); err != nil {
		if errors.Is(err, errFetchURLBlocked) {
			c.JSON(http.StatusForbidden, gin.H{"error": "url target is not allowed"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid url"})
		return
	}

	// Set defaults
	if req.Format == "" {
		req.Format = "png"
	}
	if req.Quality == 0 {
		req.Quality = 94
	}
	if req.Zoom == 0 {
		req.Zoom = 1.0
	}

	imageBytes, err := pdfService.HTMLToImage(req)
	if err != nil {
		log.Printf("handleHTMLToImage: conversion failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "image conversion failed"})
		return
	}

	contentType := "image/png"
	switch req.Format {
	case "jpg", "jpeg":
		contentType = "image/jpeg"
	case "svg":
		contentType = "image/svg+xml"
	}

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "attachment; filename=converted."+req.Format)
	c.Data(http.StatusOK, contentType, imageBytes)
}
