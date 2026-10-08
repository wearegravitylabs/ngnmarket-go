# Tasks: finish the SDK

Welcome! This SDK is a Go wrapper for the [NGN Market API](https://docs.ngnmarket.com), which gives Nigerian
Exchange (NGX) and US stock data. We built the parts our app (Silo) needs first. **Your job is to wrap the rest of
the endpoints**, one at a time, in the same style.

You do not need to be an expert. Every task is "copy the pattern, change the names". Start with the tasks marked
**S**, and ask when you get stuck.

## 1. Get set up (about 15 minutes)

1. Install Go 1.22 or newer (`go version` to check).
2. Clone the repo and run the tests. They should all pass before you change anything.

```bash
git clone git@github.com:wearegravitylabs/ngnmarket-go.git
cd ngnmarket-go
go test ./...
```

3. Get a **free** API key at <https://ngnmarket.com/developer> and keep it in your terminal session only (never in a file you commit):

```bash
export NGNMARKET_API_KEY=ngm_live_your_key_here
```

4. Run a sample to see it work against the real API:

```bash
go run ./examples/ngx_market
```

5. Install the linter we use (once):

```bash
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
```

## 2. Where things live

```
api/api.go        the list of every call (the RemoteCalls interface). Add your method here.
api/ngx.go        Nigerian Exchange calls.     api/us.go   US market calls.
api/util.go       the one function that talks to the API (makeRequest). You never edit it.
model/ngx.go      NGX types.   model/us.go   US types.   model/model.go   shared types.
model/errors.go   the errors.
api/*_test.go     tests.       examples/   small programs people can run.
api/mock/         generated, never edit by hand. Run `go generate ./...` to refresh it.
```

## 3. The recipe: add one endpoint

Do exactly this for every task. The finished example to copy is **`ListNGXSymbols`**
(`GET /companies/identifiers`). Open these four places and see how it was done:
`model/ngx.go` (look for `CompanyIdentifier`), `api/api.go` (the `ListNGXSymbols` line),
`api/ngx.go` (the `ListNGXSymbols` function) and `api/us_test.go` (`TestListUSSymbols` is the same kind of test).

**Step 1. Read the docs page for the endpoint** (the "Docs path" column below, added to `https://docs.ngnmarket.com`).
Note the URL, the query parameters and the response.

**Step 2. See a real response.** Free endpoints you can call yourself:

```bash
curl -s -H "Authorization: Bearer $NGNMARKET_API_KEY" "https://api.ngnmarket.com/v1/forex/current"
```

For Hobby and higher endpoints your free key gets `PLAN_REQUIRED`, so use the example on the docs page (or
<https://docs.ngnmarket.com/openapi.yaml>) instead, and mark the endpoint **unverified** in the README table.

**Step 3. Add the types in `model/`.** One Go struct per object in the response. The API answers
`{"success":true,"data":{...},"meta":{...}}` and the SDK strips that wrapper for you, so you only model what is
inside `data`. Rules:

- JSON name to Go name: `price_change_percent` becomes `PriceChangePercent` with the tag `` `json:"price_change_percent"` ``.
- If a value can be `null`, use a pointer (`*float64`). If a value might not exist, a plain type is fine (it becomes 0).
- Dates the API sends as `"2026-03-09"` stay as `string`. Full timestamps become `*time.Time`.
- If the endpoint takes query parameters, add a `...Params` struct with a `Values()` method (copy `ListCompaniesParams`).

**Step 4. Add the method to the interface** in `api/api.go`, with a one-line comment saying what it does and the plan it needs.

**Step 5. Write the method** in `api/ngx.go` or `api/us.go` (or a new file like `api/forex.go` for a new area):

```go
// GetForexRates returns the latest NGN exchange rates (Free plan) - https://docs.ngnmarket.com/api-reference/forex/current
func (c *Call) GetForexRates(ctx context.Context) (*model.ForexRates, error) {
	var out model.ForexRates                       // what is inside "data"
	meta, err := c.makeRequest(ctx, "/forex/current", nil, &out)
	if err != nil {
		return nil, err
	}
	out.Meta = meta                                // every result carries the quota info
	return &out, nil
}
```

For a call with a symbol, escape it: `"/companies/"+url.PathEscape(symbol)+"/dividends"`.
For a list with page and limit, copy `ListNGXCompanies`.

**Step 6. Write a test** in the matching `api/*_test.go` file. Copy any existing test: it starts a fake server
with `serve(...)` and returns a small JSON body (use a real response you saved in step 2). Check at least: the URL
and query were right, the numbers came out right, and a `null` does not crash it. **Tests never call the real API.**

**Step 7. Finish up:**

```bash
go generate ./...     # refreshes the mock
./quality.sh          # formats, vets, tests and lints. It must finish with no errors
```

Then add a row to the endpoint table in `README.md` and a line under `[Unreleased]` in `CHANGELOG.md`.
Open **one pull request per endpoint** (or per small group), titled like the issue.

## 4. Rules

- **Never commit an API key.** Not in code, tests, examples or the README. Use `os.Getenv("NGNMARKET_API_KEY")`.
- **Do not edit `api/util.go`, `api/mock/` or `model/errors.go`** unless a task says so.
- **Names:** NGX calls contain `NGX` (`GetNGXCompany`), US calls contain `US` (`GetUSTicker`). Lists start with
  `List`, one item with `Get`, a search for one exact item with `Find`.
- **Symbols:** NGX symbols are case-insensitive, **US symbols are case-sensitive**. Do not change their case.
- **Small pull requests.** One endpoint, its types, its test. Easier to review, faster to merge.
- **Paid endpoints** (Hobby and above) cannot be tried with a free key. Build them from the docs example and say
  "unverified" in the PR. Someone with a paid key will run `make integration` later.

## 5. How to turn this file into GitHub issues

Create **one issue per row** in the tables below. Use the template (New issue, then "Add an endpoint").

1. Create labels: `endpoint`, `good first issue`, `plan: free`, `plan: hobby`, `plan: starter`, `plan: pro`,
   `plan: business`, `size: S`, `size: M`, `blocked`.
2. Create one **milestone** for each section in step 6 (Warm-up, Free lists, and so on).
3. Issue title: `Add GET /forex/current (GetForexRates)`. Fill in the template. Add the plan, size and milestone.
4. Put `good first issue` on every **S** task in the first milestone.
5. When a pull request is merged, tick the box in the table below.

## 6. The tasks

**Size:** **S** = no parameters, flat response (about an hour). **M** = has parameters, paging or nested
data (half a day). **Plan** is the lowest plan whose key can call it. The docs link is `https://docs.ngnmarket.com`
plus the Docs path.

### Milestone A: Warm-up (Free plan, you can test these for real)

Do these first, top to bottom. They teach you the recipe.

| Done | Route | Suggested method | Docs path | Size |
| :-: | --- | --- | --- | :-: |
| [ ] | `GET /forex/current` | `GetForexRates` | `/api-reference/forex/current` | S |
| [ ] | `GET /market/status` | `GetNGXMarketStatus` | `/api-reference/market/status` | S |
| [ ] | `GET /market/snapshot` | `GetNGXMarketSnapshot` | `/api-reference/market/snapshot` | S |
| [ ] | `GET /indices` | `ListNGXIndices` | `/api-reference/indices/list` | S |
| [ ] | `GET /account/usage` | `GetAccountUsage` | `/api-reference/account/usage` | S |
| [ ] | `GET /us/market/status` | `GetUSMarketStatus` | `/us/market/status` | S |
| [ ] | `GET /us/market/snapshot` | `GetUSMarketSnapshot` | `/us/market/snapshot` | S |
| [ ] | `GET /us/indices` | `ListUSIndices` | `/us/indices/list` | S |

### Milestone B: More Free endpoints (lists and paging)

| Done | Route | Suggested method | Docs path | Size |
| :-: | --- | --- | --- | :-: |
| [ ] | `GET /market/holidays` (has `limit`) | `ListNGXMarketHolidays` | `/api-reference/market/holidays` | S |
| [ ] | `GET /market/available-dates` (has `limit`) | `ListNGXMarketDates` | `/api-reference/market/available-dates` | S |
| [ ] | `GET /us/market/holidays` | `ListUSMarketHolidays` | `/us/market/holidays` | S |
| [ ] | `GET /us/market/available-dates` | `ListUSMarketDates` | `/us/market/available-dates` | S |
| [ ] | `GET /etfs` (paged) | `ListNGXETFs` | `/api-reference/etfs/list` | M |
| [ ] | `GET /disclosures` (paged) | `ListDisclosures` | `/api-reference/disclosures/list` | M |
| [ ] | `GET /disclosures/types` | `ListDisclosureTypes` | `/api-reference/disclosures/types` | S |
| [ ] | `GET /disclosures/categories` | `ListDisclosureCategories` | `/api-reference/disclosures/categories` | S |
| [ ] | `GET /companies/{symbol}/disclosures` (paged) | `ListCompanyDisclosures` | `/api-reference/companies/disclosures` | M |
| [ ] | `GET /companies/{symbol}/disclosures/categories` | `ListCompanyDisclosureCategories` | `/api-reference/companies/disclosures-categories` | S |
| [ ] | `GET /blog/posts` (paged) | `ListBlogPosts` | `/api-reference/blog/posts` | M |
| [ ] | `GET /blog/posts/{slug}` | `GetBlogPost` | `/api-reference/blog/post-detail` | M |
| [ ] | `GET /blog/search` | `SearchBlogPosts` | `/api-reference/blog/search` | M |

### Milestone C: Hobby plan (build from the docs, unverified)

| Done | Route | Suggested method | Docs path | Size |
| :-: | --- | --- | --- | :-: |
| [ ] | `GET /indices/{symbol}` | `GetNGXIndex` | `/api-reference/indices/detail` | M |
| [ ] | `GET /indices/{symbol}/chart` | `GetNGXIndexChart` | `/api-reference/indices/chart` | M |
| [ ] | `GET /forex/history` | `GetForexHistory` | `/api-reference/forex/history` | M |
| [ ] | `GET /us/indices/{symbol}` | `GetUSIndex` | `/us/indices/detail` | M |
| [ ] | `GET /us/indices/{symbol}/chart` | `GetUSIndexChart` | `/us/indices/chart` | M |

### Milestone D: Starter plan (unverified)

| Done | Route | Suggested method | Docs path | Size |
| :-: | --- | --- | --- | :-: |
| [ ] | `GET /market/top-trades` | `ListNGXTopTrades` | `/api-reference/market/top-trades` | M |
| [ ] | `GET /market/movers` | `GetNGXMarketMovers` | `/api-reference/market/movers` | M |
| [ ] | `GET /us/market/movers` | `GetUSMarketMovers` | `/us/market/movers` | M |
| [ ] | `GET /us/market/trending` | `GetUSMarketTrending` | `/us/market/trending` | M |
| [ ] | `GET /companies/{symbol}/dividends` | `ListNGXCompanyDividends` | `/api-reference/companies/dividends` | M |
| [ ] | `GET /dividends/upcoming` (paged) | `ListUpcomingDividends` | `/api-reference/dividends/upcoming` | M |
| [ ] | `GET /dividends/recent` (paged) | `ListRecentDividends` | `/api-reference/dividends/recent` | M |
| [ ] | `GET /companies/{symbol}/news` | `ListNGXCompanyNews` | `/api-reference/companies/news` | M |
| [ ] | `GET /bonds` (paged) | `ListBonds` | `/api-reference/bonds/list` | M |
| [ ] | `GET /etfs/{symbol}` | `GetNGXETF` | `/api-reference/etfs/detail` | M |
| [ ] | `GET /etfs/{symbol}/chart` | `GetNGXETFChart` | `/api-reference/etfs/chart` | M |
| [ ] | `GET /us/tickers/{symbol}/officers` | `ListUSTickerOfficers` | `/us/tickers/officers` | S |
| [ ] | `GET /companies/{symbol}/financials/ratios` | `GetNGXFinancialRatios` | `/api-reference/companies/financials-ratios` | M |
| [ ] | `GET /account/logs` (paged) | `ListAccountLogs` | `/api-reference/account/logs` | M |

### Milestone E: Pro plan (unverified)

| Done | Route | Suggested method | Docs path | Size |
| :-: | --- | --- | --- | :-: |
| [ ] | `GET /market/breadth` | `ListNGXMarketBreadth` | `/api-reference/market/breadth` | M |
| [ ] | `GET /market/sectors` | `ListNGXSectors` | `/api-reference/market/sectors` | M |
| [ ] | `GET /market/ytd-performers` | `ListNGXYTDPerformers` | `/api-reference/market/ytd-performers` | M |
| [ ] | `GET /us/market/breadth` | `GetUSMarketBreadth` | `/us/market/breadth` | M |
| [ ] | `GET /us/market/sectors` | `ListUSSectors` | `/us/market/sectors` | M |
| [ ] | `GET /us/market/ytd-performers` | `ListUSYTDPerformers` | `/us/market/ytd-performers` | M |
| [ ] | `GET /companies/{symbol}/financials/income` | `GetNGXIncomeStatement` | `/api-reference/companies/financials-income` | M |
| [ ] | `GET /companies/{symbol}/financials/growth` | `GetNGXFinancialGrowth` | `/api-reference/companies/financials-growth` | M |
| [ ] | `GET /companies/{symbol}/director-dealings` (paged) | `ListNGXDirectorDealings` | `/api-reference/companies/director-dealings` | M |
| [ ] | `GET /companies/{symbol}/director-dealings/summary` | `GetNGXDirectorDealingsSummary` | `/api-reference/companies/director-dealings-summary` | M |

### Milestone F: Business plan (unverified)

| Done | Route | Suggested method | Docs path | Size |
| :-: | --- | --- | --- | :-: |
| [ ] | `GET /companies/{symbol}/financials` | `GetNGXFinancials` | `/api-reference/companies/financials` | M |
| [ ] | `GET /companies/{symbol}/financials/balance-sheet` | `GetNGXBalanceSheet` | `/api-reference/companies/financials-balance-sheet` | M |
| [ ] | `GET /companies/{symbol}/financials/cash-flow` | `GetNGXCashFlow` | `/api-reference/companies/financials-cash-flow` | M |
| [ ] | `GET /companies/{symbol}/financials/ttm` | `GetNGXFinancialsTTM` | `/api-reference/companies/financials-ttm` | M |
| [ ] | `GET /insiders` (paged) | `ListInsiderDealings` | `/api-reference/insiders/list` | M |
| [ ] | `GET /insiders/clusters` | `ListInsiderClusters` | `/api-reference/insiders/clusters` | M |

### Milestone G: SDK improvements (after the endpoints, or in between if you want a change)

These are about the SDK itself, not new endpoints. Talk to the team lead before starting any of them.

| Done | Task | Notes | Size |
| :-: | --- | --- | :-: |
| [ ] | **Paging helper** | One reusable function that walks every page of a paged endpoint, so nobody writes the page loop by hand. See `ListAllNGXCompanies` for the idea. Add it once the first two or three paged endpoints exist. | M |
| [ ] | **NGX chart formats** | `GetNGXPriceChart` always asks for `format=detailed`. Add the compact `chart` and `ohlcv` formats (used for candlestick charts). | M |
| [ ] | **Optional logging** | A `WithLogger` option so apps can see each request (method, path, status, how long). Never log the API key. | M |
| [ ] | **Check the paid shapes** | Someone with a Hobby or higher key runs `make integration`, compares the real responses with our models, and fixes differences (`GetNGXCompany`, `GetNGXPriceChart`, US detail and chart). Mark `blocked` until a paid key exists. | M |
| [ ] | **WebSocket client** | Live prices over WebSocket (<https://docs.ngnmarket.com/websocket/introduction>). Large. Not for a first task. | L |

## 7. Common mistakes (read before your first pull request)

- **The test fails with a JSON error:** a field you typed as a number is sometimes `null` or a string. Use a pointer, or check the real response.
- **A field is always 0:** the JSON tag does not match the name in the response. Copy it exactly, including underscores.
- **`quality.sh` complains about formatting:** run `go fmt ./...`. If `gofmt` complains about generics, you have an old Go on your PATH (`which -a gofmt`).
- **The linter crashes right away:** your `golangci-lint` is too old. Install v2 with the command in section 1.
- **You hit `RATE_LIMITED` or `QUOTA_EXCEEDED` while experimenting:** the free key allows 30 calls a minute and 3,000 a month. Save responses to a file and reuse them instead of calling again.
- **Your PR touches many files:** split it. One endpoint per PR.

## 8. Definition of done (for every issue)

- [ ] Types in `model/`, method on the `RemoteCalls` interface, implementation in `api/`
- [ ] Unit test with a fake server (no real API calls) that passes
- [ ] `go generate ./...` run, so `api/mock` is up to date
- [ ] `./quality.sh` finishes with no errors
- [ ] Row added to the README endpoint table (marked "unverified" if you could not call it for real)
- [ ] Line added to `CHANGELOG.md`
- [ ] No API key anywhere in the diff
