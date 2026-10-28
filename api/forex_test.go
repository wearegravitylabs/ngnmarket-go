package api

import (
	"context"
	"net/http"
	"testing"
)

func TestGetForexRates(t *testing.T) {
	c := serve(t, func(r *http.Request) {
		if r.URL.Path != "/forex/current" {
			t.Errorf("path = %s", r.URL.Path)
		}
	}, 200, `{"success":true,"data":{"target":"NGN","date":"2026-10-16T00:00:00.000Z","rates":[
		{"currency":"AUD","rate":943.44,"inverse_rate":0.00106,"daily_change":2.812,"daily_change_percent":0.2989,"last_updated":"2026-10-16T00:00:00.000Z"},
		{"currency":"XXX","rate":null,"inverse_rate":null,"daily_change":null,"daily_change_percent":null,"last_updated":null}]},
		"meta":{"plan":"free","calls_used":1,"calls_remaining":2999,"reset_at":"2026-11-01T00:00:00.000Z"}}`)

	res, err := c.GetForexRates(context.Background())
	if err != nil {
		t.Fatalf("err = %v", err)
	}
	if res.Target != "NGN" || res.Date == nil || len(res.Rates) != 2 {
		t.Fatalf("got %+v", res)
	}
	if aud := res.Rates[0]; aud.Currency != "AUD" || aud.Rate != 943.44 || aud.LastUpdated == nil {
		t.Errorf("AUD = %+v", aud)
	}
	// null values must not crash: numbers become 0, the timestamp stays nil.
	if xxx := res.Rates[1]; xxx.Rate != 0 || xxx.LastUpdated != nil {
		t.Errorf("null row = %+v", xxx)
	}
}
