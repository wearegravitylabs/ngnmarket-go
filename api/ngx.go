package api

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/wearegravitylabs/ngnmarket-go/model"
)

// ListNGXCompanies returns one page of NGX companies with current prices (Free plan) - https://docs.ngnmarket.com/api-reference/companies/list
func (c *Call) ListNGXCompanies(ctx context.Context, params *model.ListCompaniesParams) (*model.CompanyList, error) {
	var wire struct {
		Data       []model.Company  `json:"data"`
		Pagination model.Pagination `json:"pagination"`
	}
	meta, err := c.makeRequest(ctx, "/companies", params.Values(), &wire)
	if err != nil {
		return nil, err
	}
	return &model.CompanyList{Companies: wire.Data, Pagination: wire.Pagination, Meta: meta}, nil
}

// ListAllNGXCompanies returns every NGX company with its current price, following pagination (Free plan).
// The exchange lists roughly 150 companies, so this is normally a single call.
func (c *Call) ListAllNGXCompanies(ctx context.Context) ([]model.Company, model.Meta, error) {
	var (
		all  []model.Company
		meta model.Meta
	)
	for page := 1; ; page++ {
		res, err := c.ListNGXCompanies(ctx, &model.ListCompaniesParams{Page: page, Limit: 200})
		if err != nil {
			return nil, model.Meta{}, err
		}
		all = append(all, res.Companies...)
		meta = res.Meta
		if !res.Pagination.HasNext || len(res.Companies) == 0 {
			return all, meta, nil
		}
	}
}

// FindNGXCompany looks up one company by symbol (case-insensitive) using the list endpoint, so it works on the Free plan.
func (c *Call) FindNGXCompany(ctx context.Context, symbol string) (*model.Company, error) {
	symbol, err := requireSymbol(symbol)
	if err != nil {
		return nil, err
	}
	res, err := c.ListNGXCompanies(ctx, &model.ListCompaniesParams{Search: symbol, Limit: 50})
	if err != nil {
		return nil, err
	}
	for i := range res.Companies {
		if strings.EqualFold(res.Companies[i].Symbol, symbol) {
			return &res.Companies[i], nil
		}
	}
	return nil, &model.APIError{StatusCode: 404, Code: model.CodeNotFound, Message: "no NGX company with symbol " + symbol}
}

// ListNGXSymbols returns every NGX ticker symbol, unpaginated (Free plan) - https://docs.ngnmarket.com/api-reference/companies/identifiers
func (c *Call) ListNGXSymbols(ctx context.Context) (*model.CompanyIdentifiers, error) {
	var wire struct {
		Data  []model.CompanyIdentifier `json:"data"`
		Count int                       `json:"count"`
	}
	meta, err := c.makeRequest(ctx, "/companies/identifiers", nil, &wire)
	if err != nil {
		return nil, err
	}
	return &model.CompanyIdentifiers{Companies: wire.Data, Count: wire.Count, Meta: meta}, nil
}

// GetNGXCompany returns the full profile of one company (Hobby plan or above) - https://docs.ngnmarket.com/api-reference/companies/detail
func (c *Call) GetNGXCompany(ctx context.Context, symbol string) (*model.CompanyDetail, error) {
	symbol, err := requireSymbol(symbol)
	if err != nil {
		return nil, err
	}
	var d model.CompanyDetail
	meta, err := c.makeRequest(ctx, "/companies/"+url.PathEscape(symbol), nil, &d)
	if err != nil {
		return nil, err
	}
	d.Meta = meta
	return &d, nil
}

// GetNGXPriceChart returns daily price history (Hobby plan or above) - https://docs.ngnmarket.com/api-reference/companies/chart
func (c *Call) GetNGXPriceChart(ctx context.Context, symbol string, params *model.NGXChartParams) (*model.NGXChart, error) {
	symbol, err := requireSymbol(symbol)
	if err != nil {
		return nil, err
	}
	var ch model.NGXChart
	meta, err := c.makeRequest(ctx, "/companies/"+url.PathEscape(symbol)+"/chart", params.Values(), &ch)
	if err != nil {
		return nil, err
	}
	ch.Meta = meta
	return &ch, nil
}

// GetNGXClosePriceOnOrBefore returns the closing price on day, or on the nearest earlier trading day when day
// was a weekend or holiday (Hobby plan or above). It returns model.ErrNoData when the API has nothing at or
// before day, which happens when day is older than the plan's history depth.
func (c *Call) GetNGXClosePriceOnOrBefore(ctx context.Context, symbol string, day time.Time) (model.PricePoint, error) {
	want := day.UTC().Format(model.DateLayout)
	ch, err := c.GetNGXPriceChart(ctx, symbol, &model.NGXChartParams{From: day.AddDate(0, 0, -10), To: day})
	if err != nil {
		return model.PricePoint{}, err
	}
	var best model.PricePoint
	found := false
	for _, pt := range ch.Points {
		d := pt.Day()
		closePrice, ok := pt.ClosePrice()
		if !ok || d == "" || d > want {
			continue
		}
		if !found || d > best.Date {
			best, found = model.PricePoint{Date: d, Close: closePrice}, true
		}
	}
	if !found {
		return model.PricePoint{}, model.ErrNoData
	}
	return best, nil
}
