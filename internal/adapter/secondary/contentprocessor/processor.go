package contentprocessor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown/v2"
	readability "github.com/go-shiori/go-readability"
)

// Processor implements the output.ContentProcessor interface
type Processor struct{}

// NewProcessor creates a new content processor
func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) ProcessWithLynx(ctx context.Context, htmlContent string) (string, error) {
	tempFile, err := os.CreateTemp("", "bookmark-*.html")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(htmlContent); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	tempFile.Close()

	cmd := exec.CommandContext(ctx, "lynx", "-useragent=Mozilla/5.0", "-dump", tempFile.Name())
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to execute lynx: %w", err)
	}

	return string(output), nil
}

func (p *Processor) ProcessWithReader(ctx context.Context, htmlContent []byte, sourceURL string) (string, error) {
	// Parse the URL for readability
	parsedURL, err := url.Parse(sourceURL)
	if err != nil || sourceURL == "" {
		// Use a dummy URL if none provided or invalid
		parsedURL, _ = url.Parse("https://example.com")
	}

	// Parse the HTML with readability
	article, err := readability.FromReader(bytes.NewReader(htmlContent), parsedURL)
	if err != nil {
		return "", fmt.Errorf("could not parse content: Readability failed")
	}

	// Convert HTML content to Markdown using default converter
	markdown, err := md.ConvertString(article.Content)
	if err != nil {
		return "", fmt.Errorf("failed to convert HTML to markdown: %w", err)
	}

	// Build output similar to TypeScript version
	var result strings.Builder

	// Add title
	result.WriteString("# ")
	result.WriteString(article.Title)
	result.WriteString("\n")

	// Add URL attribution if provided
	if sourceURL != "" {
		result.WriteString("(extracted from **")
		result.WriteString(sourceURL)
		result.WriteString("**)")
		result.WriteString("\n")
	}

	result.WriteString("\n")
	result.WriteString(markdown)

	return result.String(), nil
}

func (p *Processor) ProcessURL(ctx context.Context, urlStr string) (string, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Fetch content
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	htmlContent, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Process the content
	return p.ProcessWithReader(ctx, htmlContent, urlStr)
}
