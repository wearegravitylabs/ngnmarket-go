package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/wearegravitylabs/ngnmarket-go/model"
)

const usListFixture = `{"success":true,"data":{"data":[{"symbol":"MSFT","name":"MICROSOFT CORP","security_type":"stock","sector":"Technology","industry":null,"country":"US","shares_outstanding":7425545491,"logo_url":"https://cdn.jsdelivr.net/npm/us-stock-logos@1.0.0/dist/png/MSFT.png","last_price":529.76,"prev_close":529.3,"change_abs":0.46,"change_pct":0.0869,"volume":16005856,"market_cap":3933756979312.16,"high_52wk":553.72,"low_52wk":349.2,"is_delayed":1,"quote_time":"2026-10-07T20:00:00.000Z"}],"pagination":{"page":1,"limit":100,"total":1,"pages":1,"has_next":false,"has_prev":false}},"meta":{"plan":"free","calls_used":10,"calls_remaining":2990,"reset_at":"2026-11-01T00:00:00.000Z","market":"US","currency":"USD","as_of":"2026-10-07T20:00:00.000Z","newest":"2026-10-07T20:00:00.000Z"}}`

func TestListUSTickers_DecodesQuoteAndMeta(t *testing.T) {
	c := serve(t, func(r *http.Request) {
		if r.URL.Path != "/us/tickers" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "etf" || r.URL.Query().Get("search") != "SPY" {
			t.Errorf("query = %v", r.URL.Query())
		}
	}, 200, usListFixture)

	res, err := c.ListUSTickers(context.Background(), &model.ListTickersParams{Type: model.SecurityETF, Search: "SPY"})
	if err != nil {
		t.Fatal(err)
	}
	tk := res.Tickers[0]
	if tk.Symbol != "MSFT" || tk.LastPrice != 529.76 || !bool(tk.IsDelayed) || tk.QuoteTime == nil {
		t.Errorf("unexpected ticker: %+v", tk)
	}
	if res.Meta.Market != "US" || res.Meta.Currency != "USD" || res.Meta.AsOf == nil {
		t.Errorf("US meta not decoded: %+v", res.Meta)
	}
}

func TestFindUSTicker_IsCaseSensitive(t *testing.T) {
	c := serve(t, nil, 200, usListFixture)
	if tk, err := c.FindUSTicker(context.Background(), "MSFT"); err != nil || tk.Symbol != "MSFT" {
		t.Fatalf("got %+v, %v", tk, err)
	}
	if _, err := c.FindUSTicker(context.Background(), "msft"); !errors.Is(err, model.ErrNotFound) {
		t.Errorf("lowercase must not match a case-sensitive symbol, got %v", err)
	}
}

func TestFindUSTicker_ScansLaterPages(t *testing.T) {
	pages := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pages++
		if r.URL.Query().Get("page") == "2" {
			writeJSON(w, 200, `{"success":true,"data":{"data":[{"symbol":"A"}],"pagination":{"page":2,"has_next":false}}}`)
			return
		}
		writeJSON(w, 200, `{"success":true,"data":{"data":[{"symbol":"AA"}],"pagination":{"page":1,"has_next":true}}}`)
	}))
	defer srv.Close()
	c := newTestClient(t, srv)

	tk, err := c.FindUSTicker(context.Background(), "A")
	if err != nil || tk.Symbol != "A" || pages != 2 {
		t.Fatalf("tk=%+v pages=%d err=%v", tk, pages, err)
	}
}

func TestListUSSymbols(t *testing.T) {
	c := serve(t, nil, 200, `{"success":true,"data":{"data":[{"symbol":"A","name":"AGILENT TECHNOLOGIES INC","security_type":"stock"}],"count":1},"meta":{"plan":"free","calls_used":1,"calls_remaining":2999,"reset_at":"2026-11-01T00:00:00.000Z"}}`)
	res, err := c.ListUSSymbols(context.Background())
	if err != nil || res.Count != 1 || res.Tickers[0].SecurityType != model.SecurityStock {
		t.Fatalf("got %+v, %v", res, err)
	}
}

func TestGetUSClosePriceOnOrBefore(t *testing.T) {
	c := serve(t, func(r *http.Request) {
		if r.URL.Path != "/us/tickers/MSFT/chart" {
			t.Errorf("path = %s", r.URL.Path)
		}
	}, 200, `{"success":true,"data":{"symbol":"MSFT","price_adjustment":"unadjusted","data":[
		{"date":"2026-03-05","close":400.5},{"date":"2026-03-06","close":null},{"date":"2026-03-09","close":405.25}]}}`)

	// close:null rows are skipped; the 9th is the latest row on/before the 10th.
	pt, err := c.GetUSClosePriceOnOrBefore(context.Background(), "MSFT", time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC))
	if err != nil || pt.Date != "2026-03-09" || pt.Close != 405.25 {
		t.Fatalf("got %+v, %v", pt, err)
	}
}
