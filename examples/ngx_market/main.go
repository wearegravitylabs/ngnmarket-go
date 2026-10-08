// Command ngx_market prices the whole Nigerian Exchange with a single API call.
//
//	NGNMARKET_API_KEY=ngm_live_... go run ./examples/ngx_market
//
// Works on the Free plan.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/wearegravitylabs/ngnmarket-go/api"
)

func main() {
	client, err := api.New(os.Getenv("NGNMARKET_API_KEY"), api.WithRateLimit(30, 0))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// One call returns every listed company with its price and logo.
	companies, meta, err := client.ListAllNGXCompanies(ctx)
	if err != nil {
		log.Fatalf("list companies: %v", err)
	}

	sort.Slice(companies, func(i, j int) bool { return companies[i].MarketCap > companies[j].MarketCap })

	fmt.Printf("%d companies listed. Ten largest by market cap:\n\n", len(companies))
	fmt.Printf("%-12s %-34s %12s %8s\n", "SYMBOL", "NAME", "PRICE (NGN)", "CHANGE")
	for _, co := range companies[:10] {
		fmt.Printf("%-12s %-34.34s %12.2f %7.2f%%\n", co.Symbol, co.Name, co.Price, co.PriceChangePercent)
	}

	fmt.Printf("\nlogo for %s: %s\n", companies[0].Symbol, companies[0].LogoURL)
	fmt.Printf("plan=%s, %d calls left this month (resets %s)\n",
		meta.Plan, meta.CallsRemaining, meta.ResetAt.Format("2006-01-02"))
}
