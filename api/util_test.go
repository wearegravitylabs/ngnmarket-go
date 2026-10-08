package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wearegravitylabs/ngnmarket-go/model"
)

// newTestClient returns a client pointed at srv with instant retries.
func newTestClient(t *testing.T, srv *httptest.Server, opts ...Option) *Call {
	t.Helper()
	opts = append([]Option{WithBaseURL(srv.URL), WithHTTPClient(srv.Client())}, opts...)
	c, err := newCall("ngm_live_test", opts...)
	if err != nil {
		t.Fatal(err)
	}
	c.sleep = func(context.Context, time.Duration) error { return nil }
	return c
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

func TestNew_Validation(t *testing.T) {
	if _, err := New("   "); err == nil {
		t.Error("empty key: want error")
	}
	if _, err := New("k", WithBaseURL("not a url")); err == nil {
		t.Error("bad base URL: want error")
	}
	if _, err := New("k", WithMaxRetries(-1)); err == nil {
		t.Error("negative retries: want error")
	}
	if _, err := New("k", WithRateLimit(0, 0)); err == nil {
		t.Error("zero rate limit: want error")
	}
	c, err := newCall("  ngm_live_x \n")
	if err != nil || c.apiKey != "ngm_live_x" {
		t.Errorf("key should be trimmed, got %q (err %v)", c.apiKey, err)
	}
}

func TestRequest_HeadersAndMeta(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer ngm_live_test" {
			t.Errorf("Authorization = %q", got)
		}
		if ua := r.Header.Get("User-Agent"); !strings.HasPrefix(ua, "silo/1 ngnmarket-go/") {
			t.Errorf("User-Agent = %q", ua)
		}
		writeJSON(w, 200, `{"success":true,"data":{"data":[],"count":0},"meta":{"plan":"free","calls_used":4,"calls_remaining":2996,"reset_at":"2026-11-01T00:00:00.000Z"}}`)
	}))
	defer srv.Close()

	c := newTestClient(t, srv, WithUserAgent("silo/1"))
	res, err := c.ListNGXSymbols(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if res.Meta.Plan != "free" || res.Meta.CallsRemaining != 2996 || res.Meta.ResetAt.IsZero() {
		t.Errorf("meta not decoded: %+v", res.Meta)
	}
}

func TestErrors_MapToSentinels(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		body     string
		sentinel error
		code     model.ErrorCode
	}{
		{"plan", 403, `{"success":false,"error":{"code":"PLAN_REQUIRED","message":"needs hobby","required_plan":"hobby","current_plan":"free"}}`, model.ErrPlanRequired, model.CodePlanRequired},
		{"bad key", 401, `{"success":false,"error":{"code":"INVALID_API_KEY","message":"revoked"}}`, model.ErrUnauthorized, model.CodeInvalidAPIKey},
		{"quota", 429, `{"success":false,"error":{"code":"QUOTA_EXCEEDED","message":"spent"},"meta":{"plan":"free","calls_used":3000,"calls_remaining":0,"reset_at":"2026-11-01T00:00:00.000Z"}}`, model.ErrQuotaExceeded, model.CodeQuotaExceeded},
		{"history", 403, `{"success":false,"error":{"code":"HISTORY_LIMIT","message":"too old"}}`, model.ErrPlanRequired, model.CodeHistoryLimit},
		{"not found", 404, `{"success":false,"error":{"code":"NOT_FOUND","message":"no such company"}}`, model.ErrNotFound, model.CodeNotFound},
		{"server", 500, `{"success":false,"error":{"code":"SERVER_ERROR","message":"boom"}}`, model.ErrServer, model.CodeServerError},
		{"gateway html", 502, `<html>Bad Gateway</html>`, model.ErrServer, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				writeJSON(w, tc.status, tc.body)
			}))
			defer srv.Close()
			c := newTestClient(t, srv, WithMaxRetries(0))

			_, err := c.GetNGXCompany(context.Background(), "DANGCEM")
			if !errors.Is(err, tc.sentinel) {
				t.Fatalf("errors.Is(%v, sentinel) = false", err)
			}
			var apiErr *model.APIError
			if !errors.As(err, &apiErr) {
				t.Fatalf("want *model.APIError, got %T", err)
			}
			if apiErr.Code != tc.code || apiErr.StatusCode != tc.status {
				t.Errorf("code/status = %q/%d, want %q/%d", apiErr.Code, apiErr.StatusCode, tc.code, tc.status)
			}
			if tc.name == "plan" && (apiErr.RequiredPlan != "hobby" || apiErr.CurrentPlan != "free") {
				t.Errorf("plans not captured: %+v", apiErr)
			}
			if tc.name == "quota" && (apiErr.Meta == nil || apiErr.Meta.CallsRemaining != 0) {
				t.Errorf("quota meta not captured: %+v", apiErr.Meta)
			}
			if tc.sentinel != model.ErrRateLimited && errors.Is(err, model.ErrRateLimited) {
				t.Error("matched ErrRateLimited unexpectedly")
			}
		})
	}
}

