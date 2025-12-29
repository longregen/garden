package httpfetch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"garden3/internal/port/output"
)

// Fetcher implements the output.HTTPFetcher interface
type Fetcher struct {
	client *http.Client
}

// NewFetcher creates a new HTTP fetcher
func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

func (f *Fetcher) Fetch(ctx context.Context, url string, timeoutMs int) (*output.FetchResponse, error) {
	timeout := time.Duration(timeoutMs) * time.Millisecond

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/136.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := f.client.Do(req)
	if err != nil {
		// Try HTTP if HTTPS fails with certificate issues
		if strings.HasPrefix(url, "https://") && isCertificateError(err) {
			httpURL := strings.Replace(url, "https://", "http://", 1)
			req, err = http.NewRequestWithContext(ctx, "GET", httpURL, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create fallback request: %w", err)
			}

			req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:120.0) Gecko/20100101 Firefox/136.0")
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
			req.Header.Set("Accept-Language", "en-US,en;q=0.5")

			resp, err = f.client.Do(req)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch (http fallback): %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to fetch: %w", err)
		}
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	headersJSON, err := json.Marshal(headers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal headers: %w", err)
	}

	return &output.FetchResponse{
		StatusCode: int32(resp.StatusCode),
		Headers:    string(headersJSON),
		Content:    content,
	}, nil
}

func isCertificateError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "certificate") ||
	       strings.Contains(errStr, "x509") ||
	       strings.Contains(errStr, "tls")
}
