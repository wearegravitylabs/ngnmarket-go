// Package api implements the NGN Market API calls (https://docs.ngnmarket.com).
//
// Create a client once with New and share it: it is safe for concurrent use.
// Code that depends on the SDK should accept the RemoteCalls interface so it can
// be mocked (see the mock package).
package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/wearegravitylabs/ngnmarket-go/model"
)

const (
	// Version is the SDK version, sent in the User-Agent header.
	Version = "0.1.0"

	// DefaultBaseURL is the production API. NGX endpoints live directly under
	// it ("/companies"); US endpoints share the same host and key under "/us".
	DefaultBaseURL = "https://api.ngnmarket.com/v1"
)

// RemoteCalls is the abstracted definition of the supported functions.
//
// Plan access (https://docs.ngnmarket.com/plans): the list, find and symbol
// methods work on the Free plan and include the current price. The detail and
// price-history methods need Hobby or above and fail with an error matching
// model.ErrPlanRequired on a Free key.
//
//go:generate mockgen -source api.go -destination ./mock/mock_remote_calls.go -package mock RemoteCalls
type RemoteCalls interface {
	// ─── Nigerian Exchange (NGX), prices in NGN ──────────────────────────────

	// ListNGXCompanies returns one page of NGX companies with current prices (Free).
	ListNGXCompanies(ctx context.Context, params *model.ListCompaniesParams) (*model.CompanyList, error)
	// ListAllNGXCompanies returns every NGX company with its price, following
	// pagination. The exchange lists ~150 companies, so this is normally a single
	// call: the cheapest way to price a whole portfolio (Free).
	ListAllNGXCompanies(ctx context.Context) ([]model.Company, model.Meta, error)
	// FindNGXCompany looks up one company by symbol (case-insensitive) (Free).
	// Returns an error matching model.ErrNotFound if there is none.
	FindNGXCompany(ctx context.Context, symbol string) (*model.Company, error)
	// ListNGXSymbols returns every NGX symbol, unpaginated (Free). Cache it.
	ListNGXSymbols(ctx context.Context) (*model.CompanyIdentifiers, error)
	// GetNGXCompany returns the full profile of one company (Hobby+).
	GetNGXCompany(ctx context.Context, symbol string) (*model.CompanyDetail, error)
	// GetNGXPriceChart returns daily price history (Hobby+).
	GetNGXPriceChart(ctx context.Context, symbol string, params *model.NGXChartParams) (*model.NGXChart, error)
	// GetNGXClosePriceOnOrBefore returns the close on day, or on the nearest
	// earlier trading day (Hobby+). model.ErrNoData if there is none.
	GetNGXClosePriceOnOrBefore(ctx context.Context, symbol string, day time.Time) (model.PricePoint, error)

	// ─── US market, prices in USD, quotes delayed, symbols case-sensitive ────

	// ListUSTickers returns one page of US tickers with last prices (Free).
	ListUSTickers(ctx context.Context, params *model.ListTickersParams) (*model.TickerList, error)
	// FindUSTicker looks up one ticker by exact, case-sensitive symbol (Free).
	// The API's search is a substring match ordered by market cap, so this may
	// scan up to three pages (one call each). Returns model.ErrNotFound if absent.
	FindUSTicker(ctx context.Context, symbol string) (*model.Ticker, error)
	// ListUSSymbols returns every US symbol (~12,500, ~1 MB) unpaginated (Free). Cache it.
	ListUSSymbols(ctx context.Context) (*model.TickerIdentifiers, error)
	// GetUSTicker returns the full profile of one ticker (Hobby+).
	GetUSTicker(ctx context.Context, symbol string) (*model.TickerDetail, error)
	// GetUSPriceChart returns daily price history, unadjusted (Hobby+).
	GetUSPriceChart(ctx context.Context, symbol string, params *model.USChartParams) (*model.USChart, error)
	// GetUSClosePriceOnOrBefore returns the close on day, or on the nearest
	// earlier trading day (Hobby+). model.ErrNoData if there is none.
	GetUSClosePriceOnOrBefore(ctx context.Context, symbol string, day time.Time) (model.PricePoint, error)
}

// Call is the NGN Market client. It implements RemoteCalls.
type Call struct {
	apiKey       string
	baseURL      string
	httpClient   *http.Client
	userAgent    string
	maxRetries   int
	maxRetryWait time.Duration
	limiter      *limiter

	sleep func(ctx context.Context, d time.Duration) error // injectable for tests
}

var _ RemoteCalls = (*Call)(nil)

// New returns a client for the given API key. Keys look like "ngm_live_...";
// create one at https://ngnmarket.com/developer.
func New(apiKey string, opts ...Option) (RemoteCalls, error) {
	return newCall(apiKey, opts...)
}

func newCall(apiKey string, opts ...Option) (*Call, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, errors.New("ngnmarket: API key is required")
	}
	c := &Call{
		apiKey:       apiKey,
		baseURL:      DefaultBaseURL,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		userAgent:    "ngnmarket-go/" + Version,
		maxRetries:   2,
		maxRetryWait: 5 * time.Second,
		sleep:        sleepCtx,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}
	return c, nil
}
