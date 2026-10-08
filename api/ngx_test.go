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

// Fixtures mirror real responses captured from the live API (Free plan).
const ngxListFixture = `{"success":true,"data":{"data":[{"id":134,"symbol":"AIRTELAFRI","name":"AIRTEL AFRICA PLC","logo_url":"https://cdn.jsdelivr.net/gh/ngnmarket/ngx-logos/dist/png/AIRTELAFRI.png","sector":"ICT","sub_sector":"Telecommunications Services","market_classification":"Main Board","shares_outstanding":3758151504,"website":null,"price":6300,"prev_close":6300,"day_high":null,"day_low":null,"volume":2791,"market_cap":23676354475200,"price_change":0,"price_change_percent":0,"change_7d_percent":0,"change_52w_percent":172.6683,"change_1m_percent":0,"change_ytd_percent":177.533,"high_52wk":6300,"low_52wk":2497,"last_updated":"2026-10-07T15:40:06.000Z"}],"pagination":{"page":1,"limit":1,"total":1,"pages":1,"has_next":false,"has_prev":false}},"meta":{"plan":"free","calls_used":9,"calls_remaining":2991,"reset_at":"2026-11-01T00:00:00.000Z"}}`

// serve starts a server answering every request with status/body, after check.
func serve(t *testing.T, check func(r *http.Request), status int, body string) *Call {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if check != nil {
			check(r)
		}
		writeJSON(w, status, body)
	}))
	t.Cleanup(srv.Close)
	return newTestClient(t, srv, WithMaxRetries(0))
}

func TestListNGXCompanies(t *testing.T) {
	c := serve(t, func(r *http.Request) {
		if r.URL.Path != "/companies" {
			t.Errorf("path = %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("search") != "airtel" || q.Get("limit") != "5" || q.Get("order") != "desc" || q.Get("sort") != "market_cap" {
			t.Errorf("query = %v", q)
		}
		if q.Has("page") || q.Has("sector") {
			t.Errorf("zero-value params must be omitted: %v", q)
		}
	}, 200, ngxListFixture)

	res, err := c.ListNGXCompanies(context.Background(), &model.ListCompaniesParams{
		Search: "airtel", Limit: 5, Sort: model.CompanySortMarketCap, Order: model.Descending,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Companies) != 1 {
		t.Fatalf("companies = %d", len(res.Companies))
	}
	co := res.Companies[0]
	if co.Symbol != "AIRTELAFRI" || co.Price != 6300 || co.DayHigh != nil || co.LastUpdated == nil {
		t.Errorf("unexpected company: %+v", co)
	}
	if res.Meta.Plan != "free" || res.Pagination.Total != 1 {
		t.Errorf("meta/pagination: %+v %+v", res.Meta, res.Pagination)
	}
}

func TestListNGXCompanies_NilParams(t *testing.T) {
	c := serve(t, func(r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("nil params must send no query, got %q", r.URL.RawQuery)
		}
	}, 200, ngxListFixture)
	if _, err := c.ListNGXCompanies(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
}

func TestFindNGXCompany(t *testing.T) {
	c := serve(t, nil, 200, ngxListFixture)
	co, err := c.FindNGXCompany(context.Background(), "airtelafri") // case-insensitive
	if err != nil || co.Symbol != "AIRTELAFRI" {
		t.Fatalf("got %+v, %v", co, err)
	}
	if _, err := c.FindNGXCompany(context.Background(), "NOPE"); !errors.Is(err, model.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
	if _, err := c.FindNGXCompany(context.Background(), " "); err == nil {
		t.Error("blank symbol: want error")
	}
}

func TestListAllNGXCompanies_FollowsPagination(t *testing.T) {
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("limit") != "200" {
			t.Errorf("limit = %q", r.URL.Query().Get("limit"))
		}
		if r.URL.Query().Get("page") == "1" {
			writeJSON(w, 200, `{"success":true,"data":{"data":[{"symbol":"A"}],"pagination":{"page":1,"has_next":true}}}`)
			return
		}
		writeJSON(w, 200, `{"success":true,"data":{"data":[{"symbol":"B"}],"pagination":{"page":2,"has_next":false}}}`)
	}))
	defer srv.Close()
	c := newTestClient(t, srv)

	all, _, err := c.ListAllNGXCompanies(context.Background())
	if err != nil || len(all) != 2 || calls != 2 {
		t.Fatalf("all=%v calls=%d err=%v", all, calls, err)
	}
}

