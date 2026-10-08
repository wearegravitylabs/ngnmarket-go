package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/wearegravitylabs/ngnmarket-go/model"
)

const maxResponseBytes = 16 << 20 // the largest endpoint (US identifiers) is ~1 MiB

// envelope is the response wrapper shared by every endpoint.
type envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Meta    *model.Meta     `json:"meta"`
	Error   *struct {
		Code         string `json:"code"`
		Message      string `json:"message"`
		RequiredPlan string `json:"required_plan"`
		CurrentPlan  string `json:"current_plan"`
	} `json:"error"`
}

// makeRequest is the one function every endpoint goes through. It performs a
// GET, retrying transient failures, and decodes the "data" field of the response
// into out. It returns the response's quota Meta.
func (c *Call) makeRequest(ctx context.Context, path string, query url.Values, out any) (model.Meta, error) {
	target := c.baseURL + path
	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	for attempt := 0; ; attempt++ {
		if c.limiter != nil {
			if err := c.limiter.wait(ctx); err != nil {
				return model.Meta{}, err
			}
		}

		meta, err := c.do(ctx, target, out)
		if err == nil {
			return meta, nil
		}

		wait, retry := c.shouldRetry(err, attempt)
		if !retry || ctx.Err() != nil {
			return model.Meta{}, err
		}
		if err := c.sleep(ctx, wait); err != nil {
			return model.Meta{}, err
		}
	}
}

// shouldRetry decides whether err warrants another attempt and how long to wait.
func (c *Call) shouldRetry(err error, attempt int) (time.Duration, bool) {
	if attempt >= c.maxRetries {
		return 0, false
	}
	var apiErr *model.APIError
	switch {
	case errors.As(err, &apiErr):
		if !apiErr.Retryable() {
			return 0, false
		}
		wait := apiErr.RetryAfter
		if wait <= 0 {
			wait = backoff(attempt)
		}
		if wait > c.maxRetryWait {
			// The API wants a longer pause than we are willing to block for.
			// Hand the error (with RetryAfter populated) back to the caller.
			return 0, false
		}
		return wait, true
	case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return 0, false
	default: // network-level failure
		return backoff(attempt), true
	}
}

func (c *Call) do(ctx context.Context, target string, out any) (model.Meta, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return model.Meta{}, fmt.Errorf("ngnmarket: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Never leak the key: the url.Error from net/http contains only the URL.
		return model.Meta{}, fmt.Errorf("ngnmarket: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return model.Meta{}, fmt.Errorf("ngnmarket: read response: %w", err)
	}

	var env envelope
	jsonErr := json.Unmarshal(body, &env)

	if resp.StatusCode < 200 || resp.StatusCode > 299 || (jsonErr == nil && !env.Success) {
		return model.Meta{}, newAPIError(resp, body, &env, jsonErr)
	}
	if jsonErr != nil {
		return model.Meta{}, fmt.Errorf("ngnmarket: decode response: %w", jsonErr)
	}
	if out != nil && len(env.Data) > 0 && string(env.Data) != "null" {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return model.Meta{}, fmt.Errorf("ngnmarket: decode data: %w", err)
		}
	}
	if env.Meta == nil {
		return model.Meta{}, nil
	}
	return *env.Meta, nil
}

func newAPIError(resp *http.Response, body []byte, env *envelope, jsonErr error) *model.APIError {
	e := &model.APIError{StatusCode: resp.StatusCode}
	if jsonErr == nil && env.Error != nil {
		e.Code = model.ErrorCode(env.Error.Code)
		e.Message = env.Error.Message
		e.RequiredPlan = env.Error.RequiredPlan
		e.CurrentPlan = env.Error.CurrentPlan
		e.Meta = env.Meta
	} else {
		// Not an API envelope (e.g. a gateway error page): keep a short excerpt.
		e.Message = strings.TrimSpace(string(body))
		if len(e.Message) > 200 {
			e.Message = e.Message[:200] + "…"
		}
		if e.Message == "" {
			e.Message = http.StatusText(resp.StatusCode)
		}
	}
	if secs, err := strconv.Atoi(strings.TrimSpace(resp.Header.Get("Retry-After"))); err == nil && secs > 0 {
		e.RetryAfter = time.Duration(secs) * time.Second
	} else if e.Meta != nil && e.Meta.RetryAfter > 0 {
		e.RetryAfter = time.Duration(e.Meta.RetryAfter) * time.Second
	}
	return e
}

// backoff returns an exponential delay with jitter: ~0.5s, 1s, 2s, ... capped at 8s.
func backoff(attempt int) time.Duration {
	d := 500 * time.Millisecond * time.Duration(math.Pow(2, float64(attempt)))
	if d > 8*time.Second {
		d = 8 * time.Second
	}
	return d + time.Duration(rand.Int63n(int64(d)/4+1))
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// requireSymbol trims a symbol and rejects an empty one.
func requireSymbol(symbol string) (string, error) {
	symbol = strings.TrimSpace(symbol)
	if symbol == "" {
		return "", errors.New("ngnmarket: symbol is required")
	}
	return symbol, nil
}
