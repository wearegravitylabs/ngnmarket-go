package api

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Option configures the client. Pass options to New.
type Option func(*Call) error

// WithBaseURL overrides the API base URL (default DefaultBaseURL). The US
// endpoints are resolved relative to it ("<base>/us/..."). Useful for tests
// and for pointing at a proxy.
func WithBaseURL(raw string) Option {
	return func(c *Call) error {
		u, err := url.Parse(strings.TrimRight(raw, "/"))
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return errors.New("ngnmarket: WithBaseURL needs an absolute http(s) URL")
		}
		c.baseURL = u.String()
		return nil
	}
}

// WithHTTPClient sets the HTTP client used for requests. Use it to supply a
// custom transport, proxy or timeout. The default has a 15 second timeout.
func WithHTTPClient(h *http.Client) Option {
	return func(c *Call) error {
		if h == nil {
			return errors.New("ngnmarket: WithHTTPClient got a nil client")
		}
		c.httpClient = h
		return nil
	}
}

// WithUserAgent prepends a product token to the User-Agent header, e.g.
// "silo/1.4". The SDK's own token is always kept.
func WithUserAgent(product string) Option {
	return func(c *Call) error {
		if p := strings.TrimSpace(product); p != "" {
			c.userAgent = p + " " + c.userAgent
		}
		return nil
	}
}

// WithMaxRetries sets how many times a failed request is retried (default 2).
// Only safe, transient failures are retried: network errors, 5xx responses and
// RATE_LIMITED. QUOTA_EXCEEDED, auth and plan errors are never retried. Use 0
// to disable retries.
func WithMaxRetries(n int) Option {
	return func(c *Call) error {
		if n < 0 {
			return errors.New("ngnmarket: WithMaxRetries must be >= 0")
		}
		c.maxRetries = n
		return nil
	}
}

// WithMaxRetryWait caps how long the SDK will sleep before a retry (default
// 5s). If the API asks for a longer wait (Retry-After on a RATE_LIMITED error
// is typically up to 60s) the error is returned immediately with
// APIError.RetryAfter set, instead of blocking the caller. Callers that run in
// a background job can raise this.
func WithMaxRetryWait(d time.Duration) Option {
	return func(c *Call) error {
		if d < 0 {
			return errors.New("ngnmarket: WithMaxRetryWait must be >= 0")
		}
		c.maxRetryWait = d
		return nil
	}
}

// WithRateLimit enables a client-side limiter of requestsPerMinute, so the SDK
// paces itself below the plan's per-minute cap (Free 30, Hobby 60, Starter and
// Pro 120, Business 300) instead of relying on 429 responses. burst is the
// number of requests allowed back to back; pass 0 for a sensible default.
func WithRateLimit(requestsPerMinute, burst int) Option {
	return func(c *Call) error {
		if requestsPerMinute <= 0 {
			return errors.New("ngnmarket: WithRateLimit needs requestsPerMinute > 0")
		}
		if burst <= 0 {
			burst = requestsPerMinute / 6
			if burst < 1 {
				burst = 1
			}
		}
		c.limiter = newLimiter(requestsPerMinute, burst)
		return nil
	}
}
