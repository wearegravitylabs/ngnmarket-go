// Package ngnmarket is the module root of the NGN Market Go SDK
// (https://docs.ngnmarket.com): Nigerian Exchange (NGX) and US market data.
//
// The SDK lives in sub-packages:
//
//   - github.com/wearegravitylabs/ngnmarket-go/api    the client (api.New) and the RemoteCalls interface
//   - github.com/wearegravitylabs/ngnmarket-go/model   request parameters, responses and errors
//   - github.com/wearegravitylabs/ngnmarket-go/api/mock a gomock mock of RemoteCalls for your tests
//
// Runnable samples are in the examples directory.
//
// A minimal program:
//
//	client, err := api.New(os.Getenv("NGNMARKET_API_KEY"))
//	if err != nil { ... }
//	companies, meta, err := client.ListAllNGXCompanies(ctx) // prices the whole exchange in one call
//	msft, err := client.FindUSTicker(ctx, "MSFT")
//
// # Plans
//
// The Free plan can list and search tickers and read their current price.
// Company/ticker detail and all price history need Hobby or above; those calls
// fail with an error matching model.ErrPlanRequired on a Free key.
//
// # Errors
//
// Failed calls return a *model.APIError carrying the HTTP status, the API's error
// code, and plan and retry hints. Use errors.Is with the model sentinels
// (ErrRateLimited, ErrQuotaExceeded, ErrPlanRequired, ErrUnauthorized,
// ErrNotFound, ErrServer) to branch without matching strings.
//
// # Rate limits and retries
//
// Plans have a monthly quota and a per-minute cap. The SDK retries RATE_LIMITED
// and 5xx responses with backoff (honouring Retry-After up to api.WithMaxRetryWait)
// and never retries QUOTA_EXCEEDED. Use api.WithRateLimit to pace requests below
// your plan's per-minute cap so 429s do not happen in the first place. Every
// result carries a model.Meta with the remaining monthly quota.
package ngnmarket
