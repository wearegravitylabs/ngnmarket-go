# ngnmarket-go

Go SDK for the [NGN Market API](https://docs.ngnmarket.com): Nigerian Exchange (NGX) and US market data.

- Standard library only at runtime. Go 1.22+.
- NGX and US behind one client (same host and key; US lives under `/us`).
- Typed errors with `errors.Is` sentinels, plan hints and `Retry-After`.
- Retries with backoff for rate limits and 5xx, optional client-side pacing, `context` everywhere.
- Every result carries your remaining monthly quota.
- A `RemoteCalls` interface and a generated gomock, so code that uses the SDK is easy to test.

> **Scope:** this first cut covers what Silo needs: listing and searching tickers, current prices, company/ticker
> detail and daily price history. Forex, indices, ETFs, bonds, disclosures, dividends, news and WebSockets are not
> wrapped yet. Adding an endpoint is one model type, one method in `api/` and one line on the interface.

## Layout

```
ngnmarket-go/
├── api/                 the client: api.New, the RemoteCalls interface, one file per market
│   ├── api.go           RemoteCalls interface, Call, New
│   ├── options.go       WithBaseURL, WithHTTPClient, WithRateLimit, WithMaxRetries, ...
│   ├── util.go          makeRequest: auth, retries, error mapping
│   ├── ratelimit.go     optional client-side token bucket
│   ├── ngx.go           Nigerian Exchange endpoints
│   ├── us.go            US market endpoints
│   └── mock/            gomock mock of RemoteCalls (generated, do not edit)
├── model/               parameters, responses and errors (no dependencies)
├── examples/            runnable samples, see below
├── quality.sh           fmt, tidy, generate, vet, test, lint
└── .golangci.yml
```

## Install

```bash
go get github.com/wearegravitylabs/ngnmarket-go
```

## Quick start

```go
import (
    "github.com/wearegravitylabs/ngnmarket-go/api"
    "github.com/wearegravitylabs/ngnmarket-go/model"
)

client, err := api.New(os.Getenv("NGNMARKET_API_KEY"), api.WithRateLimit(30, 0))
if err != nil { log.Fatal(err) }

// NGX: one call prices the whole exchange (~150 companies).
companies, meta, err := client.ListAllNGXCompanies(ctx)

// US: exact, case-sensitive symbol lookup.
msft, err := client.FindUSTicker(ctx, "MSFT")
fmt.Println(msft.LastPrice, msft.LogoURL, meta.CallsRemaining)
```

Get a key at <https://ngnmarket.com/developer>.

## Samples

Each is a small program you can run. They need `NGNMARKET_API_KEY` and work on the Free plan unless noted.

| Command | What it shows |
| --- | --- |
| `go run ./examples/ngx_market` | Price the whole NGX in one call, list the largest companies, read the quota |
| `go run ./examples/us_lookup MSFT` | Look up a US ticker by exact symbol, with quote time and delay flag |
| `go run ./examples/portfolio` | Value a mixed NGX and US holding the way a tracker would |
| `go run ./examples/price_on_date NG DANGCEM 2026-03-11` | Price on a past date, with the fallback when the plan has no history (needs Hobby+ to succeed) |
| `go run ./examples/error_handling` | Branch on every kind of failure; client options for retries and pacing |

## Endpoints covered

| Method | API route | Min. plan |
| --- | --- | --- |
| `ListNGXCompanies`, `ListAllNGXCompanies`, `FindNGXCompany` | `GET /companies` | Free |
| `ListNGXSymbols` | `GET /companies/identifiers` | Free |
| `GetNGXCompany` | `GET /companies/{symbol}` | Hobby |
| `GetNGXPriceChart`, `GetNGXClosePriceOnOrBefore` | `GET /companies/{symbol}/chart` | Hobby |
| `ListUSTickers`, `FindUSTicker` | `GET /us/tickers` | Free |
| `ListUSSymbols` | `GET /us/tickers/identifiers` | Free |
| `GetUSTicker` | `GET /us/tickers/{symbol}` | Hobby |
| `GetUSPriceChart`, `GetUSClosePriceOnOrBefore` | `GET /us/tickers/{symbol}/chart` | Hobby |

History depth is 2 years on Hobby, 5 on Starter, full on Pro and above. A `from` older than your plan allows is
clamped by the API, so the `...ClosePriceOnOrBefore` helpers return `model.ErrNoData`.

## Things worth knowing

- **Free plan means current prices only.** Listing returns price, change, volume, market cap and a `logo_url`. Detail
  and history are Hobby+ and return `model.ErrPlanRequired` on a Free key.
- **US quotes are delayed** by at least 15 minutes. Check `Ticker.QuoteTime` for how fresh a price is.
- **US symbols are case-sensitive**, NGX symbols are not. US prices are unadjusted for splits and dividends.
- **US search is a substring match** ordered by market cap, so `FindUSTicker` may scan up to three pages (one call each).
- **Quota is shared.** NGX and US calls draw from one monthly pool (Free: 3,000 calls, 30 per minute). Cache the
  `List...Symbols` results (the US list is about 1 MB) and prefer `ListAllNGXCompanies` to per-symbol lookups.
- **Prices are `float64`**, as the API sends them. Convert to a decimal type before doing money arithmetic.

## Errors

```go
_, err := client.GetNGXCompany(ctx, "DANGCEM")

switch {
case errors.Is(err, model.ErrPlanRequired):  // the key's plan lacks this endpoint
case errors.Is(err, model.ErrRateLimited):   // per-minute cap; see APIError.RetryAfter
case errors.Is(err, model.ErrQuotaExceeded): // monthly quota spent; retrying will not help
case errors.Is(err, model.ErrUnauthorized):  // missing, invalid or IP-restricted key
case errors.Is(err, model.ErrNotFound):
}

var apiErr *model.APIError
if errors.As(err, &apiErr) {
    log.Println(apiErr.StatusCode, apiErr.Code, apiErr.RequiredPlan, apiErr.RetryAfter)
}
```

## Rate limiting and retries

- `RATE_LIMITED` and 5xx responses are retried (default 2 retries, exponential backoff with jitter). `Retry-After`
  is honoured up to `WithMaxRetryWait` (default 5s); a longer wait is returned to you as a `*model.APIError` with
  `RetryAfter` set rather than blocking your request goroutine.
- `QUOTA_EXCEEDED`, auth, plan and other 4xx errors are never retried.
- `WithRateLimit(30, 0)` paces requests client-side so you stay under your plan's per-minute cap.

## Testing code that uses the SDK

Depend on `api.RemoteCalls` and use the generated mock:

```go
ctrl := gomock.NewController(t)
m := mock.NewMockRemoteCalls(ctrl)
m.EXPECT().FindUSTicker(gomock.Any(), "MSFT").Return(&model.Ticker{Symbol: "MSFT", LastPrice: 529.76}, nil)
```

## Development

```bash
make quality      # fmt, tidy, regenerate the mock, vet, race tests, golangci-lint
make test
make examples     # check every sample compiles
make integration  # about 6 live calls; needs NGNMARKET_API_KEY (Hobby-only calls skip on a Free key)
```

Unit tests use `httptest` and fixtures captured from the live API. The NGX `chart` and `detail` fixtures follow the
published OpenAPI spec because those endpoints need a paid plan; run `make integration` with a Hobby key to confirm.
