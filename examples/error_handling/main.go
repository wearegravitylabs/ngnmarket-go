// Command error_handling shows how to branch on the SDK's errors. It calls an
// endpoint that needs a paid plan so that, on a Free key, you see the
// plan-required path; the same switch handles every other failure.
//
//	NGNMARKET_API_KEY=ngm_live_... go run ./examples/error_handling
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
	// Retries and pacing are configured once, on the client.
	client, err := api.New(
		os.Getenv("NGNMARKET_API_KEY"),
		api.WithRateLimit(30, 0),                // stay under the Free plan's 30 requests/minute
		api.WithMaxRetries(2),                   // retry 429/5xx, never quota or plan errors
		api.WithMaxRetryWait(5*time.Second),     // don't block longer than this; surface RetryAfter instead
		api.WithUserAgent("error-handling/0.1"), // shows up in NGN Market's request logs
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.GetNGXCompany(ctx, "DANGCEM") // needs Hobby or above

	switch {
	case err == nil:
		fmt.Println("ok: this key's plan includes company detail")

	case errors.Is(err, model.ErrPlanRequired):
		var apiErr *model.APIError
		_ = errors.As(err, &apiErr)
		fmt.Printf("upgrade needed: this endpoint requires %q, the key is on %q\n", apiErr.RequiredPlan, apiErr.CurrentPlan)

	case errors.Is(err, model.ErrRateLimited):
		var apiErr *model.APIError
		_ = errors.As(err, &apiErr)
		fmt.Printf("slow down: try again in %s\n", apiErr.RetryAfter)

	case errors.Is(err, model.ErrQuotaExceeded):
		fmt.Println("monthly quota spent: retrying will not help until it resets")

	case errors.Is(err, model.ErrUnauthorized):
		fmt.Println("the API key is missing, revoked or IP-restricted")

	case errors.Is(err, model.ErrNotFound):
		fmt.Println("unknown symbol")

	default:
		fmt.Printf("unexpected failure (network, 5xx after retries, bad response): %v\n", err)
	}
}
