// Package model holds the request parameters, response types and errors of the
// NGN Market API. It has no dependencies, so it can be imported on its own.
package model

import (
	"bytes"
	"errors"
	"net/url"
	"strconv"
	"time"
)

// DateLayout is the date format the API uses ("2006-01-02").
const DateLayout = "2006-01-02"

type (
	// Meta is the quota state the API attaches to every response. US responses
	// add Market, Currency and the quote freshness window.
	Meta struct {
		Plan           string    `json:"plan"`
		CallsUsed      int       `json:"calls_used"`
		CallsRemaining int       `json:"calls_remaining"`
		ResetAt        time.Time `json:"reset_at"`
		// RetryAfter is in seconds; only present on rate-limit errors.
		RetryAfter int `json:"retry_after,omitempty"`

		// US endpoints only.
		Market   string     `json:"market,omitempty"`
		Currency string     `json:"currency,omitempty"`
		AsOf     *time.Time `json:"as_of,omitempty"`
		Newest   *time.Time `json:"newest,omitempty"`
	}

	// Pagination describes a page of a paginated list response.
	Pagination struct {
		Page    int  `json:"page"`
		Limit   int  `json:"limit"`
		Total   int  `json:"total"`
		Pages   int  `json:"pages"`
		HasNext bool `json:"has_next"`
		HasPrev bool `json:"has_prev"`
	}

	// PricePoint is a single day's closing price, normalised across NGX and US charts.
	PricePoint struct {
		// Date is the trading day, "YYYY-MM-DD".
		Date string
		// Close is the closing price.
		Close float64
	}

	// SortOrder is the direction of a list sort.
	SortOrder string

	// FlexBool decodes a JSON boolean that the API sometimes sends as 0/1.
	FlexBool bool
)

// Sort directions.
const (
	Ascending  SortOrder = "asc"
	Descending SortOrder = "desc"
)

// UnmarshalJSON implements json.Unmarshaler.
func (b *FlexBool) UnmarshalJSON(data []byte) error {
	switch string(bytes.TrimSpace(data)) {
	case "true", "1":
		*b = true
	case "false", "0", "null":
		*b = false
	default:
		return errors.New("ngnmarket: invalid boolean value " + string(data))
	}
	return nil
}

// ─── query-string helpers (unexported; used by the Values methods) ────────────

func setInt(q url.Values, key string, v int) {
	if v > 0 {
		q.Set(key, strconv.Itoa(v))
	}
}

func setStr(q url.Values, key, v string) {
	if v != "" {
		q.Set(key, v)
	}
}

func setFloat(q url.Values, key string, v *float64) {
	if v != nil {
		q.Set(key, strconv.FormatFloat(*v, 'f', -1, 64))
	}
}

func setDate(q url.Values, key string, t time.Time) {
	if !t.IsZero() {
		q.Set(key, t.UTC().Format(DateLayout))
	}
}
