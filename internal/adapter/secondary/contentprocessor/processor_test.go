package contentprocessor

import (
	"context"
	"strings"
	"testing"
)

func TestProcessWithReader(t *testing.T) {
	processor := NewProcessor()

	testCases := []struct {
		name        string
		htmlContent string
		url         string
		wantTitle   string
		wantURL     bool
		wantContent string
	}{
		{
			name: "basic HTML with title and URL",
			htmlContent: `
<!DOCTYPE html>
<html>
<head><title>Test Article</title></head>
<body>
	<article>
		<h1>Test Article</h1>
		<p>This is a test paragraph with <strong>bold text</strong>.</p>
		<ul>
			<li>Item 1</li>
			<li>Item 2</li>
		</ul>
	</article>
</body>
</html>`,
			url:         "https://example.com/test",
			wantTitle:   "# Test Article",
			wantURL:     true,
			wantContent: "bold text",
		},
		{
			name: "HTML without URL",
			htmlContent: `
<!DOCTYPE html>
<html>
<head><title>Another Test</title></head>
<body>
	<article>
		<h1>Another Test</h1>
		<p>Simple content here.</p>
	</article>
</body>
</html>`,
			url:         "",
			wantTitle:   "# Another Test",
			wantURL:     false,
			wantContent: "Simple content",
		},
		{
			name: "HTML with complex formatting",
			htmlContent: `
<!DOCTYPE html>
<html>
<head><title>Complex Article</title></head>
<body>
	<article>
		<h1>Complex Article</h1>
		<h2>Section 1</h2>
		<p>Text with <a href="https://example.com">a link</a>.</p>
		<h2>Section 2</h2>
		<p>More text with <code>code snippet</code>.</p>
	</article>
</body>
</html>`,
			url:         "https://test.com",
			wantTitle:   "# Complex Article",
			wantURL:     true,
			wantContent: "Section 1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := processor.ProcessWithReader(ctx, []byte(tc.htmlContent), tc.url)
			if err != nil {
				t.Fatalf("ProcessWithReader failed: %v", err)
			}

			// Check title is present
			if !strings.Contains(result, tc.wantTitle) {
				t.Errorf("Expected title '%s' not found in result:\n%s", tc.wantTitle, result)
			}

			// Check URL attribution
			if tc.wantURL {
				expectedURL := "(extracted from **" + tc.url + "**)"
				if !strings.Contains(result, expectedURL) {
					t.Errorf("Expected URL attribution '%s' not found in result:\n%s", expectedURL, result)
				}
			} else {
				if strings.Contains(result, "(extracted from") {
					t.Errorf("Unexpected URL attribution found in result when none expected:\n%s", result)
				}
			}

			// Check content is present
			if !strings.Contains(result, tc.wantContent) {
				t.Errorf("Expected content '%s' not found in result:\n%s", tc.wantContent, result)
			}

			// Verify it's markdown (starts with #)
			if !strings.HasPrefix(strings.TrimSpace(result), "#") {
				t.Errorf("Result does not start with markdown heading:\n%s", result)
			}
		})
	}
}

func TestProcessWithReaderInvalidHTML(t *testing.T) {
	processor := NewProcessor()
	ctx := context.Background()

	invalidHTML := `<html><body><p>No proper structure</body></html>`

	// This should still work with readability - it's quite forgiving
	result, err := processor.ProcessWithReader(ctx, []byte(invalidHTML), "")
	if err != nil {
		t.Fatalf("ProcessWithReader failed on simple HTML: %v", err)
	}

	if result == "" {
		t.Error("Expected some result even from minimal HTML")
	}
}

func TestProcessWithReaderEmptyContent(t *testing.T) {
	processor := NewProcessor()
	ctx := context.Background()

	// Empty content - readability is forgiving and may still return a result
	result, err := processor.ProcessWithReader(ctx, []byte(""), "")
	// Either an error or an empty/minimal result is acceptable
	if err == nil && result == "" {
		t.Error("Expected either an error or some result for empty content")
	}
}

func TestNewProcessor(t *testing.T) {
	processor := NewProcessor()
	if processor == nil {
		t.Fatal("NewProcessor returned nil")
	}
}

func TestProcessWithReaderMatchesTypeScriptBehavior(t *testing.T) {
	// This test verifies the output format matches the TypeScript ReaderService
	processor := NewProcessor()
	ctx := context.Background()

	htmlContent := `
<!DOCTYPE html>
<html>
<head><title>Example Article</title></head>
<body>
	<article>
		<h1>Example Article</h1>
		<p>First paragraph.</p>
		<p>Second paragraph.</p>
	</article>
</body>
</html>`

	url := "https://example.com/article"

	result, err := processor.ProcessWithReader(ctx, []byte(htmlContent), url)
	if err != nil {
		t.Fatalf("ProcessWithReader failed: %v", err)
	}

	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Fatalf("Expected at least 3 lines (title, URL, blank, content), got %d", len(lines))
	}

	// Line 1: Should be "# Title"
	if !strings.HasPrefix(lines[0], "# ") {
		t.Errorf("First line should be title with '# ', got: %s", lines[0])
	}

	// Line 2: Should be URL attribution
	if !strings.Contains(lines[1], "(extracted from **") {
		t.Errorf("Second line should be URL attribution, got: %s", lines[1])
	}

	// Line 3: Should be blank
	if strings.TrimSpace(lines[2]) != "" {
		t.Errorf("Third line should be blank, got: %s", lines[2])
	}
}
