package model

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// ErrorCode is the machine-readable error code returned by the NGN Market API.
type ErrorCode string

// Error codes documented at https://docs.ngnmarket.com/errors.
const (
	CodeMissingAPIKey ErrorCode = "MISSING_API_KEY"
	CodeInvalidAPIKey ErrorCode = "INVALID_API_KEY"
	CodePlanRequired  ErrorCode = "PLAN_REQUIRED"
	CodeIPNotAllowed  ErrorCode = "IP_NOT_ALLOWED"
	CodeRateLimited   ErrorCode = "RATE_LIMITED"
	CodeQuotaExceeded ErrorCode = "QUOTA_EXCEEDED"
	CodeHistoryLimit  ErrorCode = "HISTORY_LIMIT"
	CodeNotFound      ErrorCode = "NOT_FOUND"
	CodeServerError   ErrorCode = "SERVER_ERROR"
)

// Sentinel errors for use with errors.Is. Any *APIError whose code falls in the
// group matches the sentinel, so callers can branch without inspecting codes:
//
//	if errors.Is(err, model.ErrRateLimited) { ... }
var (
	// ErrUnauthorized matches MISSING_API_KEY, INVALID_API_KEY and IP_NOT_ALLOWED.
	ErrUnauthorized = errors.New("ngnmarket: unauthorized")
	// ErrPlanRequired matches PLAN_REQUIRED and HISTORY_LIMIT: the API key's plan
	// does not include the endpoint (or the requested history depth).
	ErrPlanRequired = errors.New("ngnmarket: plan upgrade required")
	// ErrRateLimited matches RATE_LIMITED: the per-minute cap was hit. Retry later.
	ErrRateLimited = errors.New("ngnmarket: rate limited")
	// ErrQuotaExceeded matches QUOTA_EXCEEDED: the monthly call quota is spent.
	// Retrying will not help until the quota resets.
	ErrQuotaExceeded = errors.New("ngnmarket: monthly quota exceeded")
	// ErrNotFound matches NOT_FOUND (unknown symbol or route).
	ErrNotFound = errors.New("ngnmarket: not found")
	// ErrServer matches SERVER_ERROR and any other 5xx response.
	ErrServer = errors.New("ngnmarket: server error")
	// ErrNoData is returned by the price-lookup helpers when the API answered but
	// has no price on or before the requested date (for example the date is older
	// than the key's plan allows, or before the security was listed).
	ErrNoData = errors.New("ngnmarket: no price data for the requested date")
)

// APIError is returned when the API responds with an error status.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int
	// Code is the API's machine-readable error code. Empty if the response body
	// was not a valid API error envelope (for example a proxy error page).
	Code ErrorCode
	// Message is the human-readable message from the API.
	Message string
	// RequiredPlan and CurrentPlan are set for PLAN_REQUIRED errors.
	RequiredPlan string
	CurrentPlan  string
	// RetryAfter is how long the API asked the caller to wait, when known
	// (from the Retry-After header). Set for RATE_LIMITED errors.
	RetryAfter time.Duration
	// Meta is the quota state attached to the error response, when present
	// (QUOTA_EXCEEDED responses include it).
	Meta *Meta
}

func (e *APIError) Error() string {
	code := string(e.Code)
	if code == "" {
		code = http.StatusText(e.StatusCode)
	}
	msg := fmt.Sprintf("ngnmarket: %s (HTTP %d): %s", code, e.StatusCode, e.Message)
	if e.Code == CodePlanRequired && e.RequiredPlan != "" {
		msg += fmt.Sprintf(" [requires %s, have %s]", e.RequiredPlan, e.CurrentPlan)
	}
	return msg
}

// Is reports whether the error belongs to one of the sentinel groups.
func (e *APIError) Is(target error) bool {
	switch target {
	case ErrUnauthorized:
		return e.Code == CodeMissingAPIKey || e.Code == CodeInvalidAPIKey || e.Code == CodeIPNotAllowed
	case ErrPlanRequired:
		return e.Code == CodePlanRequired || e.Code == CodeHistoryLimit
	case ErrRateLimited:
		return e.Code == CodeRateLimited
	case ErrQuotaExceeded:
		return e.Code == CodeQuotaExceeded
	case ErrNotFound:
		return e.Code == CodeNotFound || (e.Code == "" && e.StatusCode == http.StatusNotFound)
	case ErrServer:
		return e.Code == CodeServerError || e.StatusCode >= 500
	}
	return false
}

// Retryable reports whether a failed call is safe and useful to retry: the
// per-minute rate limit and server errors. Monthly quota, auth, plan and other
// 4xx errors are not.
func (e *APIError) Retryable() bool {
	return e.Code == CodeRateLimited || e.StatusCode >= 500
}
