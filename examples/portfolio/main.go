// Command portfolio values a small mixed Nigerian and US holding, the way a
// portfolio tracker would: one call prices every NGX stock, one lookup per US stock.
//
//	NGNMARKET_API_KEY=ngm_live_... go run ./examples/portfolio
//
// Works on the Free plan. Note the two currencies are shown separately: NGX
// prices are in NGN and US prices in USD, so converting needs an FX rate.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/wearegravitylabs/ngnmarket-go/api"
	"github.com/wearegravitylabs/ngnmarket-go/model"
)

type holding struct {
	Symbol   string
	Quantity float64
	AvgCost  float64 // price paid per unit, in the stock's own currency
}

func main() {
	ngx := []holding{{"DANGCEM", 500, 285}, {"GTCO", 2000, 55.4}, {"MTNN", 800, 198.5}}
	us := []holding{{"MSFT", 10, 310}, {"AAPL", 25, 150}}

	client, err := api.New(os.Getenv("NGNMARKET_API_KEY"), api.WithRateLimit(30, 0))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	// NGX: a single call returns every company, so pricing 3 or 300 holdings costs the same.
	companies, _, err := client.ListAllNGXCompanies(ctx)
	if err != nil {
		log.Fatalf("ngx prices: %v", err)
	}
	bySymbol := make(map[string]model.Company, len(companies))
	for _, co := range companies {
		bySymbol[strings.ToUpper(co.Symbol)] = co
	}

	fmt.Println("NIGERIAN EXCHANGE (NGN)")
	var cost, value float64
	for _, h := range ngx {
		co, ok := bySymbol[h.Symbol]
		if !ok {
			fmt.Printf("  %-8s not found\n", h.Symbol)
			continue
		}
		c, v := h.Quantity*h.AvgCost, h.Quantity*co.Price
		cost, value = cost+c, value+v
		fmt.Printf("  %-8s %6.0f x %9.2f = %14.2f   gain %+10.2f\n", h.Symbol, h.Quantity, co.Price, v, v-c)
	}
	fmt.Printf("  %-8s cost %.2f, value %.2f, gain %+.2f\n\n", "TOTAL", cost, value, value-cost)

	// US: no batch lookup on the Free plan, so each distinct symbol costs at least one call.
	fmt.Println("US MARKET (USD, delayed)")
	cost, value = 0, 0
	for _, h := range us {
		tk, err := client.FindUSTicker(ctx, h.Symbol)
		if errors.Is(err, model.ErrNotFound) {
			fmt.Printf("  %-8s not found\n", h.Symbol)
			continue
		} else if err != nil {
			log.Fatalf("us price for %s: %v", h.Symbol, err)
		}
		c, v := h.Quantity*h.AvgCost, h.Quantity*tk.LastPrice
		cost, value = cost+c, value+v
		fmt.Printf("  %-8s %6.0f x %9.2f = %14.2f   gain %+10.2f\n", h.Symbol, h.Quantity, tk.LastPrice, v, v-c)
	}
	fmt.Printf("  %-8s cost %.2f, value %.2f, gain %+.2f\n", "TOTAL", cost, value, value-cost)
}