func TestGetNGXCompany_EscapesSymbolAndHandlesPlanError(t *testing.T) {
	c := serve(t, func(r *http.Request) {
		if r.URL.EscapedPath() != "/companies/A%2FB" {
			t.Errorf("escaped path = %s", r.URL.EscapedPath())
		}
	}, 403, `{"success":false,"error":{"code":"PLAN_REQUIRED","message":"x","required_plan":"hobby","current_plan":"free"}}`)
	if _, err := c.GetNGXCompany(context.Background(), "A/B"); !errors.Is(err, model.ErrPlanRequired) {
		t.Fatalf("want ErrPlanRequired, got %v", err)
	}
}

func TestGetNGXPriceChart_AlwaysAsksForDetailedFormat(t *testing.T) {
	c := serve(t, func(r *http.Request) {
		q := r.URL.Query()
		if q.Get("format") != "detailed" || q.Get("period") != "1y" {
			t.Errorf("query = %v", q)
		}
	}, 200, `{"success":true,"data":{"symbol":"GTCO","count":1,"data":[{"date":"2026-03-09","close":56}],"statistics":{"first_price":56}}}`)

	ch, err := c.GetNGXPriceChart(context.Background(), "GTCO", &model.NGXChartParams{Period: "1y"})
	if err != nil || len(ch.Points) != 1 || ch.Statistics.FirstPrice != 56 {
		t.Fatalf("chart = %+v, %v", ch, err)
	}
}

func TestGetNGXClosePriceOnOrBefore_PicksNearestEarlierTradingDay(t *testing.T) {
	c := serve(t, func(r *http.Request) {
		q := r.URL.Query()
		if q.Get("from") != "2026-03-01" || q.Get("to") != "2026-03-11" {
			t.Errorf("window = %s..%s", q.Get("from"), q.Get("to"))
		}
	}, 200, `{"success":true,"data":{"symbol":"GTCO","count":3,"data":[
		{"date":"2026-03-06","close":55.4},
		{"date":"2026-03-09","close":56.0},
		{"date":"2026-03-12","close":99.0}]}}`)

	// 2026-03-11 has no row (holiday): expect the 9th, never the later 12th.
	pt, err := c.GetNGXClosePriceOnOrBefore(context.Background(), "GTCO", time.Date(2026, 3, 11, 0, 0, 0, 0, time.UTC))
	if err != nil || pt.Date != "2026-03-09" || pt.Close != 56.0 {
		t.Fatalf("got %+v, %v", pt, err)
	}
}

func TestGetNGXClosePriceOnOrBefore_DerivesDateFromTimestampAndFallsBackToPrice(t *testing.T) {
	// 1773014400 = 2026-03-09T00:00:00Z ; the row only has the legacy "price" alias.
	c := serve(t, nil, 200, `{"success":true,"data":{"data":[{"timestamp":1773014400,"price":56.0}]}}`)
	pt, err := c.GetNGXClosePriceOnOrBefore(context.Background(), "GTCO", time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC))
	if err != nil || pt.Date != "2026-03-09" || pt.Close != 56.0 {
		t.Fatalf("got %+v, %v", pt, err)
	}
}

func TestClosePriceOnOrBefore_NoDataWhenHistoryClamped(t *testing.T) {
	// A plan with limited depth clamps `from`, so every row is after the date we want.
	c := serve(t, nil, 200, `{"success":true,"data":{"data":[{"date":"2026-04-01","close":10}]}}`)
	if _, err := c.GetNGXClosePriceOnOrBefore(context.Background(), "GTCO", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)); !errors.Is(err, model.ErrNoData) {
		t.Fatalf("NGX: want ErrNoData, got %v", err)
	}
	if _, err := c.GetUSClosePriceOnOrBefore(context.Background(), "MSFT", time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)); !errors.Is(err, model.ErrNoData) {
		t.Fatalf("US: want ErrNoData, got %v", err)
	}
}