func TestRetry_RateLimitedThenSuccess(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.Header().Set("Retry-After", "2")
			writeJSON(w, 429, `{"success":false,"error":{"code":"RATE_LIMITED","message":"slow down"}}`)
			return
		}
		writeJSON(w, 200, `{"success":true,"data":{"data":[],"count":0}}`)
	}))
	defer srv.Close()

	var slept time.Duration
	c := newTestClient(t, srv)
	c.sleep = func(_ context.Context, d time.Duration) error { slept = d; return nil }

	if _, err := c.ListNGXSymbols(context.Background()); err != nil {
		t.Fatal(err)
	}
	if calls.Load() != 2 {
		t.Errorf("calls = %d, want 2", calls.Load())
	}
	if slept != 2*time.Second {
		t.Errorf("slept %v, want Retry-After of 2s", slept)
	}
}

func TestRetry_LongRetryAfterIsSurfacedNotSlept(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("Retry-After", "60")
		writeJSON(w, 429, `{"success":false,"error":{"code":"RATE_LIMITED","message":"slow down"}}`)
	}))
	defer srv.Close()
	c := newTestClient(t, srv) // default max wait 5s < 60s

	_, err := c.ListNGXSymbols(context.Background())
	var apiErr *model.APIError
	if !errors.As(err, &apiErr) || !errors.Is(err, model.ErrRateLimited) {
		t.Fatalf("want rate limited APIError, got %v", err)
	}
	if apiErr.RetryAfter != 60*time.Second {
		t.Errorf("RetryAfter = %v, want 60s", apiErr.RetryAfter)
	}
	if calls.Load() != 1 {
		t.Errorf("calls = %d, want 1 (no blocking retry)", calls.Load())
	}
}

func TestRetry_NeverRetriesQuotaOrPlanErrors(t *testing.T) {
	for _, tc := range []struct {
		status int
		body   string
	}{
		{429, `{"success":false,"error":{"code":"QUOTA_EXCEEDED","message":"spent"}}`},
		{403, `{"success":false,"error":{"code":"PLAN_REQUIRED","message":"nope"}}`},
	} {
		var calls atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			writeJSON(w, tc.status, tc.body)
		}))
		c := newTestClient(t, srv, WithMaxRetries(3))
		_, _ = c.ListNGXSymbols(context.Background())
		srv.Close()
		if calls.Load() != 1 {
			t.Errorf("%s: calls = %d, want exactly 1", tc.body, calls.Load())
		}
	}
}

func TestRetry_ServerErrorExhaustsRetries(t *testing.T) {
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		writeJSON(w, 503, `{"success":false,"error":{"code":"SERVER_ERROR","message":"down"}}`)
	}))
	defer srv.Close()
	c := newTestClient(t, srv, WithMaxRetries(2))

	_, err := c.ListNGXSymbols(context.Background())
	if !errors.Is(err, model.ErrServer) {
		t.Fatalf("want ErrServer, got %v", err)
	}
	if calls.Load() != 3 {
		t.Errorf("calls = %d, want 3 (1 + 2 retries)", calls.Load())
	}
}

func TestContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer srv.Close()
	c := newTestClient(t, srv)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, err := c.ListNGXSymbols(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want context deadline error, got %v", err)
	}
}

func TestErrorMessageNeverContainsAPIKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 401, `{"success":false,"error":{"code":"INVALID_API_KEY","message":"nope"}}`)
	}))
	defer srv.Close()
	c := newTestClient(t, srv, WithMaxRetries(0))
	_, err := c.ListNGXSymbols(context.Background())
	if err == nil || strings.Contains(err.Error(), "ngm_live_test") {
		t.Fatalf("error leaks the key or is nil: %v", err)
	}
}

func TestLimiter_PacesRequests(t *testing.T) {
	now := time.Unix(0, 0)
	l := newLimiter(60, 1) // 1 token/sec, burst 1
	l.now = func() time.Time { return now }
	l.last = now

	if err := l.wait(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Bucket is empty: advance the fake clock so the next wait is satisfied.
	now = now.Add(1100 * time.Millisecond)
	done := make(chan error, 1)
	go func() { done <- l.wait(context.Background()) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("limiter did not release after the clock advanced")
	}
}

func TestLimiter_RespectsContext(t *testing.T) {
	l := newLimiter(1, 1)
	_ = l.wait(context.Background()) // drain the burst
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := l.wait(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("want deadline exceeded, got %v", err)
	}
}
