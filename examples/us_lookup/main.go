// Command us_lookup finds a US ticker by its exact symbol and prints its quote.
//
//	NGNMARKET_API_KEY=ngm_live_... go run ./examples/us_lookup [SYMBOL]
//
// US symbols are case-sensitive. Quotes are delayed by at least 15 minutes.
// Works on the Free plan.
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
	symbol := "MSFT"
	if len(os.Args) > 1 {
		symbol = os.Args[1]
	}

	client, err := api.New(os.Getenv("NGNMARKET_API_KEY"), api.WithRateLimit(30, 0))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tk, err := client.FindUSTicker(ctx, symbol)
	switch {
	case errors.Is(err, model.ErrNotFound):
		log.Fatalf("%q is not a US ticker (symbols are case-sensitive)", symbol)
	case err != nil:
		log.Fatalf("lookup: %v", err)
	}

	fmt.Printf("%s  %s\n", tk.Symbol, tk.Name)
	fmt.Printf("  price      %.2f USD (%+.2f%% on the day)\n", tk.LastPrice, tk.ChangePct)
	fmt.Printf("  market cap %.0f\n", tk.MarketCap)
	fmt.Printf("  52 week    %.2f - %.2f\n", tk.Low52Wk, tk.High52Wk)
	if tk.QuoteTime != nil {
		fmt.Printf("  quoted at  %s (delayed: %v)\n", tk.QuoteTime.Format(time.RFC3339), bool(tk.IsDelayed))
	}
	fmt.Printf("  logo       %s\n", tk.LogoURL)
}
