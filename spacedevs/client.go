// Package spacedevs is the library behind the spacedevs command line:
// the HTTP client, request shaping, and the typed data models for The Space Devs
// Launch Library 2 API at ll.thespacedevs.com.
//
// The Client here is the spine every command shares. It sets a real
// User-Agent, paces requests so a busy session stays polite (15 req/hour
// anonymous limit), and retries transient failures (429 and 5xx).
package spacedevs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Host is the API hostname.
const Host = "ll.thespacedevs.com"

// BaseURL is the root every request is built from.
const BaseURL = "https://ll.thespacedevs.com/2.2.0"

// DefaultUserAgent identifies the client to the API.
const DefaultUserAgent = "spacedevs-cli/0.1 (tamnd87@gmail.com)"

// Config holds tunable knobs.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns safe defaults for anonymous access.
// The API rate-limits anonymous users to 15 req/hour; 4s between requests
// keeps typical usage well within that budget.
func DefaultConfig() Config {
	return Config{
		BaseURL:   BaseURL,
		Rate:      4 * time.Second,
		Timeout:   30 * time.Second,
		Retries:   3,
		UserAgent: DefaultUserAgent,
	}
}

// Client talks to The Space Devs Launch Library 2 API.
type Client struct {
	HTTP      *http.Client
	BaseURL   string
	UserAgent string
	// Rate is the minimum gap between requests. Zero means no pacing.
	Rate    time.Duration
	Retries int

	last time.Time
}

// NewClient returns a Client with sensible defaults.
func NewClient() *Client {
	cfg := DefaultConfig()
	return &Client{
		HTTP:      &http.Client{Timeout: cfg.Timeout},
		BaseURL:   cfg.BaseURL,
		UserAgent: cfg.UserAgent,
		Rate:      cfg.Rate,
		Retries:   cfg.Retries,
	}
}

// Get fetches url and returns the response body. It paces and retries
// according to the client's settings. The caller owns nothing extra; the body
// is read fully and closed here.
func (c *Client) Get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) (body []byte, retry bool, err error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

// pace blocks until at least Rate has passed since the previous request.
func (c *Client) pace() {
	if c.Rate <= 0 {
		return
	}
	if wait := c.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
