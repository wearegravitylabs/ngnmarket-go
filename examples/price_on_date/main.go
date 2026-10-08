// Command price_on_date looks up what a stock closed at on a past date, the way
// Silo works out the price paid for a lot, and falls back gracefully when the
// key's plan has no price history.
//
//	NGNMARKET_API_KEY=ngm_live_... go run ./examples/price_on_date NG DANGCEM 2026-03-11
//	NGNMARKET_API_KEY=ngm_live_... go run ./examples/price_on_date US MSFT 2026-03-11
//
// Price history needs the Hobby plan or above. On the Free plan this prints the
// fallback message, which is the cue to ask the user for the price they paid.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/wearegravitylabs/ngnmarket-go/api"
	"github.com/wearegravitylabs/ngnmarket-go/model"
)

func main() {
	if len(os.Args) != 4 {
		log.Fatal("usage: price_on_date NG|US SYMBOL YYYY-MM-DD")
	}
	market, symbol := os.Args[1], os.Args[2]
	day, err := time.Parse(model.DateLayout, os.Args[3])
	if err != nil {
		log.Fatalf("bad date: %v", err)
	}

	client, err := api.New(os.Getenv("NGNMARKET_API_KEY"), api.WithRateLimit(30, 0))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var pt model.PricePoint
	switch market {
	case "NG":
		pt, err = client.GetNGXClosePriceOnOrBefore(ctx, symbol, day)
	case "US":
		pt, err = client.GetUSClosePriceOnOrBefore(ctx, symbol, day)
	default:
		log.Fatal("market must be NG or US")
	}

	switch {
	case err == nil:
		// The date can differ from the one asked for: weekends and holidays use the trading day before.
		fmt.Printf("%s closed at %.2f on %s (asked for %s)\n", symbol, pt.Close, pt.Date, os.Args[3])
	case errors.Is(err, model.ErrPlanRequired):
		fmt.Println("This API key's plan has no price history (needs Hobby or above).")
		fmt.Println("Ask the user to enter the price they paid instead.")
	case errors.Is(err, model.ErrNoData):
		fmt.Println("No price on or before that date (older than the plan's history, or before listing).")
	case errors.Is(err, model.ErrNotFound):
		fmt.Printf("%q is not a listed symbol.\n", symbol)
	default:
		log.Fatalf("lookup failed: %v", err)
	}
}
