//go:build integration

package api_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/wearegravitylabs/ngnmarket-go/api"
	"github.com/wearegravitylabs/ngnmarket-go/model"
)

// Run against the real API with:
//
//	NGNMARKET_API_KEY=ngm_live_... go test -tags integration -run Live -v ./api/
//
// It makes about six calls, so it is safe on the Free plan. Endpoints that need
// a higher plan are skipped when the API answers PLAN_REQUIRED.
func liveClient(t *testing.T) api.RemoteCalls {
	t.Helper()
	key := os.Getenv("NGNMARKET_API_KEY")
	if key == "" {
		t.Skip("NGNMARKET_API_KEY not set")
	}
	c, err := api.New(key, api.WithRateLimit(20, 2))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestLive_NGX_PricesWholeMarketInOneCall(t *testing.T) {
	c := liveClient(t)
	all, meta, err := c.ListAllNGXCompanies(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(all) < 100 {
		t.Fatalf("expected the whole exchange (~150 companies), got %d", len(all))
	}
	priced := 0
	for _, co := range all {
		if co.Symbol == "" || co.Name == "" {
			t.Errorf("incomplete row: %+v", co)
		}
		if co.Price > 0 {
			priced++
		}
	}
	t.Logf("%d companies, %d priced; plan=%s remaining=%d", len(all), priced, meta.Plan, meta.CallsRemaining)
	if priced == 0 {
		t.Error("no company has a price")
	}
}

func TestLive_NGX_Symbols(t *testing.T) {
	c := liveClient(t)
	res, err := c.ListNGXSymbols(context.Background())
	if err != nil || res.Count == 0 || len(res.Companies) != res.Count {
		t.Fatalf("res=%+v err=%v", res, err)
	}
}

func TestLive_US_FindTicker(t *testing.T) {
	c := liveClient(t)
	tk, err := c.FindUSTicker(context.Background(), "MSFT")
	if err != nil {
		t.Fatal(err)
	}
	if tk.LastPrice <= 0 || tk.QuoteTime == nil || tk.LogoURL == "" {
		t.Errorf("unexpected ticker: %+v", tk)
	}
	t.Logf("MSFT %.2f USD quoted %s delayed=%v", tk.LastPrice, tk.QuoteTime.Format(time.RFC3339), tk.IsDelayed)
}

func TestLive_US_FindTicker_NotFound(t *testing.T) {
	c := liveClient(t)
	if _, err := c.FindUSTicker(context.Background(), "ZZZZNOTREAL"); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestLive_US_Symbols(t *testing.T) {
	c := liveClient(t)
	res, err := c.ListUSSymbols(context.Background())
	if err != nil || res.Count < 5000 {
		t.Fatalf("count=%d err=%v", res.Count, err)
	}
}

func TestLive_HistoryEndpoints_NeedHobby(t *testing.T) {
	c := liveClient(t)
	day := time.Now().AddDate(0, -3, 0)

	_, err := c.GetNGXClosePriceOnOrBefore(context.Background(), "DANGCEM", day)
	if errors.Is(err, model.ErrPlanRequired) {
		t.Skipf("key's plan has no history access (expected on Free): %v", err)
	}
	if err != nil {
		t.Fatal(err)
	}
	pt, err := c.GetUSClosePriceOnOrBefore(context.Background(), "MSFT", day)
	if err != nil || pt.Close <= 0 {
		t.Fatalf("US history: %+v %v", pt, err)
	}
}
