package model

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestFlexBool(t *testing.T) {
	for in, want := range map[string]bool{"1": true, "true": true, "0": false, "false": false, "null": false} {
		var b FlexBool
		if err := b.UnmarshalJSON([]byte(in)); err != nil || bool(b) != want {
			t.Errorf("%s -> %v (%v), want %v", in, b, err, want)
		}
	}
	var b FlexBool
	if err := b.UnmarshalJSON([]byte(`"yes"`)); err == nil {
		t.Error("invalid value: want error")
	}
}

func TestValues_OmitZeroValues(t *testing.T) {
	var nilParams *ListCompaniesParams
	if len(nilParams.Values()) != 0 {
		t.Error("nil params must encode to nothing")
	}
	cap1 := 1000.5
	q := (&ListCompaniesParams{Page: 2, Search: "gt", MinMarketCap: &cap1}).Values()
	if q.Get("page") != "2" || q.Get("search") != "gt" || q.Get("minMarketCap") != "1000.5" || q.Has("limit") || q.Has("sector") {
		t.Errorf("query = %v", q)
	}

	us := (&ListTickersParams{Type: SecurityETF, Limit: 10}).Values()
	if us.Get("type") != "etf" || us.Get("limit") != "10" || us.Has("page") {
		t.Errorf("us query = %v", us)
	}

	day := time.Date(2026, 3, 11, 23, 0, 0, 0, time.FixedZone("WAT", 3600))
	chart := (&NGXChartParams{From: day}).Values()
	if chart.Get("format") != "detailed" || chart.Get("from") != "2026-03-11" || chart.Has("to") {
		t.Errorf("chart query = %v (dates are UTC calendar days)", chart)
	}
}

func TestNGXChartPoint_DayAndClose(t *testing.T) {
	ms := NGXChartPoint{Timestamp: 1773014400000} // milliseconds
	if ms.Day() != "2026-03-09" {
		t.Errorf("ms timestamp -> %q", ms.Day())
	}
	if (NGXChartPoint{}).Day() != "" {
		t.Error("empty point has no day")
	}
	p := 5.0
	if v, ok := (NGXChartPoint{Price: &p}).ClosePrice(); !ok || v != 5 {
		t.Errorf("price alias: %v %v", v, ok)
	}
	if _, ok := (NGXChartPoint{}).ClosePrice(); ok {
		t.Error("no close, no price: want !ok")
	}
}

func TestAPIError_SentinelGroups(t *testing.T) {
	cases := map[ErrorCode][]error{
		CodeMissingAPIKey: {ErrUnauthorized},
		CodeIPNotAllowed:  {ErrUnauthorized},
		CodeHistoryLimit:  {ErrPlanRequired},
		CodeRateLimited:   {ErrRateLimited},
		CodeQuotaExceeded: {ErrQuotaExceeded},
	}
	for code, want := range cases {
		e := &APIError{Code: code}
		for _, s := range want {
			if !errors.Is(e, s) {
				t.Errorf("%s should match %v", code, s)
			}
		}
	}
	if !errors.Is(&APIError{StatusCode: http.StatusNotFound}, ErrNotFound) {
		t.Error("a bare 404 should match ErrNotFound")
	}
	if (&APIError{Code: CodeQuotaExceeded}).Retryable() || !(&APIError{Code: CodeRateLimited}).Retryable() {
		t.Error("only rate limits and 5xx are retryable")
	}
}
