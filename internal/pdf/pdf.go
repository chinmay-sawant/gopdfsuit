package pdf

// The original `pdf.go` was large. It has been split into smaller files by responsibility:
// - types.go        (page size and dimensions)
// - utils.go        (parsing helpers and string escaping)
// - pagemanager.go  (PageManager and page lifecycle)
// - draw.go         (drawing helpers: title, table, footer, watermark)
// - generator.go    (GenerateTemplatePDF and orchestration)
// - xfdf.go         (XFDF parsing and PDF form filling)

// This file intentionally left minimal to keep package build roots simple.

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/chinmay-sawant/gopdfsuit/internal/models"
)

// ConvertHTMLToPDF converts HTML content to PDF using wkhtmltopdf
func ConvertHTMLToPDF(req models.WKHTMLToPDFRequest) ([]byte, error) {
	// Check if wkhtmltopdf is available
	if _, err := exec.LookPath("wkhtmltopdf"); err != nil {
		return nil, fmt.Errorf("wkhtmltopdf is not installed or not in PATH: %v", err)
	}

	// Create temporary HTML file if HTML content is provided
	var tempHTML string
	if req.HTML != "" {
		tempFile, err := os.CreateTemp("", "wkhtml_*.html")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp HTML file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		if _, err := tempFile.WriteString(req.HTML); err != nil {
			return nil, fmt.Errorf("failed to write HTML content: %v", err)
		}
		tempHTML = tempFile.Name()
	}

	// Build wkhtmltopdf command arguments
	args := []string{}

	// Page settings
	args = append(args, "--page-size", req.PageSize)
	args = append(args, "--orientation", req.Orientation)
	args = append(args, "--margin-top", req.MarginTop)
	args = append(args, "--margin-right", req.MarginRight)
	args = append(args, "--margin-bottom", req.MarginBottom)
	args = append(args, "--margin-left", req.MarginLeft)
	args = append(args, "--dpi", fmt.Sprintf("%d", req.DPI))

	// Quality settings
	if req.Grayscale {
		args = append(args, "--grayscale")
	}
	if req.LowQuality {
		args = append(args, "--lowquality")
	}

	// Network and loading options for better reliability
	args = append(args, "--load-error-handling", "ignore")
	args = append(args, "--load-media-error-handling", "ignore")
	args = append(args, "--disable-external-links")
	args = append(args, "--disable-internal-links")
	args = append(args, "--disable-javascript")
	args = append(args, "--no-background")
	args = append(args, "--encoding", "utf-8")
	args = append(args, "--minimum-font-size", "8")

	// Additional options
	for key, value := range req.Options {
		args = append(args, fmt.Sprintf("--%s", key))
		if value != "" {
			args = append(args, value)
		}
	}

	// Input source
	if tempHTML != "" {
		args = append(args, tempHTML)
	} else if req.URL != "" {
		args = append(args, req.URL)
	} else {
		return nil, fmt.Errorf("either HTML content or URL must be provided")
	}

	// Output to stdout
	args = append(args, "-")

	// Execute wkhtmltopdf
	cmd := exec.Command("wkhtmltopdf", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if it's a network-related error and try with more restrictive options
		if strings.Contains(stderr.String(), "HostNotFoundError") ||
			strings.Contains(stderr.String(), "network status code 3") {
			// Retry with even more restrictive network options
			retryArgs := append(args, "--disable-plugins", "--disable-local-file-access")
			cmd = exec.Command("wkhtmltopdf", retryArgs...)
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			if retryErr := cmd.Run(); retryErr != nil {
				return nil, fmt.Errorf("wkhtmltopdf execution failed after retry: %v, stderr: %s", retryErr, stderr.String())
			}
		} else {
			return nil, fmt.Errorf("wkhtmltopdf execution failed: %v, stderr: %s", err, stderr.String())
		}
	}

	return stdout.Bytes(), nil
}

// ConvertHTMLToImage converts HTML content to image using wkhtmltoimage
func ConvertHTMLToImage(req models.WKHTMLToImageRequest) ([]byte, error) {
	// Check if wkhtmltoimage is available
	if _, err := exec.LookPath("wkhtmltoimage"); err != nil {
		return nil, fmt.Errorf("wkhtmltoimage is not installed or not in PATH: %v", err)
	}

	// Create temporary HTML file if HTML content is provided
	var tempHTML string
	if req.HTML != "" {
		tempFile, err := os.CreateTemp("", "wkhtml_*.html")
		if err != nil {
			return nil, fmt.Errorf("failed to create temp HTML file: %v", err)
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		if _, err := tempFile.WriteString(req.HTML); err != nil {
			return nil, fmt.Errorf("failed to write HTML content: %v", err)
		}
		tempHTML = tempFile.Name()
	}

	// Build wkhtmltoimage command arguments
	args := []string{}

	// Format
	args = append(args, "--format", req.Format)

	// Quality
	args = append(args, "--quality", fmt.Sprintf("%d", req.Quality))

	// Zoom
	if req.Zoom > 0 {
		args = append(args, "--zoom", fmt.Sprintf("%.2f", req.Zoom))
	}

	// Dimensions
	if req.Width > 0 {
		args = append(args, "--width", fmt.Sprintf("%d", req.Width))
	}
	if req.Height > 0 {
		args = append(args, "--height", fmt.Sprintf("%d", req.Height))
	}

	// Crop settings
	if req.CropWidth > 0 {
		args = append(args, "--crop-w", fmt.Sprintf("%d", req.CropWidth))
	}
	if req.CropHeight > 0 {
		args = append(args, "--crop-h", fmt.Sprintf("%d", req.CropHeight))
	}
	if req.CropX > 0 {
		args = append(args, "--crop-x", fmt.Sprintf("%d", req.CropX))
	}
	if req.CropY > 0 {
		args = append(args, "--crop-y", fmt.Sprintf("%d", req.CropY))
	}

	// Network and loading options for better reliability
	args = append(args, "--load-error-handling", "ignore")
	args = append(args, "--load-media-error-handling", "ignore")
	args = append(args, "--disable-plugins")
	args = append(args, "--disable-local-file-access")
	args = append(args, "--disable-javascript")
	args = append(args, "--encoding", "utf-8")

	// Additional options
	for key, value := range req.Options {
		args = append(args, fmt.Sprintf("--%s", key))
		if value != "" {
			args = append(args, value)
		}
	}

	// Input source
	if tempHTML != "" {
		args = append(args, tempHTML)
	} else if req.URL != "" {
		args = append(args, req.URL)
	} else {
		return nil, fmt.Errorf("either HTML content or URL must be provided")
	}

	// Output to stdout
	args = append(args, "-")

	// Execute wkhtmltoimage
	cmd := exec.Command("wkhtmltoimage", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		// Check if it's a network-related error
		if strings.Contains(stderr.String(), "HostNotFoundError") ||
			strings.Contains(stderr.String(), "network status code 3") {
			return nil, fmt.Errorf("wkhtmltoimage failed due to network connectivity issues. Please ensure all images use local data URIs or check your internet connection: %v", err)
		}
		return nil, fmt.Errorf("wkhtmltoimage execution failed: %v, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}
