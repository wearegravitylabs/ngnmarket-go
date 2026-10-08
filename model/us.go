package model

import (
	"net/url"
	"time"
)

// Types for the US market endpoints. Prices are in USD; quotes are delayed by at
// least 15 minutes. US symbols are CASE-SENSITIVE: pass them exactly as returned.
type (
	// SecurityType is the kind of US instrument.
	SecurityType string

	// Ticker is one row of the US ticker list, with its last (delayed) quote.
	Ticker struct {
		Symbol            string       `json:"symbol"`
		Name              string       `json:"name"`
		SecurityType      SecurityType `json:"security_type"`
		Sector            string       `json:"sector"`
		Industry          string       `json:"industry"`
		Country           string       `json:"country"`
		SharesOutstanding float64      `json:"shares_outstanding"`
		LogoURL           string       `json:"logo_url"`
		LastPrice         float64      `json:"last_price"`
		PrevClose         float64      `json:"prev_close"`
		ChangeAbs         float64      `json:"change_abs"`
		ChangePct         float64      `json:"change_pct"`
		Volume            float64      `json:"volume"`
		MarketCap         float64      `json:"market_cap"`
		High52Wk          float64      `json:"high_52wk"`
		Low52Wk           float64      `json:"low_52wk"`
		IsDelayed         FlexBool     `json:"is_delayed"`
		// QuoteTime is when LastPrice was captured. Use it, not the response
		// Meta, to judge how fresh a price is.
		QuoteTime *time.Time `json:"quote_time"`
	}

	// TickerList is a page of US tickers.
	TickerList struct {
		Tickers    []Ticker
		Pagination Pagination
		Meta       Meta
	}

	// TickerSort is a sortable field of the US ticker list.
	TickerSort string

	// ListTickersParams filters and pages the US ticker list. The zero value
	// requests the first page with the API defaults (50 per page, by market cap).
	ListTickersParams struct {
		Type   SecurityType
		Sector string
		Search string // matches name or symbol
		Sort   TickerSort
		Order  SortOrder
		Page   int
		Limit  int
	}

	// TickerIdentifier is a lightweight symbol entry.
	TickerIdentifier struct {
		Symbol       string       `json:"symbol"`
		Name         string       `json:"name"`
		SecurityType SecurityType `json:"security_type"`
	}

	// TickerIdentifiers is the full US symbol list (about 12,500 entries).
	TickerIdentifiers struct {
		Tickers []TickerIdentifier
		Count   int
		Meta    Meta
	}

	// TickerDetail is the full profile and quote of one US ticker (Hobby plan).
	TickerDetail struct {
		Symbol            string       `json:"symbol"`
		Name              string       `json:"name"`
		SecurityType      SecurityType `json:"security_type"`
		Sector            string       `json:"sector"`
		Industry          string       `json:"industry"`
		Country           string       `json:"country"`
		CEO               string       `json:"ceo"`
		Website           string       `json:"website"`
		About             string       `json:"about"`
		SharesOutstanding float64      `json:"shares_outstanding"`
		LogoURL           string       `json:"logo_url"`
		IsActive          bool         `json:"is_active"`
		Exchange          string       `json:"exchange"`
		ExchangeName      string       `json:"exchange_name"`
		ExchangeTimezone  string       `json:"exchange_timezone"`
		LastPrice         *float64     `json:"last_price"`
		PrevClose         *float64     `json:"prev_close"`
		ChangeAbs         *float64     `json:"change_abs"`
		ChangePct         *float64     `json:"change_pct"`
		OpenPrice         *float64     `json:"open_price"`
		HighPrice         *float64     `json:"high_price"`
		LowPrice          *float64     `json:"low_price"`
		Volume            *float64     `json:"volume"`
		MarketCap         *float64     `json:"market_cap"`
		EPS               *float64     `json:"eps"`
		PERatio           *float64     `json:"pe_ratio"`
		Beta              *float64     `json:"beta"`
		DividendYield     *float64     `json:"dividend_yield"`
		High52Wk          *float64     `json:"high_52wk"`
		Low52Wk           *float64     `json:"low_52wk"`
		IsDelayed         FlexBool     `json:"is_delayed"`
		QuoteTime         *time.Time   `json:"quote_time"`
		Meta              Meta         `json:"-"`
	}

	// USChartParams selects the history window as UTC calendar days. A From
	// older than the plan's history depth (Hobby 2y, Starter 5y, Pro+ from
	// 2021-08-03) is silently clamped by the API.
	USChartParams struct {
		From time.Time
		To   time.Time
	}

	// USChartPoint is one trading day. Prices are unadjusted for splits/dividends.
	USChartPoint struct {
		Date      string   `json:"date"`
		Open      *float64 `json:"open"`
		High      *float64 `json:"high"`
		Low       *float64 `json:"low"`
		Close     *float64 `json:"close"`
		Volume    *float64 `json:"volume"`
		ChangeAbs *float64 `json:"change_abs"`
		ChangePct *float64 `json:"change_pct"`
	}

	// USChartStats summarises the returned range.
	USChartStats struct {
		StartDate     string   `json:"start_date"`
		EndDate       string   `json:"end_date"`
		StartPrice    float64  `json:"start_price"`
		EndPrice      float64  `json:"end_price"`
		Change        float64  `json:"change"`
		ChangePercent *float64 `json:"change_percent"`
		MinPrice      float64  `json:"min_price"`
		MaxPrice      float64  `json:"max_price"`
	}

	// USChart is a ticker's daily price history.
	USChart struct {
		Symbol          string         `json:"symbol"`
		Count           int            `json:"count"`
		PriceAdjustment string         `json:"price_adjustment"`
		Points          []USChartPoint `json:"data"`
		Statistics      USChartStats   `json:"statistics"`
		Meta            Meta           `json:"-"`
	}
)

// Security types accepted by ListTickersParams.Type.
const (
	SecurityStock     SecurityType = "stock"
	SecurityETF       SecurityType = "etf"
	SecurityADR       SecurityType = "adr"
	SecurityPreferred SecurityType = "preferred"
	SecurityWarrant   SecurityType = "warrant"
	SecurityUnit      SecurityType = "unit"
	SecurityRight     SecurityType = "right"
	SecurityOther     SecurityType = "other"
)

// Sortable fields for ListTickersParams.Sort.
const (
	TickerSortSymbol    TickerSort = "symbol"
	TickerSortName      TickerSort = "name"
	TickerSortSector    TickerSort = "sector"
	TickerSortPrice     TickerSort = "price"
	TickerSortMarketCap TickerSort = "market_cap"
	TickerSortVolume    TickerSort = "volume"
	TickerSortChangePct TickerSort = "change_pct"
)

// Values encodes the parameters as a query string. Zero values are omitted.
func (p *ListTickersParams) Values() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	setStr(q, "type", string(p.Type))
	setStr(q, "sector", p.Sector)
	setStr(q, "search", p.Search)
	setStr(q, "sort", string(p.Sort))
	setStr(q, "order", string(p.Order))
	setInt(q, "page", p.Page)
	setInt(q, "limit", p.Limit)
	return q
}

// Values encodes the parameters as a query string. Zero values are omitted.
func (p *USChartParams) Values() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	setDate(q, "from", p.From)
	setDate(q, "to", p.To)
	return q
}
