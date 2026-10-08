package api

import (
	"context"
	"net/url"
	"time"

	"github.com/wearegravitylabs/ngnmarket-go/model"
)

// findMaxPages bounds how many list pages FindUSTicker will scan.
const findMaxPages = 3

// ListUSTickers returns one page of US tickers with their last (delayed) price (Free plan) - https://docs.ngnmarket.com/us/tickers/list
func (c *Call) ListUSTickers(ctx context.Context, params *model.ListTickersParams) (*model.TickerList, error) {
	var wire struct {
		Data       []model.Ticker   `json:"data"`
		Pagination model.Pagination `json:"pagination"`
	}
	meta, err := c.makeRequest(ctx, "/us/tickers", params.Values(), &wire)
	if err != nil {
		return nil, err
	}
	return &model.TickerList{Tickers: wire.Data, Pagination: wire.Pagination, Meta: meta}, nil
}

// FindUSTicker looks up one US ticker by its exact, case-sensitive symbol using the list endpoint, so it works
// on the Free plan. The API's search is a substring match on name and symbol and results are ordered by market
// cap, so the symbol may not be on the first page: it scans up to three pages of 100, one call per page.
func (c *Call) FindUSTicker(ctx context.Context, symbol string) (*model.Ticker, error) {
	symbol, err := requireSymbol(symbol)
	if err != nil {
		return nil, err
	}
	for page := 1; page <= findMaxPages; page++ {
		res, err := c.ListUSTickers(ctx, &model.ListTickersParams{Search: symbol, Page: page, Limit: 100})
		if err != nil {
			return nil, err
		}
		for i := range res.Tickers {
			if res.Tickers[i].Symbol == symbol {
				return &res.Tickers[i], nil
			}
		}
		if !res.Pagination.HasNext {
			break
		}
	}
	return nil, &model.APIError{StatusCode: 404, Code: model.CodeNotFound, Message: "no US ticker with symbol " + symbol}
}

// ListUSSymbols returns every active US ticker symbol, unpaginated (Free plan) - https://docs.ngnmarket.com/us/tickers/identifiers
func (c *Call) ListUSSymbols(ctx context.Context) (*model.TickerIdentifiers, error) {
	var wire struct {
		Data  []model.TickerIdentifier `json:"data"`
		Count int                      `json:"count"`
	}
	meta, err := c.makeRequest(ctx, "/us/tickers/identifiers", nil, &wire)
	if err != nil {
		return nil, err
	}
	return &model.TickerIdentifiers{Tickers: wire.Data, Count: wire.Count, Meta: meta}, nil
}

// GetUSTicker returns the full profile of one US ticker (Hobby plan or above) - https://docs.ngnmarket.com/us/tickers/detail
func (c *Call) GetUSTicker(ctx context.Context, symbol string) (*model.TickerDetail, error) {
	symbol, err := requireSymbol(symbol)
	if err != nil {
		return nil, err
	}
	var d model.TickerDetail
	meta, err := c.makeRequest(ctx, "/us/tickers/"+url.PathEscape(symbol), nil, &d)
	if err != nil {
		return nil, err
	}
	d.Meta = meta
	return &d, nil
}

// GetUSPriceChart returns daily price history (Hobby plan or above) - https://docs.ngnmarket.com/us/tickers/chart
func (c *Call) GetUSPriceChart(ctx context.Context, symbol string, params *model.USChartParams) (*model.USChart, error) {
	symbol, err := requireSymbol(symbol)
	if err != nil {
		return nil, err
	}
	var ch model.USChart
	meta, err := c.makeRequest(ctx, "/us/tickers/"+url.PathEscape(symbol)+"/chart", params.Values(), &ch)
	if err != nil {
		return nil, err
	}
	ch.Meta = meta
	return &ch, nil
}

// GetUSClosePriceOnOrBefore returns the closing price on day, or on the nearest earlier trading day when day
// was a weekend or holiday (Hobby plan or above). Prices are unadjusted for splits and dividends. It returns
// model.ErrNoData when the API has nothing at or before day.
func (c *Call) GetUSClosePriceOnOrBefore(ctx context.Context, symbol string, day time.Time) (model.PricePoint, error) {
	want := day.UTC().Format(model.DateLayout)
	ch, err := c.GetUSPriceChart(ctx, symbol, &model.USChartParams{From: day.AddDate(0, 0, -10), To: day})
	if err != nil {
		return model.PricePoint{}, err
	}
	var best model.PricePoint
	found := false
	for _, pt := range ch.Points {
		if pt.Close == nil || pt.Date == "" || pt.Date > want {
			continue
		}
		if !found || pt.Date > best.Date {
			best, found = model.PricePoint{Date: pt.Date, Close: *pt.Close}, true
		}
	}
	if !found {
		return model.PricePoint{}, model.ErrNoData
	}
	return best, nil
}
